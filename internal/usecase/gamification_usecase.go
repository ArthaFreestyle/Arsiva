package usecase

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/repository"
)

// streakLocation is the single fixed timezone used for all day-boundary calculations.
var streakLocation, _ = time.LoadLocation("Asia/Jakarta")

type GamificationUseCase interface {
	GetStreak(ctx context.Context, claims *model.Claims) (*model.StreakResponse, error)
	GetTodayTasks(ctx context.Context, claims *model.Claims) (*model.DailyTasksResponse, error)

	// HandleContentFinished updates streak and daily-task progress after a member
	// finishes any content. Errors are expected to be logged and swallowed by the caller
	// so they never roll back the already-committed progress record.
	HandleContentFinished(ctx context.Context, memberId int, contentType string, awardedXp int) error
}

type gamificationUseCaseImpl struct {
	Repo repository.GamificationRepository
	Log  *logrus.Logger
}

func NewGamificationUseCase(repo repository.GamificationRepository, log *logrus.Logger) GamificationUseCase {
	return &gamificationUseCaseImpl{Repo: repo, Log: log}
}

// ─── GetStreak ───────────────────────────────────────────────────────────────

func (u *gamificationUseCaseImpl) GetStreak(ctx context.Context, claims *model.Claims) (*model.StreakResponse, error) {
	memberId, err := memberIdFromClaims(claims)
	if err != nil {
		return nil, fiber.ErrForbidden
	}

	streak, err := u.Repo.GetStreak(ctx, memberId)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	today := todayInJakarta()
	activeToday := streak.LastActiveDate != nil && sameDay(*streak.LastActiveDate, today)

	lastActiveDateStr := ""
	if streak.LastActiveDate != nil {
		lastActiveDateStr = streak.LastActiveDate.In(streakLocation).Format("2006-01-02")
	}

	return &model.StreakResponse{
		CurrentStreak:    streak.CurrentStreak,
		LongestStreak:    streak.LongestStreak,
		FreezesAvailable: streak.FreezesAvailable,
		LastActiveDate:   lastActiveDateStr,
		ActiveToday:      activeToday,
	}, nil
}

// ─── GetTodayTasks ───────────────────────────────────────────────────────────

func (u *gamificationUseCaseImpl) GetTodayTasks(ctx context.Context, claims *model.Claims) (*model.DailyTasksResponse, error) {
	memberId, err := memberIdFromClaims(claims)
	if err != nil {
		return nil, fiber.ErrForbidden
	}

	today := todayInJakarta()

	if err := u.Repo.EnsureTodayTasks(ctx, memberId, today); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	tasks, err := u.Repo.GetTodayTasks(ctx, memberId, today)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return buildDailyTasksResponse(tasks, today), nil
}

// ─── HandleContentFinished ───────────────────────────────────────────────────

func (u *gamificationUseCaseImpl) HandleContentFinished(ctx context.Context, memberId int, contentType string, awardedXp int) error {
	today := todayInJakarta()

	if _, err := u.Repo.UpdateStreakAtomic(ctx, memberId, func(streak *entity.MemberStreak) {
		applyStreakRules(streak, today)
	}); err != nil {
		return err
	}

	if err := u.Repo.ProgressDailyTasks(ctx, memberId, today, contentType, awardedXp); err != nil {
		return err
	}

	return nil
}

// ─── Streak rules (pure — no DB calls) ───────────────────────────────────────

// applyStreakRules applies the streak advancement rules (1–6) in-place.
// Called inside the FOR UPDATE transaction so concurrent finishes are serialized.
func applyStreakRules(streak *entity.MemberStreak, today time.Time) {
	// Rule 1: already counted today — nothing to change.
	if streak.LastActiveDate != nil && sameDay(*streak.LastActiveDate, today) {
		return
	}

	if streak.LastActiveDate == nil {
		// Rule 4: very first active day.
		streak.CurrentStreak = 1
	} else {
		yesterday := today.AddDate(0, 0, -1)
		twoDaysAgo := today.AddDate(0, 0, -2)

		switch {
		case sameDay(*streak.LastActiveDate, yesterday):
			// Rule 2: consecutive day — extend streak.
			streak.CurrentStreak++
		case sameDay(*streak.LastActiveDate, twoDaysAgo) && streak.FreezesAvailable > 0:
			// Rule 3a: exactly one missed day + freeze available — spend freeze, keep streak.
			streak.FreezesAvailable--
		default:
			// Rule 3b: gap too large or no freeze — reset streak.
			streak.CurrentStreak = 1
		}
	}

	last := today
	streak.LastActiveDate = &last

	// Rule 5: track all-time longest streak.
	if streak.CurrentStreak > streak.LongestStreak {
		streak.LongestStreak = streak.CurrentStreak
	}

	// Rule 6: award one freeze at every multiple of 7, capped at 2.
	if streak.CurrentStreak%7 == 0 && streak.FreezesAvailable < 2 {
		streak.FreezesAvailable++
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func todayInJakarta() time.Time {
	now := time.Now().In(streakLocation)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, streakLocation)
}

func sameDay(a, b time.Time) bool {
	a = a.In(streakLocation)
	b = b.In(streakLocation)
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func memberIdFromClaims(claims *model.Claims) (int, error) {
	memberIdStr := extractMemberIdFromClaims(claims)
	memberId, err := strconv.Atoi(memberIdStr)
	if err != nil || memberId == 0 {
		return 0, fiber.ErrForbidden
	}
	return memberId, nil
}

// buildDailyTasksResponse converts entity rows to the API response,
// filtering out the hidden daily_complete_bonus row.
func buildDailyTasksResponse(tasks []*entity.DailyTask, today time.Time) *model.DailyTasksResponse {
	var items []model.DailyTaskItem
	allCompleted := true
	visibleCount := 0

	for _, t := range tasks {
		if t.TaskType == "daily_complete_bonus" {
			continue
		}
		visibleCount++
		items = append(items, model.DailyTaskItem{
			TaskType:     t.TaskType,
			TargetCount:  t.TargetCount,
			CurrentCount: t.CurrentCount,
			XpReward:     t.XpReward,
			Completed:    t.CompletedAt != nil,
		})
		if t.CompletedAt == nil {
			allCompleted = false
		}
	}

	if visibleCount == 0 {
		allCompleted = false
	}

	return &model.DailyTasksResponse{
		TaskDate:     today.In(streakLocation).Format("2006-01-02"),
		Tasks:        items,
		AllCompleted: allCompleted,
	}
}
