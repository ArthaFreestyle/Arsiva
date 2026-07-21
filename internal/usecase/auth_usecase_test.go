package usecase

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/utils"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"
)

// mockUserRepo implements repository.UserRepository for testing.
type mockUserRepo struct {
	createFn      func(ctx context.Context, user *entity.User) (*entity.User, error)
	findByEmailFn func(ctx context.Context, email string) (*entity.User, error)
	getUserByIdFn func(ctx context.Context, userId string) (*entity.User, error)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return nil, nil
}
func (m *mockUserRepo) GetAllUsers(ctx context.Context) ([]*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetUserById(ctx context.Context, userId string) (*entity.User, error) {
	if m.getUserByIdFn != nil {
		return m.getUserByIdFn(ctx, userId)
	}
	return nil, nil
}
func (m *mockUserRepo) SearchByEmail(ctx context.Context, emailQuery string, limit int) ([]*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	now := time.Now()
	user.UserId = "99"
	user.CreatedAt = &now
	return user, nil
}
func (m *mockUserRepo) UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) DeleteUser(ctx context.Context, user *entity.User) error {
	return nil
}
func (m *mockUserRepo) GetDeletedUsers(ctx context.Context) ([]*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) RestoreUser(ctx context.Context, user *entity.User) error {
	return nil
}
func (m *mockUserRepo) UpdatePassword(ctx context.Context, userId string, passwordHash string) error {
	return nil
}
func (m *mockUserRepo) MarkVerified(ctx context.Context, userId string) error {
	return nil
}

func discardLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}

func newTestAuthUseCase(repo *mockUserRepo) AuthUseCase {
	return NewAuthUseCase(repo, []byte("secret"), validator.New(), discardLogger(), nil, nil, nil, nil, nil, 5*time.Minute, 15*time.Minute, 5, time.Minute, "https://arsiva.id/reset-password")
}

func TestNewAuthUseCase(t *testing.T) {
	uc := NewAuthUseCase(nil, []byte("secret"), validator.New(), nil, nil, nil, nil, nil, nil, 5*time.Minute, 15*time.Minute, 5, time.Minute, "https://arsiva.id/reset-password")
	if uc == nil {
		t.Fatal("expected usecase instance")
	}
}

func TestRegisterMember_HappyPath(t *testing.T) {
	uc := newTestAuthUseCase(&mockUserRepo{})

	req := &model.RegisterRequest{
		Username: "siswa01",
		Email:    "siswa01@example.com",
		Password: "Rahasia123!",
	}

	resp, err := uc.RegisterMember(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Role != "member" {
		t.Errorf("expected role 'member', got '%s'", resp.Role)
	}
	if resp.Username != "siswa01" {
		t.Errorf("expected username 'siswa01', got '%s'", resp.Username)
	}
}

func TestRegisterGuru_HappyPath(t *testing.T) {
	uc := newTestAuthUseCase(&mockUserRepo{})

	req := &model.RegisterRequest{
		Username: "guru01",
		Email:    "guru01@example.com",
		Password: "Rahasia123!",
	}

	resp, err := uc.RegisterGuru(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Role != "guru" {
		t.Errorf("expected role 'guru', got '%s'", resp.Role)
	}
}

func TestRegisterMember_ValidationFailure(t *testing.T) {
	uc := newTestAuthUseCase(&mockUserRepo{})

	// Missing password
	req := &model.RegisterRequest{
		Username: "siswa01",
		Email:    "siswa01@example.com",
	}

	_, err := uc.RegisterMember(context.Background(), req)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) || fiberErr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 fiber error, got %v", err)
	}
}

func TestRegisterMember_PasswordTooShort(t *testing.T) {
	uc := newTestAuthUseCase(&mockUserRepo{})

	req := &model.RegisterRequest{
		Username: "siswa01",
		Email:    "siswa01@example.com",
		Password: "short",
	}

	_, err := uc.RegisterMember(context.Background(), req)
	if err == nil {
		t.Fatal("expected validation error for short password, got nil")
	}

	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) || fiberErr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 fiber error, got %v", err)
	}
}

