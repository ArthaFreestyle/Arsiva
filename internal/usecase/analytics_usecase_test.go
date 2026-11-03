package usecase

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/repository"
)

// ─── Mock analytics repo ──────────────────────────────────────────────────────

type mockAnalyticsRepo struct {
	userCounts    entity.UserRoleCount
	contentCounts entity.ContentCount
	err           error

	userBaseline    entity.UserRoleCount
	userBuckets     []entity.UserGrowthBucket
	contentBaseline entity.ContentTypeCount
	contentBuckets  []entity.ContentGrowthBucket

	// captured arguments of the last growth call
	gotInterval repository.AnalyticsInterval
	gotFrom     time.Time
	gotTo       time.Time
}

func (m *mockAnalyticsRepo) CountUsersByRole(_ context.Context) (entity.UserRoleCount, error) {
	return m.userCounts, m.err
}
func (m *mockAnalyticsRepo) CountContent(_ context.Context) (entity.ContentCount, error) {
	return m.contentCounts, m.err
}
func (m *mockAnalyticsRepo) UserGrowth(_ context.Context, interval repository.AnalyticsInterval, from, to time.Time) (entity.UserRoleCount, []entity.UserGrowthBucket, error) {
	m.gotInterval, m.gotFrom, m.gotTo = interval, from, to
	return m.userBaseline, m.userBuckets, m.err
}
func (m *mockAnalyticsRepo) ContentGrowth(_ context.Context, interval repository.AnalyticsInterval, from, to time.Time) (entity.ContentTypeCount, []entity.ContentGrowthBucket, error) {
	m.gotInterval, m.gotFrom, m.gotTo = interval, from, to
	return m.contentBaseline, m.contentBuckets, m.err
}

// 2026-11-04 01:30 WIB == 2026-11-03 18:30 UTC — "today" must be the WIB date.
var analyticsTestNow = time.Date(2026, 11, 3, 18, 30, 0, 0, time.UTC)

func newTestAnalyticsUseCase(repo *mockAnalyticsRepo) *analyticsUseCaseImpl {
	log := logrus.New()
	log.SetOutput(io.Discard)
	uc := NewAnalyticsUseCase(repo, log).(*analyticsUseCaseImpl)
	uc.Now = func() time.Time { return analyticsTestNow }
	return uc
}

