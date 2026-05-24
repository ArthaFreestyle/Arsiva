package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
)

type GamificationRepository interface {
	// Streak — reads the current row, or returns a zeroed struct if none yet.
	GetStreak(ctx context.Context, memberId int) (*entity.MemberStreak, error)

	// Atomically locks the streak row (creating it if missing), calls fn to apply
	// business rules in-place, then saves the result. One tx + SELECT FOR UPDATE.
	UpdateStreakAtomic(ctx context.Context, memberId int, fn func(*entity.MemberStreak)) (*entity.MemberStreak, error)

	// Daily tasks — idempotent setup of today's fixed task set via ON CONFLICT DO NOTHING.
	EnsureTodayTasks(ctx context.Context, memberId int, today time.Time) error

	// Returns today's task rows (all types including the hidden bonus row).
	GetTodayTasks(ctx context.Context, memberId int, today time.Time) ([]*entity.DailyTask, error)

	// Atomically increments daily task progress after content is finished.
	// Handles XP credit and the all-tasks bonus, all in one transaction.
	ProgressDailyTasks(ctx context.Context, memberId int, today time.Time, contentType string, awardedXp int) error
}

type gamificationRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewGamificationRepository(db *pgxpool.Pool, log *logrus.Logger) GamificationRepository {
	return &gamificationRepositoryImpl{DB: db, Log: log}
}

// ─── Streak ──────────────────────────────────────────────────────────────────

func (r *gamificationRepositoryImpl) GetStreak(ctx context.Context, memberId int) (*entity.MemberStreak, error) {
	streak := &entity.MemberStreak{}
	err := r.DB.QueryRow(ctx, `
		SELECT member_id, current_streak, longest_streak, last_active_date, freezes_available, updated_at
		FROM member_streaks
		WHERE member_id = $1
	`, memberId).Scan(
		&streak.MemberId,
		&streak.CurrentStreak,
		&streak.LongestStreak,
		&streak.LastActiveDate,
		&streak.FreezesAvailable,
		&streak.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return &entity.MemberStreak{MemberId: memberId}, nil
	}
	if err != nil {
		r.Log.Errorf("GetStreak member_id=%d: %v", memberId, err)
		return nil, err
	}
	return streak, nil
}