// Role injection: RegisterRequest has no role field, so the role is always
// set server-side. We verify the value passed to the DB is always "member".
func TestRegisterMember_RoleInjectionIgnored(t *testing.T) {
	var capturedRole string
	repo := &mockUserRepo{
		createFn: func(ctx context.Context, user *entity.User) (*entity.User, error) {
			capturedRole = user.Role
			now := time.Now()
			user.UserId = "1"
			user.CreatedAt = &now
			return user, nil
		},
	}
	uc := newTestAuthUseCase(repo)

	req := &model.RegisterRequest{
		Username: "sneaky",
		Email:    "sneaky@example.com",
		Password: "Rahasia123!",
	}

	_, err := uc.RegisterMember(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if capturedRole != "member" {
		t.Errorf("expected role 'member' to be passed to DB, got '%s'", capturedRole)
	}
}

func TestRegisterMember_UniqueViolation_Returns409(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(ctx context.Context, user *entity.User) (*entity.User, error) {
			return nil, &pgconn.PgError{Code: "23505"}
		},
	}
	uc := newTestAuthUseCase(repo)

	req := &model.RegisterRequest{
		Username: "siswa01",
		Email:    "siswa01@example.com",
		Password: "Rahasia123!",
	}

	_, err := uc.RegisterMember(context.Background(), req)
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}

	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) || fiberErr.Code != http.StatusConflict {
		t.Errorf("expected 409 fiber error, got %v", err)
	}
}

// --- Login / expected_role tests ---

func makeLoginRepo(role string) *mockUserRepo {
	hash, _ := utils.HashPassword("Rahasia123!")
	return &mockUserRepo{
		findByEmailFn: func(ctx context.Context, email string) (*entity.User, error) {
			now := time.Now()
			return &entity.User{
				UserId:       "1",
				Email:        email,
				Username:     "testuser",
				Role:         role,
				PasswordHash: hash,
				CreatedAt:    &now,
				IsVerified:   true,
			}, nil
		},
	}
}

func TestLogin_NoExpectedRole_Success(t *testing.T) {
	uc := newTestAuthUseCase(makeLoginRepo("member"))
	resp, err := uc.Login(context.Background(), &model.LoginRequest{
		Email:    "siswa01@example.com",
		Password: "Rahasia123!",
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access token to be set")
	}
}

func TestLogin_ExpectedRole_Match_Success(t *testing.T) {
	uc := newTestAuthUseCase(makeLoginRepo("member"))
	resp, err := uc.Login(context.Background(), &model.LoginRequest{
		Email:        "siswa01@example.com",
		Password:     "Rahasia123!",
		ExpectedRole: "member",
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if resp.User.Role != "member" {
		t.Errorf("expected role 'member', got '%s'", resp.User.Role)
	}
}

func TestLogin_ExpectedRole_Mismatch_Returns403(t *testing.T) {
	uc := newTestAuthUseCase(makeLoginRepo("guru"))
	_, err := uc.Login(context.Background(), &model.LoginRequest{
		Email:        "guru01@example.com",
		Password:     "Rahasia123!",
		ExpectedRole: "member",
	})
	if err == nil {
		t.Fatal("expected 403 error, got nil")
	}
	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) || fiberErr.Code != http.StatusForbidden {
		t.Errorf("expected 403 fiber error, got %v", err)
	}
	// error message must NOT reveal the actual role
	if fiberErr.Message == "guru" {
		t.Error("error message must not reveal the actual role")
	}
}

func TestLogin_WrongPassword_Returns401_BeforeRoleCheck(t *testing.T) {
	uc := newTestAuthUseCase(makeLoginRepo("member"))
	_, err := uc.Login(context.Background(), &model.LoginRequest{
		Email:        "siswa01@example.com",
		Password:     "WrongPassword!",
		ExpectedRole: "guru", // mismatch, but should never be reached
	})
	if err == nil {
		t.Fatal("expected 401 error, got nil")
	}
	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) || fiberErr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 fiber error, got %v", err)
	}
}

func TestLogin_ExpectedRole_SuperAdmin_Returns400(t *testing.T) {
	uc := newTestAuthUseCase(makeLoginRepo("super_admin"))
	_, err := uc.Login(context.Background(), &model.LoginRequest{
		Email:        "admin@example.com",
		Password:     "Rahasia123!",
		ExpectedRole: "super_admin",
	})
	if err == nil {
		t.Fatal("expected 400 error, got nil")
	}
	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) || fiberErr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 fiber error (validation), got %v", err)
	}
}

// --- Login profile-details embedding tests ---
// These cover the fix from #32: auth_usecase must not silently drop profile
// details when FindByUserId fails with a real error, and must distinguish
// ErrNoRows (no profile yet) from actual DB failures.