func day(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

func assertStatus(t *testing.T, err error, want int) {
	t.Helper()
	var fe *fiber.Error
	if !errors.As(err, &fe) {
		t.Fatalf("expected *fiber.Error with status %d, got %v", want, err)
	}
	if fe.Code != want {
		t.Fatalf("expected status %d, got %d (%s)", want, fe.Code, fe.Message)
	}
}

// ─── Summary ──────────────────────────────────────────────────────────────────

func TestAnalytics_UserSummary_TotalIsSumOfRoles(t *testing.T) {
	repo := &mockAnalyticsRepo{userCounts: entity.UserRoleCount{SuperAdmin: 0, Guru: 2, Member: 8}}
	resp, err := newTestAnalyticsUseCase(repo).GetUserSummary(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 10 {
		t.Errorf("total = %d, want 10", resp.Total)
	}
	if resp.ByRole.SuperAdmin != 0 || resp.ByRole.Guru != 2 || resp.ByRole.Member != 8 {
		t.Errorf("by_role = %+v", resp.ByRole)
	}
}

func TestAnalytics_UserSummary_RepoErrorIs500(t *testing.T) {
	repo := &mockAnalyticsRepo{err: errors.New("db down")}
	_, err := newTestAnalyticsUseCase(repo).GetUserSummary(context.Background())
	assertStatus(t, err, fiber.StatusInternalServerError)
}

func TestAnalytics_ContentSummary_DerivedCounts(t *testing.T) {
	repo := &mockAnalyticsRepo{contentCounts: entity.ContentCount{
		KuisTotal: 9, KuisPublished: 5,
		CeritaTotal: 9, CeritaPublished: 4,
		PuzzleTotal: 4, PuzzlePublished: 2,
		ArtikelTotal: 5, ArtikelPublished: 3,
	}}
	resp, err := newTestAnalyticsUseCase(repo).GetContentSummary(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	g := resp.Game
	if g.Total != 22 || g.Published != 11 || g.Draft != 11 {
		t.Errorf("game = %d/%d/%d, want 22/11/11 (artikel must not be counted)", g.Total, g.Published, g.Draft)
	}
	for name, c := range map[string]model.PublishCountResponse{
		"kuis": g.Kuis, "cerita": g.Cerita, "puzzle": g.Puzzle, "artikel": resp.Artikel,
	} {
		if c.Published+c.Draft != c.Total {
			t.Errorf("%s: published+draft (%d+%d) != total %d", name, c.Published, c.Draft, c.Total)
		}
	}
	if g.Kuis.Draft != 4 || g.Cerita.Draft != 5 || g.Puzzle.Draft != 2 || resp.Artikel.Draft != 2 {
		t.Errorf("unexpected drafts: %+v / %+v", g, resp.Artikel)
	}
}

// ─── Growth range resolution ──────────────────────────────────────────────────

func TestAnalytics_Growth_Defaults(t *testing.T) {
	repo := &mockAnalyticsRepo{}
	resp, err := newTestAnalyticsUseCase(repo).GetUserGrowth(context.Background(), &model.GrowthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Interval != "daily" || len(resp.Series) != 30 {
		t.Fatalf("interval=%s buckets=%d, want daily/30", resp.Interval, len(resp.Series))
	}
	// today is the WIB date (2026-11-04), not the UTC date (2026-11-03).
	if resp.To != "2026-11-04" || resp.From != "2026-10-06" {
		t.Errorf("range = %s..%s, want 2026-10-06..2026-11-04", resp.From, resp.To)
	}
	if resp.Series[29].PeriodStart != "2026-11-04" {
		t.Errorf("last period_start = %s", resp.Series[29].PeriodStart)
	}
	if !repo.gotFrom.Equal(day("2026-10-06")) || !repo.gotTo.Equal(day("2026-11-04")) {
		t.Errorf("repo got %v..%v", repo.gotFrom, repo.gotTo)
	}
}

func TestAnalytics_Growth_DefaultWeeklyAndMonthly(t *testing.T) {
	uc := newTestAnalyticsUseCase(&mockAnalyticsRepo{})

	weekly, err := uc.GetUserGrowth(context.Background(), &model.GrowthRequest{Interval: "weekly"})
	if err != nil {
		t.Fatal(err)
	}
	// 2026-11-04 is a Wednesday → current week starts Monday 2026-11-02.
	if len(weekly.Series) != 12 || weekly.To != "2026-11-02" || weekly.From != "2026-08-17" {
		t.Errorf("weekly = %s..%s (%d)", weekly.From, weekly.To, len(weekly.Series))
	}

	monthly, err := uc.GetUserGrowth(context.Background(), &model.GrowthRequest{Interval: "monthly"})
	if err != nil {
		t.Fatal(err)
	}
	if len(monthly.Series) != 12 || monthly.To != "2026-11-01" || monthly.From != "2025-12-01" {
		t.Errorf("monthly = %s..%s (%d)", monthly.From, monthly.To, len(monthly.Series))
	}
}

func TestAnalytics_Growth_WeeklySnapsToMonday(t *testing.T) {
	uc := newTestAnalyticsUseCase(&mockAnalyticsRepo{})
	// 2026-10-14 is a Wednesday, 2026-10-25 a Sunday.
	resp, err := uc.GetUserGrowth(context.Background(), &model.GrowthRequest{Interval: "weekly", From: "2026-10-14", To: "2026-10-25"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.From != "2026-10-12" || resp.To != "2026-10-19" || len(resp.Series) != 2 {
		t.Errorf("got %s..%s (%d), want 2026-10-12..2026-10-19 (2)", resp.From, resp.To, len(resp.Series))
	}
}

func TestAnalytics_Growth_MonthlySnapsToFirst(t *testing.T) {
	uc := newTestAnalyticsUseCase(&mockAnalyticsRepo{})
	resp, err := uc.GetUserGrowth(context.Background(), &model.GrowthRequest{Interval: "monthly", From: "2026-01-15", To: "2026-11-03"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.From != "2026-01-01" || resp.To != "2026-11-01" || len(resp.Series) != 11 {
		t.Errorf("got %s..%s (%d)", resp.From, resp.To, len(resp.Series))
	}
}

func TestAnalytics_Growth_FutureToIsClampedToToday(t *testing.T) {
	uc := newTestAnalyticsUseCase(&mockAnalyticsRepo{})
	resp, err := uc.GetUserGrowth(context.Background(), &model.GrowthRequest{From: "2026-11-01", To: "2027-01-01"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.To != "2026-11-04" || len(resp.Series) != 4 {
		t.Errorf("got to=%s (%d buckets), want 2026-11-04 (4)", resp.To, len(resp.Series))
	}
}

func TestAnalytics_Growth_BadRequests(t *testing.T) {
	cases := map[string]*model.GrowthRequest{
		"unknown interval": {Interval: "yearly"},
		"malformed from":   {From: "03-11-2026"},
		"malformed to":     {To: "2026/11/03"},
		"to before from":   {From: "2026-10-10", To: "2026-10-01"},
		"future from":      {From: "2026-12-01"},
		"daily too long":   {Interval: "daily", From: "2026-01-01", To: "2026-07-20"},
		"weekly too long":  {Interval: "weekly", From: "2020-01-01"},
		"monthly too long": {Interval: "monthly", From: "2020-01-01"},
	}
	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &mockAnalyticsRepo{}
			_, err := newTestAnalyticsUseCase(repo).GetUserGrowth(context.Background(), req)
			assertStatus(t, err, fiber.StatusBadRequest)
		})
	}
}

func TestAnalytics_Growth_MaxBucketsAllowed(t *testing.T) {
	uc := newTestAnalyticsUseCase(&mockAnalyticsRepo{})
	// 2026-08-05 .. 2026-11-04 is exactly 92 days.
	resp, err := uc.GetUserGrowth(context.Background(), &model.GrowthRequest{From: "2026-08-05"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Series) != 92 {
		t.Errorf("buckets = %d, want 92", len(resp.Series))
	}
}

// ─── Growth series ────────────────────────────────────────────────────────────

func TestAnalytics_UserGrowth_ZeroFillAndCumulative(t *testing.T) {
	repo := &mockAnalyticsRepo{
		userBaseline: entity.UserRoleCount{Guru: 1, Member: 4}, // 5 users before range
		userBuckets: []entity.UserGrowthBucket{ // sparse: 2026-11-02 has no rows
			{Bucket: day("2026-11-01"), UserRoleCount: entity.UserRoleCount{Member: 1}},
			{Bucket: day("2026-11-03"), UserRoleCount: entity.UserRoleCount{Guru: 1, Member: 1}},
		},
	}
	resp, err := newTestAnalyticsUseCase(repo).GetUserGrowth(context.Background(),
		&model.GrowthRequest{From: "2026-11-01", To: "2026-11-03"})
	if err != nil {
		t.Fatal(err)
	}

	wantStarts := []string{"2026-11-01", "2026-11-02", "2026-11-03"}
	wantNew := []int{1, 0, 2}
	wantCum := []int{6, 6, 8}
	if len(resp.Series) != 3 {
		t.Fatalf("buckets = %d, want 3", len(resp.Series))
	}
	for i, p := range resp.Series {
		if p.PeriodStart != wantStarts[i] || p.NewUsers != wantNew[i] || p.Cumulative != wantCum[i] {
			t.Errorf("series[%d] = %+v, want start=%s new=%d cum=%d", i, p, wantStarts[i], wantNew[i], wantCum[i])
		}
		if p.NewUsers != p.NewByRole.SuperAdmin+p.NewByRole.Guru+p.NewByRole.Member {
			t.Errorf("series[%d]: new_users != sum(new_by_role)", i)
		}
	}
}

func TestAnalytics_ContentGrowth_ZeroFillAndCumulative(t *testing.T) {
	repo := &mockAnalyticsRepo{
		contentBaseline: entity.ContentTypeCount{Kuis: 3, Cerita: 1, Puzzle: 0, Artikel: 2},
		contentBuckets: []entity.ContentGrowthBucket{
			{Bucket: day("2026-10-01"), ContentTypeCount: entity.ContentTypeCount{Kuis: 1, Artikel: 1}},
		},
	}
	resp, err := newTestAnalyticsUseCase(repo).GetContentGrowth(context.Background(),
		&model.GrowthRequest{Interval: "monthly", From: "2026-09-10"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Series) != 3 || resp.From != "2026-09-01" || resp.To != "2026-11-01" {
		t.Fatalf("got %s..%s (%d)", resp.From, resp.To, len(resp.Series))
	}

	sep, oct, nov := resp.Series[0], resp.Series[1], resp.Series[2]
	if sep.New.Total != 0 || sep.Cumulative.Total != 6 {
		t.Errorf("sep = %+v", sep)
	}
	if oct.New.Kuis != 1 || oct.New.Artikel != 1 || oct.New.Total != 2 {
		t.Errorf("oct.new = %+v", oct.New)
	}
	want := model.ContentTypeCountResponse{Kuis: 4, Cerita: 1, Puzzle: 0, Artikel: 3, Total: 8}
	if oct.Cumulative != want || nov.Cumulative != want {
		t.Errorf("cumulative oct=%+v nov=%+v, want %+v", oct.Cumulative, nov.Cumulative, want)
	}
	if repo.gotInterval != repository.AnalyticsIntervalMonthly {
		t.Errorf("repo interval = %s", repo.gotInterval)
	}
}

func TestAnalytics_Growth_RepoErrorIs500(t *testing.T) {
	repo := &mockAnalyticsRepo{err: errors.New("db down")}
	_, err := newTestAnalyticsUseCase(repo).GetContentGrowth(context.Background(), &model.GrowthRequest{})
	assertStatus(t, err, fiber.StatusInternalServerError)
}
