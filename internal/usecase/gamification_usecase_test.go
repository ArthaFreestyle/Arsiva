package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

// ─── Mock GamificationRepository ─────────────────────────────────────────────

type mockGamificationRepo struct {
	getStreakFn          func(ctx context.Context, memberId int) (*entity.MemberStreak, error)
	updateStreakAtomicFn func(ctx context.Context, memberId int, fn func(*entity.MemberStreak)) (*entity.MemberStreak, error)
	ensureTodayTasksFn   func(ctx context.Context, memberId int, today time.Time) error
	getTodayTasksFn      func(ctx context.Context, memberId int, today time.Time) ([]*entity.DailyTask, error)
	progressDailyTasksFn func(ctx context.Context, memberId int, today time.Time, contentType string, awardedXp int) error
}

func (m *mockGamificationRepo) GetStreak(ctx context.Context, memberId int) (*entity.MemberStreak, error) {
	if m.getStreakFn != nil {
		return m.getStreakFn(ctx, memberId)
	}
	return &entity.MemberStreak{MemberId: memberId}, nil
}

func (m *mockGamificationRepo) UpdateStreakAtomic(ctx context.Context, memberId int, fn func(*entity.MemberStreak)) (*entity.MemberStreak, error) {
	if m.updateStreakAtomicFn != nil {
		return m.updateStreakAtomicFn(ctx, memberId, fn)
	}
	streak := &entity.MemberStreak{MemberId: memberId}
	fn(streak)
	return streak, nil
}

func (m *mockGamificationRepo) EnsureTodayTasks(ctx context.Context, memberId int, today time.Time) error {
	if m.ensureTodayTasksFn != nil {
		return m.ensureTodayTasksFn(ctx, memberId, today)
	}
	return nil
}

func (m *mockGamificationRepo) GetTodayTasks(ctx context.Context, memberId int, today time.Time) ([]*entity.DailyTask, error) {
	if m.getTodayTasksFn != nil {
		return m.getTodayTasksFn(ctx, memberId, today)
	}
	return nil, nil
}

func (m *mockGamificationRepo) ProgressDailyTasks(ctx context.Context, memberId int, today time.Time, contentType string, awardedXp int) error {
	if m.progressDailyTasksFn != nil {
		return m.progressDailyTasksFn(ctx, memberId, today, contentType, awardedXp)
	}
	return nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func jakartaDay(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, streakLocation)
}

func ptr(t time.Time) *time.Time { return &t }

func gamificationMemberClaims(memberId string) *model.Claims {
	return &model.Claims{
		Role: "member",
		Details: map[string]any{
			"member_id": memberId,
		},
	}
}

// ─── applyStreakRules tests ───────────────────────────────────────────────────