func (r *gamificationRepositoryImpl) UpdateStreakAtomic(ctx context.Context, memberId int, fn func(*entity.MemberStreak)) (*entity.MemberStreak, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		r.Log.Errorf("UpdateStreakAtomic begin tx: %v", err)
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Lazily create the streak row if it doesn't exist yet.
	_, err = tx.Exec(ctx, `
		INSERT INTO member_streaks (member_id)
		VALUES ($1)
		ON CONFLICT (member_id) DO NOTHING
	`, memberId)
	if err != nil {
		r.Log.Errorf("UpdateStreakAtomic upsert: %v", err)
		return nil, err
	}

	// Lock the row so concurrent finalize calls don't double-count the same day.
	var streak entity.MemberStreak
	err = tx.QueryRow(ctx, `
		SELECT member_id, current_streak, longest_streak, last_active_date, freezes_available, updated_at
		FROM member_streaks
		WHERE member_id = $1
		FOR UPDATE
	`, memberId).Scan(
		&streak.MemberId,
		&streak.CurrentStreak,
		&streak.LongestStreak,
		&streak.LastActiveDate,
		&streak.FreezesAvailable,
		&streak.UpdatedAt,
	)
	if err != nil {
		r.Log.Errorf("UpdateStreakAtomic select for update: %v", err)
		return nil, err
	}

	// Apply streak rules in the usecase (fn modifies the struct in-place).
	fn(&streak)

	_, err = tx.Exec(ctx, `
		UPDATE member_streaks
		SET current_streak    = $1,
		    longest_streak    = $2,
		    last_active_date  = $3,
		    freezes_available = $4,
		    updated_at        = NOW()
		WHERE member_id = $5
	`, streak.CurrentStreak, streak.LongestStreak, streak.LastActiveDate, streak.FreezesAvailable, memberId)
	if err != nil {
		r.Log.Errorf("UpdateStreakAtomic update: %v", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		r.Log.Errorf("UpdateStreakAtomic commit: %v", err)
		return nil, err
	}
	return &streak, nil
}

// ─── Daily tasks ─────────────────────────────────────────────────────────────

// fixedDailyTasks is the fixed set of tasks generated each day.
// daily_complete_bonus tracks whether all visible tasks are done (hidden from list response).
var fixedDailyTasks = []struct {
	taskType    string
	targetCount int
	xpReward    int
}{
	{"complete_quiz", 1, 20},
	{"complete_story", 1, 20},
	{"solve_puzzle", 1, 20},
	{"earn_xp", 100, 30},
	{"daily_complete_bonus", 4, 50},
}

func (r *gamificationRepositoryImpl) EnsureTodayTasks(ctx context.Context, memberId int, today time.Time) error {
	for _, t := range fixedDailyTasks {
		_, err := r.DB.Exec(ctx, `
			INSERT INTO daily_tasks (member_id, task_date, task_type, target_count, xp_reward)
			VALUES ($1, $2, $3::daily_task_type_enum, $4, $5)
			ON CONFLICT (member_id, task_date, task_type) DO NOTHING
		`, memberId, today, t.taskType, t.targetCount, t.xpReward)
		if err != nil {
			r.Log.Errorf("EnsureTodayTasks insert %s: %v", t.taskType, err)
			return err
		}
	}
	return nil
}

func (r *gamificationRepositoryImpl) GetTodayTasks(ctx context.Context, memberId int, today time.Time) ([]*entity.DailyTask, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT daily_task_id, member_id, task_date, task_type, target_count, current_count, xp_reward, completed_at, created_at
		FROM daily_tasks
		WHERE member_id = $1 AND task_date = $2
		ORDER BY daily_task_id
	`, memberId, today)
	if err != nil {
		r.Log.Errorf("GetTodayTasks member_id=%d: %v", memberId, err)
		return nil, err
	}
	defer rows.Close()

	var tasks []*entity.DailyTask
	for rows.Next() {
		t := &entity.DailyTask{}
		if err := rows.Scan(
			&t.DailyTaskId, &t.MemberId, &t.TaskDate, &t.TaskType,
			&t.TargetCount, &t.CurrentCount, &t.XpReward, &t.CompletedAt, &t.CreatedAt,
		); err != nil {
			r.Log.Errorf("GetTodayTasks scan: %v", err)
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return tasks, nil
}

func (r *gamificationRepositoryImpl) ProgressDailyTasks(ctx context.Context, memberId int, today time.Time, contentType string, awardedXp int) error {
	// Ensure task rows exist before starting the transaction.
	if err := r.EnsureTodayTasks(ctx, memberId, today); err != nil {
		return err
	}

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		r.Log.Errorf("ProgressDailyTasks begin tx: %v", err)
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Increment the task matching this content type (kuis→complete_quiz, etc.).
	taskType := contentTypeToTaskType(contentType)
	if taskType != "" {
		var completed bool
		var xpReward int
		completed, xpReward, err = r.incrementTask(ctx, tx, memberId, today, taskType, 1)
		if err != nil {
			return err
		}
		if completed {
			if err = r.creditXP(ctx, tx, memberId, xpReward); err != nil {
				return err
			}
			if err = r.advanceBonusTask(ctx, tx, memberId, today); err != nil {
				return err
			}
		}
	}

	// Increment earn_xp by how much XP was actually awarded this finish.
	if awardedXp > 0 {
		var completed bool
		var xpReward int
		completed, xpReward, err = r.incrementTask(ctx, tx, memberId, today, "earn_xp", awardedXp)
		if err != nil {
			return err
		}
		if completed {
			if err = r.creditXP(ctx, tx, memberId, xpReward); err != nil {
				return err
			}
			if err = r.advanceBonusTask(ctx, tx, memberId, today); err != nil {
				return err
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		r.Log.Errorf("ProgressDailyTasks commit: %v", err)
		return err
	}
	return nil
}

// incrementTask increments current_count for the given task type (capped at target_count).
// If the task just crossed target_count for the first time, it sets completed_at and
// returns (true, xpReward). If it was already complete or doesn't exist, returns (false, 0).
func (r *gamificationRepositoryImpl) incrementTask(ctx context.Context, tx pgx.Tx, memberId int, today time.Time, taskType string, increment int) (bool, int, error) {
	var currentCount, targetCount, xpReward int
	err := tx.QueryRow(ctx, `
		UPDATE daily_tasks
		SET current_count = LEAST(current_count + $1, target_count)
		WHERE member_id = $2 AND task_date = $3 AND task_type = $4::daily_task_type_enum AND completed_at IS NULL
		RETURNING current_count, target_count, xp_reward
	`, increment, memberId, today, taskType).Scan(&currentCount, &targetCount, &xpReward)
	if errors.Is(err, pgx.ErrNoRows) {
		// Task already completed or doesn't exist — either way, nothing to do.
		return false, 0, nil
	}
	if err != nil {
		r.Log.Errorf("incrementTask %s: %v", taskType, err)
		return false, 0, err
	}

	if currentCount < targetCount {
		return false, 0, nil
	}

	// Task just reached its target — stamp completed_at.
	_, err = tx.Exec(ctx, `
		UPDATE daily_tasks
		SET completed_at = NOW()
		WHERE member_id = $1 AND task_date = $2 AND task_type = $3::daily_task_type_enum AND completed_at IS NULL
	`, memberId, today, taskType)
	if err != nil {
		r.Log.Errorf("incrementTask set completed_at %s: %v", taskType, err)
		return false, 0, err
	}
	return true, xpReward, nil
}

// advanceBonusTask increments the hidden daily_complete_bonus task by 1.
// When it reaches its target (4), it awards 50 XP — the all-tasks completion bonus.
func (r *gamificationRepositoryImpl) advanceBonusTask(ctx context.Context, tx pgx.Tx, memberId int, today time.Time) error {
	completed, xpReward, err := r.incrementTask(ctx, tx, memberId, today, "daily_complete_bonus", 1)
	if err != nil {
		return err
	}
	if completed {
		return r.creditXP(ctx, tx, memberId, xpReward)
	}
	return nil
}

func (r *gamificationRepositoryImpl) creditXP(ctx context.Context, tx pgx.Tx, memberId, xpReward int) error {
	_, err := tx.Exec(ctx,
		`UPDATE members SET total_xp = total_xp + $1 WHERE member_id = $2`,
		xpReward, memberId)
	if err != nil {
		r.Log.Errorf("creditXP member_id=%d xp=%d: %v", memberId, xpReward, err)
	}
	return err
}

// contentTypeToTaskType maps content_type_enum values to daily_task_type_enum values.
func contentTypeToTaskType(contentType string) string {
	switch contentType {
	case "kuis":
		return "complete_quiz"
	case "cerita":
		return "complete_story"
	case "puzzle":
		return "solve_puzzle"
	default:
		return ""
	}
}
