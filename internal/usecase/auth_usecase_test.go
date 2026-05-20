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

func discardLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}

func newTestAuthUseCase(repo *mockUserRepo) AuthUseCase {
	return NewAuthUseCase(repo, []byte("secret"), validator.New(), discardLogger(), nil, nil, nil)
}

func TestNewAuthUseCase(t *testing.T) {
	uc := NewAuthUseCase(nil, []byte("secret"), validator.New(), nil, nil, nil, nil)
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
