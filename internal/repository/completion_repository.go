package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/utils"
)

type CompletionRepository interface {
	FindCompletion(context.Context, int, string, string, int) (*model.ProgressFinalizeResponse, error)
	Complete(context.Context, string, *entity.MemberProgress, int, string, time.Time, func(*entity.MemberStreak)) (*model.ProgressFinalizeResponse, error)
}
type completionRepository struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewCompletionRepository(db *pgxpool.Pool, log *logrus.Logger) CompletionRepository {
	return &completionRepository{db, log}
}
func (r *completionRepository) FindCompletion(ctx context.Context, member int, id, kind string, content int) (*model.ProgressFinalizeResponse, error) {
	var result model.ProgressFinalizeResponse
	err := r.DB.QueryRow(ctx, `SELECT result FROM progress_completions WHERE completion_id=$1 AND member_id=$2 AND content_type=$3 AND content_id=$4`, id, member, kind, content).Scan(&result)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &result, err
}
func (r *completionRepository) Complete(ctx context.Context, id string, p *entity.MemberProgress, xp int, ending string, today time.Time, streakFn func(*entity.MemberStreak)) (*model.ProgressFinalizeResponse, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var initial int
	if err = tx.QueryRow(ctx, `SELECT total_xp FROM members WHERE member_id=$1 FOR UPDATE`, p.MemberId).Scan(&initial); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidReference
		}
		return nil, err
	}
	var saved model.ProgressFinalizeResponse
	err = tx.QueryRow(ctx, `SELECT result FROM progress_completions WHERE completion_id=$1 AND member_id=$2 AND content_type=$3 AND content_id=$4`, id, p.MemberId, p.ContentType, p.ContentId).Scan(&saved)
	if err == nil {
		return &saved, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	var group *string
	if p.GroupId != "" {
		// Lock the reference against deletion until the progress row is inserted.
		var existing string
		err = tx.QueryRow(ctx, `SELECT group_id::text FROM groups WHERE group_id::text=$1 FOR KEY SHARE`, p.GroupId).Scan(&existing)
		if err == nil {
			group = &existing
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}
	pr := &memberProgressRepositoryImpl{DB: r.DB, Log: r.Log}
	pid, _, _, effective, err := pr.saveProgressTx(ctx, tx, p, xp, group)
	if err != nil {
		return nil, err
	}
	gr := &gamificationRepositoryImpl{DB: r.DB, Log: r.Log}
	if _, err = gr.updateStreakTx(ctx, tx, p.MemberId, streakFn); err != nil {
		return nil, err
	}
	if err = gr.progressDailyTasksTx(ctx, tx, p.MemberId, today, p.ContentType, effective); err != nil {
		return nil, err
	}
	var total int
	if err = tx.QueryRow(ctx, `SELECT total_xp FROM members WHERE member_id=$1`, p.MemberId).Scan(&total); err != nil {
		return nil, err
	}
	level := utils.LevelForXP(total)
	if _, err = tx.Exec(ctx, `UPDATE members SET level=$1 WHERE member_id=$2`, level, p.MemberId); err != nil {
		return nil, err
	}
	unlocked, notifications, err := AwardEligibleTx(ctx, tx, p.MemberId, true)
	if err != nil {
		return nil, err
	}
	result := &model.ProgressFinalizeResponse{
		CompletionID: id, ProgresId: pid, Skor: p.Skor,
		AwardedXP: effective, BonusXP: total - initial - effective,
		TotalAwardedXP: total - initial, TotalXP: total,
		PreviousLevel: utils.LevelForXP(initial), NewLevel: level,
		LeveledUp: level > utils.LevelForXP(initial), EndingType: ending,
		NewlyUnlockedAchievements: unlocked, Notifications: notifications,
	}

	if _, err = tx.Exec(ctx, `INSERT INTO progress_completions(completion_id,member_id,content_type,content_id,result) VALUES($1,$2,$3,$4,$5)`, id, p.MemberId, p.ContentType, p.ContentId, result); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

// AwardEligibleTx requires the caller to lock the member row first.
func AwardEligibleTx(ctx context.Context, tx pgx.Tx, member int, notify bool) ([]model.UnlockedAchievement, []model.Notification, error) {
	unlocked := []model.UnlockedAchievement{}
	notifications := []model.Notification{}
	rows, err := tx.Query(ctx, `INSERT INTO member_achievements(member_id,achievement_id) SELECT m.member_id,a.achievement_id FROM members m JOIN achievements a ON a.xp_required<=m.total_xp WHERE m.member_id=$1 ORDER BY a.achievement_id ON CONFLICT(member_id,achievement_id) DO NOTHING RETURNING achievement_id::text,unlocked_at`, member)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var a model.UnlockedAchievement
		if err = rows.Scan(&a.AchievementID, &a.UnlockedAt); err != nil {
			rows.Close()
			return nil, nil, err
		}
		unlocked = append(unlocked, a)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, nil, err
	}
	for i := range unlocked {
		a := &unlocked[i]
		if err = tx.QueryRow(ctx, `SELECT nama,COALESCE(badge_icon,''),tier::text FROM achievements WHERE achievement_id=$1`, a.AchievementID).Scan(&a.Nama, &a.BadgeIcon, &a.Tier); err != nil {
			return nil, nil, err
		}
		if !notify {
			continue
		}
		n := model.Notification{Type: "achievement_unlocked", Title: "Achievement baru!", Message: "Kamu mendapatkan " + a.Nama + ".", Payload: a.AchievementPayload}
		payload, _ := json.Marshal(n.Payload)
		err = tx.QueryRow(ctx, `INSERT INTO member_notifications(member_id,type,dedupe_key,title,message,payload) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(member_id,dedupe_key) DO NOTHING RETURNING notification_id::text,created_at`, member, n.Type, "achievement_unlocked:"+a.AchievementID, n.Title, n.Message, payload).Scan(&n.NotificationID, &n.CreatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			if err = tx.QueryRow(ctx, `SELECT notification_id::text FROM member_notifications WHERE member_id=$1 AND dedupe_key=$2`, member, "achievement_unlocked:"+a.AchievementID).Scan(&a.NotificationID); err != nil {
				return nil, nil, err
			}
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		a.NotificationID = n.NotificationID
		notifications = append(notifications, n)
	}
	return unlocked, notifications, nil
}

// ReconcileMember uses the same serialization and evaluator as completion.
func ReconcileMember(ctx context.Context, db *pgxpool.Pool, member int, notify bool) (int, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	var id int
	if err = tx.QueryRow(ctx, `SELECT member_id FROM members WHERE member_id=$1 FOR UPDATE`, member).Scan(&id); err != nil {
		return 0, err
	}
	a, _, err := AwardEligibleTx(ctx, tx, member, notify)
	if err != nil {
		return 0, err
	}
	return len(a), tx.Commit(ctx)
}