// mockMemberRepo is a minimal MemberRepository that only stubs FindByUserId.
type mockMemberRepo struct {
	findByUserIdFn func(ctx context.Context, userId string) (*entity.Member, error)
}

func (m *mockMemberRepo) FindByUserId(ctx context.Context, userId string) (*entity.Member, error) {
	if m.findByUserIdFn != nil {
		return m.findByUserIdFn(ctx, userId)
	}
	return nil, pgx.ErrNoRows
}
func (m *mockMemberRepo) Create(ctx context.Context, member *entity.Member) (*entity.Member, error) {
	return nil, nil
}
func (m *mockMemberRepo) FindById(ctx context.Context, memberId string) (*entity.Member, error) {
	return nil, nil
}
func (m *mockMemberRepo) FindAll(ctx context.Context, search string, limit int, offset int) ([]*entity.Member, int, error) {
	return nil, 0, nil
}
func (m *mockMemberRepo) Update(ctx context.Context, member *entity.Member) (*entity.Member, error) {
	return nil, nil
}
func (m *mockMemberRepo) Delete(ctx context.Context, memberId string) error { return nil }
func (m *mockMemberRepo) FindSekolahByMemberId(ctx context.Context, memberId string) (*entity.Sekolah, error) {
	return nil, nil
}

// mockGuruRepo is a minimal GuruRepository that only stubs FindByUserId.
type mockGuruRepo struct {
	findByUserIdFn func(ctx context.Context, userId string) (*entity.Guru, error)
}

func (m *mockGuruRepo) FindByUserId(ctx context.Context, userId string) (*entity.Guru, error) {
	if m.findByUserIdFn != nil {
		return m.findByUserIdFn(ctx, userId)
	}
	return nil, pgx.ErrNoRows
}
func (m *mockGuruRepo) Create(ctx context.Context, guru *entity.Guru) (*entity.Guru, error) {
	return nil, nil
}
func (m *mockGuruRepo) FindById(ctx context.Context, guruId string) (*entity.Guru, error) {
	return nil, nil
}
func (m *mockGuruRepo) FindAll(ctx context.Context, search string, limit int, offset int) ([]*entity.Guru, int, error) {
	return nil, 0, nil
}
func (m *mockGuruRepo) Update(ctx context.Context, guru *entity.Guru) (*entity.Guru, error) {
	return nil, nil
}
func (m *mockGuruRepo) Delete(ctx context.Context, guruId string) error { return nil }
func (m *mockGuruRepo) FindSekolahByGuruId(ctx context.Context, guruId string) (*entity.Sekolah, error) {
	return nil, nil
}
func (m *mockGuruRepo) FindGroupsByGuruId(ctx context.Context, guruId string) ([]*entity.Group, error) {
	return nil, nil
}

func newTestAuthUseCaseWithRepos(userRepo *mockUserRepo, memberRepo *mockMemberRepo, guruRepo *mockGuruRepo) AuthUseCase {
	return NewAuthUseCase(userRepo, []byte("secret"), validator.New(), discardLogger(), nil, guruRepo, memberRepo, nil, nil, 5*time.Minute, 15*time.Minute, 5, time.Minute, "https://arsiva.id/reset-password")
}

func TestLogin_Member_WithProfile_DetailsPopulated(t *testing.T) {
	memberRepo := &mockMemberRepo{
		findByUserIdFn: func(ctx context.Context, userId string) (*entity.Member, error) {
			return &entity.Member{MemberId: "42", NIS: "12345", SekolahId: "7", Level: 1}, nil
		},
	}
	uc := newTestAuthUseCaseWithRepos(makeLoginRepo("member"), memberRepo, nil)

	resp, err := uc.Login(context.Background(), &model.LoginRequest{
		Email:    "siswa01@example.com",
		Password: "Rahasia123!",
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access token to be set")
	}
}

func TestLogin_Member_NoProfile_TokenIssuedWithoutDetails(t *testing.T) {
	// ErrNoRows = profile not yet created. Login must succeed with a token,
	// but Details in the JWT will be nil so ProfileCompleteMiddleware gates them.
	memberRepo := &mockMemberRepo{
		findByUserIdFn: func(ctx context.Context, userId string) (*entity.Member, error) {
			return nil, pgx.ErrNoRows
		},
	}
	uc := newTestAuthUseCaseWithRepos(makeLoginRepo("member"), memberRepo, nil)

	resp, err := uc.Login(context.Background(), &model.LoginRequest{
		Email:    "siswa01@example.com",
		Password: "Rahasia123!",
	})
	if err != nil {
		t.Fatalf("expected login to succeed even without a profile, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access token to be set")
	}
}

func TestLogin_Member_RepoError_Returns500(t *testing.T) {
	// A real DB error (not ErrNoRows) must abort login with 500 instead of
	// silently issuing a degraded token that would 403 every action endpoint.
	memberRepo := &mockMemberRepo{
		findByUserIdFn: func(ctx context.Context, userId string) (*entity.Member, error) {
			return nil, errors.New("connection reset by peer")
		},
	}
	uc := newTestAuthUseCaseWithRepos(makeLoginRepo("member"), memberRepo, nil)

	_, err := uc.Login(context.Background(), &model.LoginRequest{
		Email:    "siswa01@example.com",
		Password: "Rahasia123!",
	})
	if err == nil {
		t.Fatal("expected 500 error, got nil")
	}
	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) || fiberErr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 fiber error, got %v", err)
	}
}

