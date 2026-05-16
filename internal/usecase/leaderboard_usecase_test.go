package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/repository"
)

// ─── Mock leaderboard repo ────────────────────────────────────────────────────

type mockLeaderboardRepo struct {
	publicEntries  []entity.LeaderboardEntry
	publicTotal    int
	publicPeriod   time.Time
	publicErr      error
	groupEntries   []entity.LeaderboardEntry
	groupTotal     int
	groupErr       error
	headerName     string
	headerThumb    *string
	headerFound    bool
	headerErr      error
}

func (m *mockLeaderboardRepo) FetchPublic(_ context.Context, _ repository.LeaderboardPeriod, _ int, _, _ int) ([]entity.LeaderboardEntry, int, time.Time, error) {
	return m.publicEntries, m.publicTotal, m.publicPeriod, m.publicErr
}
func (m *mockLeaderboardRepo) FetchGroup(_ context.Context, _ string, _, _ int) ([]entity.LeaderboardEntry, int, error) {
	return m.groupEntries, m.groupTotal, m.groupErr
}
func (m *mockLeaderboardRepo) GetGroupHeader(_ context.Context, _ string) (string, *string, bool, error) {
	return m.headerName, m.headerThumb, m.headerFound, m.headerErr
}

// ─── Mock group repo ──────────────────────────────────────────────────────────

type mockGroupRepoForLeaderboard struct {
	guruId       int
	guruIdErr    error
	memberId     int
	memberIdErr  error
	isMember     bool
	isMemberErr  error
	groupCreator int // groups.created_by for GetGroupById
}

func (m *mockGroupRepoForLeaderboard) GetGuruIdByUserId(_ context.Context, _ string) (int, error) {
	return m.guruId, m.guruIdErr
}
func (m *mockGroupRepoForLeaderboard) GetMemberIdByUserId(_ context.Context, _ string) (int, error) {
	return m.memberId, m.memberIdErr
}
func (m *mockGroupRepoForLeaderboard) IsMemberInGroup(_ context.Context, _ string, _ int) (bool, error) {
	return m.isMember, m.isMemberErr
}
func (m *mockGroupRepoForLeaderboard) GetGroupById(_ context.Context, _ string) (*entity.Group, error) {
	if m.guruIdErr != nil {
		return nil, errors.New("not found")
	}
	return &entity.Group{CreatedBy: m.groupCreator}, nil
}

