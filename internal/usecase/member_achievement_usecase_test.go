package usecase

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/repository"
)

// ==================== Inline mocks ====================

type mockMemberAchievementRepo struct {
	findAllByMemberIdFn func(ctx context.Context, memberId string) ([]*entity.MemberAchievement, error)
	findOneFn           func(ctx context.Context, memberId, achievementId string) (*entity.MemberAchievement, error)
	existsFn            func(ctx context.Context, memberId, achievementId string) (bool, error)
	createFn            func(ctx context.Context, memberId, achievementId string) (*entity.MemberAchievement, error)
	deleteFn            func(ctx context.Context, memberId, achievementId string) error
	createCallCount     int
}

func (m *mockMemberAchievementRepo) FindAllByMemberId(ctx context.Context, memberId string) ([]*entity.MemberAchievement, error) {
	return m.findAllByMemberIdFn(ctx, memberId)
}
func (m *mockMemberAchievementRepo) FindOne(ctx context.Context, memberId, achievementId string) (*entity.MemberAchievement, error) {
	return m.findOneFn(ctx, memberId, achievementId)
}
func (m *mockMemberAchievementRepo) Exists(ctx context.Context, memberId, achievementId string) (bool, error) {
	return m.existsFn(ctx, memberId, achievementId)
}
func (m *mockMemberAchievementRepo) Create(ctx context.Context, memberId, achievementId string) (*entity.MemberAchievement, error) {
	m.createCallCount++
	return m.createFn(ctx, memberId, achievementId)
}
func (m *mockMemberAchievementRepo) Delete(ctx context.Context, memberId, achievementId string) error {
	return m.deleteFn(ctx, memberId, achievementId)
}

type mockMemberRepoForAch struct {
	findByIdFn func(ctx context.Context, memberId string) (*entity.Member, error)
}

func (m *mockMemberRepoForAch) Create(ctx context.Context, member *entity.Member) (*entity.Member, error) {
	return nil, nil
}
func (m *mockMemberRepoForAch) FindById(ctx context.Context, memberId string) (*entity.Member, error) {
	return m.findByIdFn(ctx, memberId)
}
func (m *mockMemberRepoForAch) FindByUserId(ctx context.Context, userId string) (*entity.Member, error) {
	return nil, nil
}
func (m *mockMemberRepoForAch) FindAll(ctx context.Context, search string, limit int, offset int) ([]*entity.Member, int, error) {
	return nil, 0, nil
}
func (m *mockMemberRepoForAch) Update(ctx context.Context, member *entity.Member) (*entity.Member, error) {
	return nil, nil
}
func (m *mockMemberRepoForAch) Delete(ctx context.Context, memberId string) error { return nil }
func (m *mockMemberRepoForAch) FindSekolahByMemberId(ctx context.Context, memberId string) (*entity.Sekolah, error) {
	return nil, nil
}

type mockAchievementRepoForAch struct {
	findByIdFn func(ctx context.Context, achievementId string) (*entity.Achievement, error)
}

func (m *mockAchievementRepoForAch) Create(ctx context.Context, ach *entity.Achievement) (*entity.Achievement, error) {
	return nil, nil
}
func (m *mockAchievementRepoForAch) FindById(ctx context.Context, achievementId string) (*entity.Achievement, error) {
	return m.findByIdFn(ctx, achievementId)
}
func (m *mockAchievementRepoForAch) FindByNama(ctx context.Context, nama string) (*entity.Achievement, error) {
	return nil, nil
}
func (m *mockAchievementRepoForAch) FindAll(ctx context.Context, search string, tier string, limit int, offset int) ([]*entity.Achievement, int, error) {
	return nil, 0, nil
}
func (m *mockAchievementRepoForAch) Update(ctx context.Context, ach *entity.Achievement) (*entity.Achievement, error) {
	return nil, nil
}
func (m *mockAchievementRepoForAch) Delete(ctx context.Context, achievementId string) error {
	return nil
}
func (m *mockAchievementRepoForAch) CountMembersAwarded(ctx context.Context, achievementId string) (int, error) {
	return 0, nil
}