func TestApplyStreakRules(t *testing.T) {
	today := jakartaDay(2026, 5, 24)
	yesterday := jakartaDay(2026, 5, 23)
	twoDaysAgo := jakartaDay(2026, 5, 22)
	threeDaysAgo := jakartaDay(2026, 5, 21)

	tests := []struct {
		name  string
		input entity.MemberStreak
		want  entity.MemberStreak
	}{
		{
			name: "same day no-op",
			input: entity.MemberStreak{
				CurrentStreak: 5, LongestStreak: 5,
				LastActiveDate: ptr(today),
			},
			want: entity.MemberStreak{
				CurrentStreak: 5, LongestStreak: 5,
				LastActiveDate: ptr(today),
			},
		},
		{
			name: "first ever active day",
			input: entity.MemberStreak{CurrentStreak: 0, LongestStreak: 0},
			want: entity.MemberStreak{
				CurrentStreak: 1, LongestStreak: 1,
				LastActiveDate: ptr(today),
			},
		},
		{
			name: "consecutive day increments streak",
			input: entity.MemberStreak{
				CurrentStreak: 3, LongestStreak: 3,
				LastActiveDate: ptr(yesterday),
			},
			want: entity.MemberStreak{
				CurrentStreak: 4, LongestStreak: 4,
				LastActiveDate: ptr(today),
			},
		},
		{
			name: "one-day gap with freeze preserves streak",
			input: entity.MemberStreak{
				CurrentStreak: 6, LongestStreak: 6,
				LastActiveDate:   ptr(twoDaysAgo),
				FreezesAvailable: 1,
			},
			want: entity.MemberStreak{
				CurrentStreak: 6, LongestStreak: 6,
				LastActiveDate:   ptr(today),
				FreezesAvailable: 0,
			},
		},
		{
			name: "one-day gap without freeze resets streak",
			input: entity.MemberStreak{
				CurrentStreak: 6, LongestStreak: 6,
				LastActiveDate:   ptr(twoDaysAgo),
				FreezesAvailable: 0,
			},
			want: entity.MemberStreak{
				CurrentStreak: 1, LongestStreak: 6,
				LastActiveDate:   ptr(today),
				FreezesAvailable: 0,
			},
		},
		{
			name: "multi-day gap resets streak even with freeze",
			input: entity.MemberStreak{
				CurrentStreak: 10, LongestStreak: 10,
				LastActiveDate:   ptr(threeDaysAgo),
				FreezesAvailable: 2,
			},
			want: entity.MemberStreak{
				CurrentStreak: 1, LongestStreak: 10,
				LastActiveDate:   ptr(today),
				FreezesAvailable: 2,
			},
		},
		{
			name: "streak at multiple of 7 grants a freeze",
			input: entity.MemberStreak{
				CurrentStreak: 6, LongestStreak: 6,
				LastActiveDate:   ptr(yesterday),
				FreezesAvailable: 0,
			},
			want: entity.MemberStreak{
				CurrentStreak: 7, LongestStreak: 7,
				LastActiveDate:   ptr(today),
				FreezesAvailable: 1,
			},
		},
		{
			name: "freeze capped at 2",
			input: entity.MemberStreak{
				CurrentStreak: 6, LongestStreak: 6,
				LastActiveDate:   ptr(yesterday),
				FreezesAvailable: 2,
			},
			want: entity.MemberStreak{
				CurrentStreak: 7, LongestStreak: 7,
				LastActiveDate:   ptr(today),
				FreezesAvailable: 2,
			},
		},
		{
			name: "longest_streak updated when new record",
			input: entity.MemberStreak{
				CurrentStreak: 14, LongestStreak: 10,
				LastActiveDate:   ptr(yesterday),
				FreezesAvailable: 0,
			},
			want: entity.MemberStreak{
				CurrentStreak: 15, LongestStreak: 15,
				LastActiveDate:   ptr(today),
				FreezesAvailable: 0,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input
			applyStreakRules(&got, today)

			if got.CurrentStreak != tc.want.CurrentStreak {
				t.Errorf("CurrentStreak: got %d, want %d", got.CurrentStreak, tc.want.CurrentStreak)
			}
			if got.LongestStreak != tc.want.LongestStreak {
				t.Errorf("LongestStreak: got %d, want %d", got.LongestStreak, tc.want.LongestStreak)
			}
			if got.FreezesAvailable != tc.want.FreezesAvailable {
				t.Errorf("FreezesAvailable: got %d, want %d", got.FreezesAvailable, tc.want.FreezesAvailable)
			}
			if tc.want.LastActiveDate == nil && got.LastActiveDate != nil {
				t.Error("LastActiveDate: expected nil, got non-nil")
			}
			if tc.want.LastActiveDate != nil {
				if got.LastActiveDate == nil {
					t.Error("LastActiveDate: expected non-nil, got nil")
				} else if !sameDay(*got.LastActiveDate, *tc.want.LastActiveDate) {
					t.Errorf("LastActiveDate: got %v, want %v", got.LastActiveDate, tc.want.LastActiveDate)
				}
			}
		})
	}
}

// ─── Constructor test ─────────────────────────────────────────────────────────

func TestNewGamificationUseCase(t *testing.T) {
	uc := NewGamificationUseCase(nil, nil)
	if uc == nil {
		t.Fatal("expected non-nil usecase")
	}
}

// ─── GetStreak tests ──────────────────────────────────────────────────────────

