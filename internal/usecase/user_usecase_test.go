package usecase

import (
	"context"
	"testing"

	"ArthaFreestyle/Arsiva/internal/entity"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

func TestNewUserUseCase(t *testing.T) {
	uc := NewUserUseCase(nil, nil, nil, validator.New())
	if uc == nil {
		t.Fatal("expected usecase instance")
	}
}

func TestGetDeletedUsers_ReturnsEmptySlice(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewUserUseCase(repo, discardLogger(), nil, validator.New())

	users, err := uc.GetDeletedUsers(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("expected empty slice, got %d users", len(users))
	}
}

func TestRestoreUser_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		getUserByIdFn: func(ctx context.Context, userId string) (*entity.User, error) {
			return nil, pgx.ErrNoRows
		},
	}
	uc := NewUserUseCase(repo, discardLogger(), nil, validator.New())

	_, err := uc.RestoreUser(context.Background(), "non-existent-id")
	if err != fiber.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRestoreUser_AlreadyActive_ReturnsConflict(t *testing.T) {
	activeUser := &entity.User{UserId: "abc", Username: "alice", IsActive: true}
	repo := &mockUserRepo{
		getUserByIdFn: func(ctx context.Context, userId string) (*entity.User, error) {
			return activeUser, nil
		},
	}
	uc := NewUserUseCase(repo, discardLogger(), nil, validator.New())

	_, err := uc.RestoreUser(context.Background(), "abc")
	if err != fiber.ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestRestoreUser_Success(t *testing.T) {
	inactiveUser := &entity.User{UserId: "xyz", Username: "bob", IsActive: false}
	repo := &mockUserRepo{
		getUserByIdFn: func(ctx context.Context, userId string) (*entity.User, error) {
			return inactiveUser, nil
		},
	}
	uc := NewUserUseCase(repo, discardLogger(), nil, validator.New())

	res, err := uc.RestoreUser(context.Background(), "xyz")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.ID != "xyz" {
		t.Fatalf("expected user id xyz, got %s", res.ID)
	}
}