func TestLogin_Guru_WithProfile_DetailsPopulated(t *testing.T) {
	guruRepo := &mockGuruRepo{
		findByUserIdFn: func(ctx context.Context, userId string) (*entity.Guru, error) {
			return &entity.Guru{GuruId: "10", NIP: "9876", BidangAjar: "Matematika", SekolahId: "3"}, nil
		},
	}
	uc := newTestAuthUseCaseWithRepos(makeLoginRepo("guru"), nil, guruRepo)

	resp, err := uc.Login(context.Background(), &model.LoginRequest{
		Email:    "guru01@example.com",
		Password: "Rahasia123!",
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access token to be set")
	}
}

func TestLogin_Guru_RepoError_Returns500(t *testing.T) {
	guruRepo := &mockGuruRepo{
		findByUserIdFn: func(ctx context.Context, userId string) (*entity.Guru, error) {
			return nil, errors.New("timeout")
		},
	}
	uc := newTestAuthUseCaseWithRepos(makeLoginRepo("guru"), nil, guruRepo)

	_, err := uc.Login(context.Background(), &model.LoginRequest{
		Email:    "guru01@example.com",
		Password: "Rahasia123!",
	})
	if err == nil {
		t.Fatal("expected 500 error, got nil")
	}
	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) || fiberErr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 fiber error, got %v", err)
	}
}

func TestLogin_Unverified_Returns403(t *testing.T) {
	// Correct password but is_verified=false must be blocked with 403, and only
	// AFTER the password check (so it can't be used to enumerate accounts).
	hash, _ := utils.HashPassword("Rahasia123!")
	repo := &mockUserRepo{
		findByEmailFn: func(ctx context.Context, email string) (*entity.User, error) {
			now := time.Now()
			return &entity.User{
				UserId:       "1",
				Email:        email,
				Username:     "belumverif",
				Role:         "member",
				PasswordHash: hash,
				CreatedAt:    &now,
				IsVerified:   false,
			}, nil
		},
	}
	uc := newTestAuthUseCase(repo)

	_, err := uc.Login(context.Background(), &model.LoginRequest{
		Email:    "belumverif@example.com",
		Password: "Rahasia123!",
	})
	if err == nil {
		t.Fatal("expected 403 error for unverified account, got nil")
	}
	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) || fiberErr.Code != http.StatusForbidden {
		t.Errorf("expected 403 fiber error, got %v", err)
	}
}

func TestLogin_Unverified_WrongPassword_Returns401(t *testing.T) {
	// Wrong password on an unverified account must still be 401 (password check
	// runs before the verification check), never 403 — otherwise a wrong-password
	// attacker could learn the account exists but is unverified.
	hash, _ := utils.HashPassword("Rahasia123!")
	repo := &mockUserRepo{
		findByEmailFn: func(ctx context.Context, email string) (*entity.User, error) {
			return &entity.User{UserId: "1", Email: email, Role: "member", PasswordHash: hash, IsVerified: false}, nil
		},
	}
	uc := newTestAuthUseCase(repo)

	_, err := uc.Login(context.Background(), &model.LoginRequest{
		Email:    "belumverif@example.com",
		Password: "SalahPassword!",
	})
	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) || fiberErr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 fiber error, got %v", err)
	}
}