func TestGetStreak_ReturnsActiveToday(t *testing.T) {
	today := todayInJakarta()
	repo := &mockGamificationRepo{
		getStreakFn: func(_ context.Context, memberId int) (*entity.MemberStreak, error) {
			return &entity.MemberStreak{
				MemberId: memberId, CurrentStreak: 3, LongestStreak: 5,
				LastActiveDate: ptr(today), FreezesAvailable: 1,
			}, nil
		},
	}
	uc := NewGamificationUseCase(repo, nil)

	resp, err := uc.GetStreak(context.Background(), gamificationMemberClaims("7"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.ActiveToday {
		t.Error("expected ActiveToday=true")
	}
	if resp.CurrentStreak != 3 {
		t.Errorf("CurrentStreak: got %d, want 3", resp.CurrentStreak)
	}
}

func TestGetStreak_NeverActive(t *testing.T) {
	repo := &mockGamificationRepo{
		getStreakFn: func(_ context.Context, memberId int) (*entity.MemberStreak, error) {
			return &entity.MemberStreak{MemberId: memberId}, nil
		},
	}
	uc := NewGamificationUseCase(repo, nil)

	resp, err := uc.GetStreak(context.Background(), gamificationMemberClaims("7"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ActiveToday {
		t.Error("expected ActiveToday=false for member with no streak")
	}
	if resp.LastActiveDate != "" {
		t.Errorf("expected empty LastActiveDate, got %q", resp.LastActiveDate)
	}
}

func TestGetStreak_ForbiddenWithoutMemberId(t *testing.T) {
	uc := NewGamificationUseCase(&mockGamificationRepo{}, nil)

	// Claims with no member_id details (guru role).
	_, err := uc.GetStreak(context.Background(), &model.Claims{Role: "guru"})
	if err == nil {
		t.Fatal("expected forbidden error, got nil")
	}
}

// ─── GetTodayTasks tests ──────────────────────────────────────────────────────

func TestGetTodayTasks_FiltersBonusRow(t *testing.T) {
	today := todayInJakarta()
	repo := &mockGamificationRepo{
		getTodayTasksFn: func(_ context.Context, _ int, _ time.Time) ([]*entity.DailyTask, error) {
			return []*entity.DailyTask{
				{TaskType: "complete_quiz", TargetCount: 1, CurrentCount: 0, XpReward: 20},
				{TaskType: "complete_story", TargetCount: 1, CurrentCount: 1, XpReward: 20, CompletedAt: ptr(today)},
				{TaskType: "solve_puzzle", TargetCount: 1, CurrentCount: 0, XpReward: 20},
				{TaskType: "earn_xp", TargetCount: 100, CurrentCount: 50, XpReward: 30},
				{TaskType: "daily_complete_bonus", TargetCount: 4, CurrentCount: 1, XpReward: 50},
			}, nil
		},
	}
	uc := NewGamificationUseCase(repo, nil)

	resp, err := uc.GetTodayTasks(context.Background(), gamificationMemberClaims("7"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The hidden bonus row must not appear in the response.
	for _, item := range resp.Tasks {
		if item.TaskType == "daily_complete_bonus" {
			t.Error("daily_complete_bonus must be filtered from the list response")
		}
	}
	if len(resp.Tasks) != 4 {
		t.Errorf("expected 4 visible tasks, got %d", len(resp.Tasks))
	}
	if resp.AllCompleted {
		t.Error("expected AllCompleted=false when some tasks are pending")
	}
}

func TestGetTodayTasks_AllCompleted(t *testing.T) {
	now := time.Now()
	repo := &mockGamificationRepo{
		getTodayTasksFn: func(_ context.Context, _ int, _ time.Time) ([]*entity.DailyTask, error) {
			return []*entity.DailyTask{
				{TaskType: "complete_quiz", TargetCount: 1, CurrentCount: 1, XpReward: 20, CompletedAt: &now},
				{TaskType: "complete_story", TargetCount: 1, CurrentCount: 1, XpReward: 20, CompletedAt: &now},
				{TaskType: "solve_puzzle", TargetCount: 1, CurrentCount: 1, XpReward: 20, CompletedAt: &now},
				{TaskType: "earn_xp", TargetCount: 100, CurrentCount: 100, XpReward: 30, CompletedAt: &now},
				{TaskType: "daily_complete_bonus", TargetCount: 4, CurrentCount: 4, XpReward: 50, CompletedAt: &now},
			}, nil
		},
	}
	uc := NewGamificationUseCase(repo, nil)

	resp, err := uc.GetTodayTasks(context.Background(), gamificationMemberClaims("7"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.AllCompleted {
		t.Error("expected AllCompleted=true when all visible tasks have completed_at set")
	}
}

// ─── HandleContentFinished tests ─────────────────────────────────────────────

func TestHandleContentFinished_DelegatesCorrectly(t *testing.T) {
	streakCalled := false
	tasksCalled := false
	capturedContentType := ""
	capturedAwardedXp := 0

	repo := &mockGamificationRepo{
		updateStreakAtomicFn: func(_ context.Context, _ int, fn func(*entity.MemberStreak)) (*entity.MemberStreak, error) {
			streakCalled = true
			s := &entity.MemberStreak{}
			fn(s)
			return s, nil
		},
		progressDailyTasksFn: func(_ context.Context, _ int, _ time.Time, contentType string, awardedXp int) error {
			tasksCalled = true
			capturedContentType = contentType
			capturedAwardedXp = awardedXp
			return nil
		},
	}
	uc := NewGamificationUseCase(repo, nil)

	err := uc.HandleContentFinished(context.Background(), 42, "kuis", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !streakCalled {
		t.Error("expected UpdateStreakAtomic to be called")
	}
	if !tasksCalled {
		t.Error("expected ProgressDailyTasks to be called")
	}
	if capturedContentType != "kuis" {
		t.Errorf("contentType: got %q, want %q", capturedContentType, "kuis")
	}
	if capturedAwardedXp != 50 {
		t.Errorf("awardedXp: got %d, want 50", capturedAwardedXp)
	}
}

func TestHandleContentFinished_StreakErrorPropagates(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &mockGamificationRepo{
		updateStreakAtomicFn: func(_ context.Context, _ int, _ func(*entity.MemberStreak)) (*entity.MemberStreak, error) {
			return nil, repoErr
		},
	}
	uc := NewGamificationUseCase(repo, nil)

	err := uc.HandleContentFinished(context.Background(), 42, "kuis", 0)
	if err == nil {
		t.Fatal("expected error from failed streak update, got nil")
	}
}

// ─── buildDailyTasksResponse tests ───────────────────────────────────────────

func TestBuildDailyTasksResponse_EmptyTasksNotAllCompleted(t *testing.T) {
	today := todayInJakarta()
	resp := buildDailyTasksResponse(nil, today)
	if resp.AllCompleted {
		t.Error("empty task list should not be AllCompleted")
	}
	if resp.Tasks != nil {
		t.Error("expected nil tasks slice for empty input")
	}
}