// ==================== Helpers ====================

func newTestMemberAchievementUseCase(
	repo repository.MemberAchievementRepository,
	memberRepo repository.MemberRepository,
	achRepo repository.AchievementRepository,
) MemberAchievementUseCase {
	return NewMemberAchievementUseCase(repo, memberRepo, achRepo, logrus.New(), validator.New())
}

func memberClaims(memberId string) *model.Claims {
	return &model.Claims{
		Role:    "member",
		Details: map[string]interface{}{"member_id": memberId},
	}
}

func superAdminClaims() *model.Claims {
	return &model.Claims{Role: "super_admin"}
}

// ==================== Tests ====================

func TestNewMemberAchievementUseCase(t *testing.T) {
	uc := NewMemberAchievementUseCase(nil, nil, nil, logrus.New(), validator.New())
	if uc == nil {
		t.Fatal("expected usecase instance")
	}
}

func TestMemberAchievementCreate_MissingAchievementId(t *testing.T) {
	uc := newTestMemberAchievementUseCase(nil, nil, nil)
	req := &model.MemberAchievementCreateRequest{AchievementId: ""}
	_, err := uc.Create(context.Background(), req, memberClaims("7"))
	if err == nil {
		t.Fatal("expected error for missing achievement_id")
	}
}

func TestMemberAchievementCreate_NonNumericAchievementId(t *testing.T) {
	uc := newTestMemberAchievementUseCase(nil, nil, nil)
	req := &model.MemberAchievementCreateRequest{AchievementId: "abc"}
	_, err := uc.Create(context.Background(), req, memberClaims("7"))
	if err == nil {
		t.Fatal("expected error for non-numeric achievement_id")
	}
}

func TestMemberAchievementCreate_EmptyMemberIdInClaims(t *testing.T) {
	uc := newTestMemberAchievementUseCase(nil, nil, nil)
	req := &model.MemberAchievementCreateRequest{AchievementId: "1"}
	_, err := uc.Create(context.Background(), req, &model.Claims{Role: "member"})
	if err == nil {
		t.Fatal("expected forbidden for empty member_id in claims")
	}
}

func TestMemberAchievementCreate_AchievementNotFound(t *testing.T) {
	achRepo := &mockAchievementRepoForAch{
		findByIdFn: func(ctx context.Context, achievementId string) (*entity.Achievement, error) {
			return nil, repository.ErrAchievementFKViolation
		},
	}
	uc := newTestMemberAchievementUseCase(nil, nil, achRepo)
	req := &model.MemberAchievementCreateRequest{AchievementId: "999"}
	_, err := uc.Create(context.Background(), req, memberClaims("7"))
	if err == nil {
		t.Fatal("expected 404 when achievement not found")
	}
}

func TestMemberAchievementCreate_XPGateBlocks(t *testing.T) {
	achRepo := &mockAchievementRepoForAch{
		findByIdFn: func(ctx context.Context, achievementId string) (*entity.Achievement, error) {
			return &entity.Achievement{AchievementId: "2", XPRequired: 500}, nil
		},
	}
	memberRepo := &mockMemberRepoForAch{
		findByIdFn: func(ctx context.Context, memberId string) (*entity.Member, error) {
			return &entity.Member{MemberId: memberId, TotalXP: 320}, nil
		},
	}
	mockRepo := &mockMemberAchievementRepo{}
	uc := newTestMemberAchievementUseCase(mockRepo, memberRepo, achRepo)

	req := &model.MemberAchievementCreateRequest{AchievementId: "2"}
	_, err := uc.Create(context.Background(), req, memberClaims("7"))

	if err == nil {
		t.Fatal("expected forbidden when XP insufficient")
	}
	if mockRepo.createCallCount != 0 {
		t.Fatalf("repo Create must not be called when XP gate blocks; got %d calls", mockRepo.createCallCount)
	}
}