// Satisfy full GroupRepository interface with no-ops.
func (m *mockGroupRepoForLeaderboard) CreateGroup(_ context.Context, g *entity.Group) (*entity.Group, error) {
	return g, nil
}
func (m *mockGroupRepoForLeaderboard) GetAllGroupsByGuru(_ context.Context, _ int, _, _ int, _ string) ([]*entity.Group, int, error) {
	return nil, 0, nil
}
func (m *mockGroupRepoForLeaderboard) UpdateGroup(_ context.Context, g *entity.Group) (*entity.Group, error) {
	return g, nil
}
func (m *mockGroupRepoForLeaderboard) DeleteGroup(_ context.Context, _ string) error { return nil }
func (m *mockGroupRepoForLeaderboard) AddMember(_ context.Context, _ string, _ int) error {
	return nil
}
func (m *mockGroupRepoForLeaderboard) RemoveMember(_ context.Context, _ string, _ int) error {
	return nil
}
func (m *mockGroupRepoForLeaderboard) GetGroupMembers(_ context.Context, _ string) ([]*entity.GroupMember, error) {
	return nil, nil
}
func (m *mockGroupRepoForLeaderboard) AddContent(_ context.Context, _ string, _ string, _ int) (*entity.GroupContent, error) {
	return nil, nil
}
func (m *mockGroupRepoForLeaderboard) GetContentsByGroupId(_ context.Context, _ string, _ string) ([]*entity.GroupContent, error) {
	return nil, nil
}
func (m *mockGroupRepoForLeaderboard) RemoveContent(_ context.Context, _ int, _ string) error {
	return nil
}
func (m *mockGroupRepoForLeaderboard) ContentExists(_ context.Context, _ string, _ int) (bool, error) {
	return false, nil
}
func (m *mockGroupRepoForLeaderboard) IsContentAlreadyAssigned(_ context.Context, _ string, _ string, _ int) (bool, error) {
	return false, nil
}
func (m *mockGroupRepoForLeaderboard) GetMemberIdByEmail(_ context.Context, _ string) (int, error) {
	return 0, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func newLeaderboardUC(lb *mockLeaderboardRepo, gr *mockGroupRepoForLeaderboard) LeaderboardUseCase {
	return NewLeaderboardUseCase(lb, gr, nil)
}

func defaultPublicReq() *model.PublicLeaderboardRequest {
	return &model.PublicLeaderboardRequest{Period: "alltime", Page: 1, Size: 20}
}

// ─── Public leaderboard tests ─────────────────────────────────────────────────

func TestNewLeaderboardUseCase(t *testing.T) {
	uc := NewLeaderboardUseCase(nil, nil, nil)
	if uc == nil {
		t.Fatal("expected usecase instance")
	}
}

func TestGetPublic_EmptyBoard(t *testing.T) {
	lb := &mockLeaderboardRepo{publicEntries: []entity.LeaderboardEntry{}, publicTotal: 0}
	uc := newLeaderboardUC(lb, nil)

	resp, err := uc.GetPublic(context.Background(), defaultPublicReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 0 || len(resp.Items) != 0 {
		t.Errorf("expected empty board, got total=%d items=%d", resp.Total, len(resp.Items))
	}
}

func TestGetPublic_Alltime_NoPeriodStart(t *testing.T) {
	lb := &mockLeaderboardRepo{
		publicEntries: []entity.LeaderboardEntry{{Rank: 1, MemberId: 1, TotalXP: 100, TotalCount: 1}},
		publicTotal:   1,
		publicPeriod:  time.Time{},
	}
	uc := newLeaderboardUC(lb, nil)

	resp, err := uc.GetPublic(context.Background(), defaultPublicReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PeriodStart != nil {
		t.Errorf("alltime period_start must be nil, got %s", *resp.PeriodStart)
	}
	if resp.Period != "alltime" {
		t.Errorf("expected period=alltime, got %s", resp.Period)
	}
}

func TestGetPublic_Monthly_HasPeriodStart(t *testing.T) {
	// period_start must include +07:00 offset for Jakarta
	jakarta := time.FixedZone("WIB", 7*3600)
	periodStart := time.Date(2026, 5, 1, 0, 0, 0, 0, jakarta)
	lb := &mockLeaderboardRepo{
		publicEntries: []entity.LeaderboardEntry{{Rank: 1, MemberId: 1, MonthlyXP: 50, TotalCount: 1}},
		publicTotal:   1,
		publicPeriod:  periodStart,
	}
	uc := newLeaderboardUC(lb, nil)

	resp, err := uc.GetPublic(context.Background(), &model.PublicLeaderboardRequest{
		Period: "monthly", Page: 1, Size: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PeriodStart == nil {
		t.Fatal("monthly period_start must not be nil")
	}
	if resp.Period != "monthly" {
		t.Errorf("expected period=monthly, got %s", resp.Period)
	}
	// Must contain +07:00 offset.
	if len(*resp.PeriodStart) == 0 {
		t.Error("period_start must not be empty")
	}
}

func TestGetPublic_Monthly_EmptyEarlyMonth(t *testing.T) {
	// Early in a new month — no completions yet → total:0, items:[], NOT 404
	lb := &mockLeaderboardRepo{publicEntries: []entity.LeaderboardEntry{}, publicTotal: 0, publicPeriod: time.Now()}
	uc := newLeaderboardUC(lb, nil)

	resp, err := uc.GetPublic(context.Background(), &model.PublicLeaderboardRequest{
		Period: "monthly", Page: 1, Size: 20,
	})
	if err != nil {
		t.Fatalf("expected no error for empty monthly board, got: %v", err)
	}
	if resp.Total != 0 || len(resp.Items) != 0 {
		t.Errorf("expected empty monthly board, got total=%d", resp.Total)
	}
}

func TestGetPublic_InvalidPeriod_Returns400(t *testing.T) {
	lb := &mockLeaderboardRepo{}
	uc := newLeaderboardUC(lb, nil)

	_, err := uc.GetPublic(context.Background(), &model.PublicLeaderboardRequest{
		Period: "weekly", Page: 1, Size: 20,
	})
	if err == nil {
		t.Fatal("expected 400 for invalid period")
	}
	fiberErr, ok := err.(*fiber.Error)
	if !ok || fiberErr.Code != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %v", err)
	}
}

func TestGetPublic_InvalidPagination_Returns400(t *testing.T) {
	lb := &mockLeaderboardRepo{}
	uc := newLeaderboardUC(lb, nil)

	cases := []model.PublicLeaderboardRequest{
		{Period: "alltime", Page: 0, Size: 20},
		{Period: "alltime", Page: 1, Size: 0},
		{Period: "alltime", Page: 1, Size: 101},
	}
	for _, req := range cases {
		req := req
		_, err := uc.GetPublic(context.Background(), &req)
		if err == nil {
			t.Errorf("expected 400 for page=%d size=%d", req.Page, req.Size)
			continue
		}
		fiberErr, ok := err.(*fiber.Error)
		if !ok || fiberErr.Code != fiber.StatusBadRequest {
			t.Errorf("expected 400 for page=%d size=%d, got %v", req.Page, req.Size, err)
		}
	}
}

func TestGetPublic_Ties_SameRank(t *testing.T) {
	// Two members with identical XP → both rank 1.
	lb := &mockLeaderboardRepo{
		publicEntries: []entity.LeaderboardEntry{
			{Rank: 1, MemberId: 1, TotalXP: 500, TotalCount: 2},
			{Rank: 1, MemberId: 2, TotalXP: 500, TotalCount: 2},
		},
		publicTotal: 2,
	}
	uc := newLeaderboardUC(lb, nil)

	resp, err := uc.GetPublic(context.Background(), defaultPublicReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Items[0].Rank != 1 || resp.Items[1].Rank != 1 {
		t.Errorf("tied members should both have rank 1, got %d and %d", resp.Items[0].Rank, resp.Items[1].Rank)
	}
}

func TestGetPublic_PaginationOffset(t *testing.T) {
	// Page 2 size 20 → ranks 21–40.
	entries := make([]entity.LeaderboardEntry, 20)
	for i := range entries {
		entries[i] = entity.LeaderboardEntry{Rank: 21 + i, MemberId: 21 + i, TotalCount: 50}
	}
	lb := &mockLeaderboardRepo{publicEntries: entries, publicTotal: 50}
	uc := newLeaderboardUC(lb, nil)

	resp, err := uc.GetPublic(context.Background(), &model.PublicLeaderboardRequest{
		Period: "alltime", Page: 2, Size: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 50 {
		t.Errorf("expected total=50, got %d", resp.Total)
	}
	if len(resp.Items) != 20 {
		t.Errorf("expected 20 items on page 2, got %d", len(resp.Items))
	}
	if resp.Items[0].Rank != 21 {
		t.Errorf("page 2 first rank should be 21, got %d", resp.Items[0].Rank)
	}
}

func TestGetPublic_OffsetExceedsTotal_EmptyItems(t *testing.T) {
	// Page 10 size 20 with only 5 total → empty items, not 404.
	lb := &mockLeaderboardRepo{publicEntries: []entity.LeaderboardEntry{}, publicTotal: 0}
	uc := newLeaderboardUC(lb, nil)

	resp, err := uc.GetPublic(context.Background(), &model.PublicLeaderboardRequest{
		Period: "alltime", Page: 10, Size: 20,
	})
	if err != nil {
		t.Fatalf("expected no error for empty page, got: %v", err)
	}
	if len(resp.Items) != 0 {
		t.Error("expected empty items when offset exceeds total")
	}
}

// ─── Group leaderboard tests ──────────────────────────────────────────────────

func TestGetGroup_NotFound_Returns404(t *testing.T) {
	lb := &mockLeaderboardRepo{headerFound: false}
	uc := newLeaderboardUC(lb, &mockGroupRepoForLeaderboard{})

	_, err := uc.GetGroup(context.Background(), "missing-group", &model.GroupLeaderboardRequest{Page: 1, Size: 20},
		&model.Claims{Role: "super_admin"})
	if err == nil {
		t.Fatal("expected 404 for missing group")
	}
	fiberErr, ok := err.(*fiber.Error)
	if !ok || fiberErr.Code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %v", err)
	}
}

func TestGetGroup_NonMember_Returns403(t *testing.T) {
	lb := &mockLeaderboardRepo{headerFound: true, headerName: "Kelas XI"}
	gr := &mockGroupRepoForLeaderboard{memberId: 99, isMember: false}
	uc := newLeaderboardUC(lb, gr)

	_, err := uc.GetGroup(context.Background(), "group-1", &model.GroupLeaderboardRequest{Page: 1, Size: 20},
		&model.Claims{Role: "member", UserId: "10"})
	if err == nil {
		t.Fatal("expected 403 for non-member")
	}
	fiberErr, ok := err.(*fiber.Error)
	if !ok || fiberErr.Code != fiber.StatusForbidden {
		t.Errorf("expected 403, got %v", err)
	}
}

func TestGetGroup_SuperAdmin_CanAccess(t *testing.T) {
	lb := &mockLeaderboardRepo{
		headerFound:  true,
		headerName:   "Kelas XI",
		groupEntries: []entity.LeaderboardEntry{},
		groupTotal:   0,
	}
	uc := newLeaderboardUC(lb, &mockGroupRepoForLeaderboard{})

	resp, err := uc.GetGroup(context.Background(), "group-1", &model.GroupLeaderboardRequest{Page: 1, Size: 20},
		&model.Claims{Role: "super_admin"})
	if err != nil {
		t.Fatalf("super_admin should access any group: %v", err)
	}
	if resp.Group.GroupName != "Kelas XI" {
		t.Errorf("expected group name 'Kelas XI', got %s", resp.Group.GroupName)
	}
}

func TestGetGroup_MemberWithAccess_CanAccess(t *testing.T) {
	lb := &mockLeaderboardRepo{
		headerFound:  true,
		headerName:   "Kelas XI",
		groupEntries: []entity.LeaderboardEntry{},
		groupTotal:   0,
	}
	gr := &mockGroupRepoForLeaderboard{memberId: 42, isMember: true}
	uc := newLeaderboardUC(lb, gr)

	resp, err := uc.GetGroup(context.Background(), "group-1", &model.GroupLeaderboardRequest{Page: 1, Size: 20},
		&model.Claims{Role: "member", UserId: "10"})
	if err != nil {
		t.Fatalf("member of group should access leaderboard: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response")
	}
}

func TestGetGroup_ZeroProgress_EntryStillPresent(t *testing.T) {
	// Member in group with no member_progress yet → appears with group_xp=0, completed_count=0.
	lb := &mockLeaderboardRepo{
		headerFound: true,
		headerName:  "Kelas XI",
		groupEntries: []entity.LeaderboardEntry{
			{Rank: 1, MemberId: 7, GroupXP: 0, CompletedCount: 0, TotalXP: 0, TotalCount: 1},
		},
		groupTotal: 1,
	}
	uc := newLeaderboardUC(lb, &mockGroupRepoForLeaderboard{})

	resp, err := uc.GetGroup(context.Background(), "group-1", &model.GroupLeaderboardRequest{Page: 1, Size: 20},
		&model.Claims{Role: "super_admin"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item (zero-progress member), got %d", len(resp.Items))
	}
	if resp.Items[0].GroupXP != 0 || resp.Items[0].CompletedCount != 0 {
		t.Errorf("zero-progress member should have group_xp=0 completed_count=0, got xp=%d count=%d",
			resp.Items[0].GroupXP, resp.Items[0].CompletedCount)
	}
}

func TestGetGroup_Ties_SameRank(t *testing.T) {
	// Two members with identical group_xp → both rank 1.
	lb := &mockLeaderboardRepo{
		headerFound: true,
		headerName:  "Kelas XI",
		groupEntries: []entity.LeaderboardEntry{
			{Rank: 1, MemberId: 1, GroupXP: 200, TotalCount: 2},
			{Rank: 1, MemberId: 2, GroupXP: 200, TotalCount: 2},
		},
		groupTotal: 2,
	}
	uc := newLeaderboardUC(lb, &mockGroupRepoForLeaderboard{})

	resp, err := uc.GetGroup(context.Background(), "group-1", &model.GroupLeaderboardRequest{Page: 1, Size: 20},
		&model.Claims{Role: "super_admin"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Items[0].Rank != 1 || resp.Items[1].Rank != 1 {
		t.Errorf("tied members should both have rank 1")
	}
}

func TestGetGroup_ResponseHasGroupHeader(t *testing.T) {
	thumb := "/uploads/thumb.webp"
	lb := &mockLeaderboardRepo{
		headerFound: true, headerName: "Kelas XI IPA 3", headerThumb: &thumb,
		groupEntries: []entity.LeaderboardEntry{}, groupTotal: 0,
	}
	uc := newLeaderboardUC(lb, &mockGroupRepoForLeaderboard{})

	resp, err := uc.GetGroup(context.Background(), "abc123", &model.GroupLeaderboardRequest{Page: 1, Size: 20},
		&model.Claims{Role: "super_admin"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Group == nil {
		t.Fatal("group header must be present")
	}
	if resp.Group.GroupId != "abc123" {
		t.Errorf("expected group_id=abc123, got %s", resp.Group.GroupId)
	}
	if resp.Group.GroupThumbnail == nil || *resp.Group.GroupThumbnail != thumb {
		t.Errorf("expected thumbnail %s", thumb)
	}
}

func TestGetGroup_InvalidPagination_Returns400(t *testing.T) {
	lb := &mockLeaderboardRepo{headerFound: true, headerName: "G"}
	uc := newLeaderboardUC(lb, &mockGroupRepoForLeaderboard{})
	claims := &model.Claims{Role: "super_admin"}

	cases := []model.GroupLeaderboardRequest{
		{Page: 0, Size: 20},
		{Page: 1, Size: 0},
		{Page: 1, Size: 101},
	}
	for _, req := range cases {
		req := req
		_, err := uc.GetGroup(context.Background(), "g1", &req, claims)
		if err == nil {
			t.Errorf("expected 400 for page=%d size=%d", req.Page, req.Size)
		}
	}
}