func TestMemberAchievementCreate_AlreadyUnlocked(t *testing.T) {
	achRepo := &mockAchievementRepoForAch{
		findByIdFn: func(ctx context.Context, achievementId string) (*entity.Achievement, error) {
			return &entity.Achievement{AchievementId: "2", XPRequired: 100}, nil
		},
	}
	memberRepo := &mockMemberRepoForAch{
		findByIdFn: func(ctx context.Context, memberId string) (*entity.Member, error) {
			return &entity.Member{MemberId: memberId, TotalXP: 500}, nil
		},
	}
	mockRepo := &mockMemberAchievementRepo{
		existsFn: func(ctx context.Context, memberId, achievementId string) (bool, error) {
			return true, nil
		},
	}
	uc := newTestMemberAchievementUseCase(mockRepo, memberRepo, achRepo)

	req := &model.MemberAchievementCreateRequest{AchievementId: "2"}
	_, err := uc.Create(context.Background(), req, memberClaims("7"))

	if err == nil {
		t.Fatal("expected conflict when already unlocked")
	}
	if mockRepo.createCallCount != 0 {
		t.Fatalf("repo Create must not be called on duplicate; got %d calls", mockRepo.createCallCount)
	}
}

func TestMemberAchievementCreate_HappyPath(t *testing.T) {
	achRepo := &mockAchievementRepoForAch{
		findByIdFn: func(ctx context.Context, achievementId string) (*entity.Achievement, error) {
			return &entity.Achievement{AchievementId: "2", Nama: "Penjelajah Awal", XPRequired: 100}, nil
		},
	}
	memberRepo := &mockMemberRepoForAch{
		findByIdFn: func(ctx context.Context, memberId string) (*entity.Member, error) {
			return &entity.Member{MemberId: memberId, TotalXP: 500}, nil
		},
	}
	mockRepo := &mockMemberAchievementRepo{
		existsFn: func(ctx context.Context, memberId, achievementId string) (bool, error) {
			return false, nil
		},
		createFn: func(ctx context.Context, memberId, achievementId string) (*entity.MemberAchievement, error) {
			return &entity.MemberAchievement{
				AchievementId: "2",
				Nama:          "Penjelajah Awal",
				XPRequired:    100,
				Tier:          "silver",
				UnlockedAt:    "2026-05-11 10:14:33",
			}, nil
		},
	}
	uc := newTestMemberAchievementUseCase(mockRepo, memberRepo, achRepo)

	req := &model.MemberAchievementCreateRequest{AchievementId: "2"}
	resp, err := uc.Create(context.Background(), req, memberClaims("7"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.AchievementId != "2" {
		t.Fatal("expected populated response")
	}
	if mockRepo.createCallCount != 1 {
		t.Fatalf("expected exactly 1 Create call; got %d", mockRepo.createCallCount)
	}
}

func TestMemberAchievementFindAllMine_EmptyReturnsSlice(t *testing.T) {
	mockRepo := &mockMemberAchievementRepo{
		findAllByMemberIdFn: func(ctx context.Context, memberId string) ([]*entity.MemberAchievement, error) {
			return []*entity.MemberAchievement{}, nil
		},
	}
	uc := newTestMemberAchievementUseCase(mockRepo, nil, nil)

	resp, err := uc.FindAllMine(context.Background(), memberClaims("7"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil slice for empty results")
	}
}

func TestMemberAchievementDelete_NonSuperAdminForbidden(t *testing.T) {
	uc := newTestMemberAchievementUseCase(nil, nil, nil)
	err := uc.Delete(context.Background(), "7", "2", memberClaims("7"))
	if err == nil {
		t.Fatal("expected forbidden for non-superadmin delete")
	}
}

func TestMemberAchievementDelete_NotFound(t *testing.T) {
	mockRepo := &mockMemberAchievementRepo{
		deleteFn: func(ctx context.Context, memberId, achievementId string) error {
			return repository.ErrMemberAchievementNotFound
		},
	}
	uc := newTestMemberAchievementUseCase(mockRepo, nil, nil)

	err := uc.Delete(context.Background(), "7", "2", superAdminClaims())
	if err == nil {
		t.Fatal("expected 404 when row not found")
	}
}
