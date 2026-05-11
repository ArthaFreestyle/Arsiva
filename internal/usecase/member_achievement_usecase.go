package usecase

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
)

type MemberAchievementUseCase interface {
	Create(ctx context.Context, req *model.MemberAchievementCreateRequest, claims *model.Claims) (*model.MemberAchievementResponse, error)
	FindAllMine(ctx context.Context, claims *model.Claims) ([]*model.MemberAchievementResponse, error)
	FindOne(ctx context.Context, achievementId string, claims *model.Claims) (*model.MemberAchievementResponse, error)
	Delete(ctx context.Context, memberId, achievementId string, claims *model.Claims) error
}

type memberAchievementUseCaseImpl struct {
	Repo            repository.MemberAchievementRepository
	MemberRepo      repository.MemberRepository
	AchievementRepo repository.AchievementRepository
	Log             *logrus.Logger
	Validator       *validator.Validate
}

func NewMemberAchievementUseCase(
	repo repository.MemberAchievementRepository,
	memberRepo repository.MemberRepository,
	achievementRepo repository.AchievementRepository,
	log *logrus.Logger,
	validate *validator.Validate,
) MemberAchievementUseCase {
	return &memberAchievementUseCaseImpl{
		Repo:            repo,
		MemberRepo:      memberRepo,
		AchievementRepo: achievementRepo,
		Log:             log,
		Validator:       validate,
	}
}

func (u *memberAchievementUseCaseImpl) Create(ctx context.Context, req *model.MemberAchievementCreateRequest, claims *model.Claims) (*model.MemberAchievementResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Invalid request body: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}

	ach, err := u.AchievementRepo.FindById(ctx, req.AchievementId)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "achievement tidak ditemukan")
	}

	member, err := u.MemberRepo.FindById(ctx, memberId)
	if err != nil {
		return nil, fiber.ErrForbidden
	}

	if member.TotalXP < ach.XPRequired {
		return nil, fiber.NewError(fiber.StatusForbidden, "XP belum mencukupi untuk membuka achievement ini")
	}

	exists, err := u.Repo.Exists(ctx, memberId, req.AchievementId)
	if err != nil {
		u.Log.Warnf("Failed check achievement exists: %v", err)
		return nil, fiber.ErrInternalServerError
	}
	if exists {
		return nil, fiber.NewError(fiber.StatusConflict, "achievement sudah pernah di-unlock")
	}

	result, err := u.Repo.Create(ctx, memberId, req.AchievementId)
	if err != nil {
		if errors.Is(err, repository.ErrMemberAchievementExists) {
			return nil, fiber.NewError(fiber.StatusConflict, "achievement sudah pernah di-unlock")
		}
		if errors.Is(err, repository.ErrAchievementFKViolation) {
			return nil, fiber.NewError(fiber.StatusNotFound, "achievement tidak ditemukan")
		}
		u.Log.Errorf("Failed create member achievement: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToMemberAchievementResponse(result), nil
}

func (u *memberAchievementUseCaseImpl) FindAllMine(ctx context.Context, claims *model.Claims) ([]*model.MemberAchievementResponse, error) {
	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}

	achievements, err := u.Repo.FindAllByMemberId(ctx, memberId)
	if err != nil {
		u.Log.Warnf("Failed get member achievements: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToMemberAchievementResponses(achievements), nil
}

func (u *memberAchievementUseCaseImpl) FindOne(ctx context.Context, achievementId string, claims *model.Claims) (*model.MemberAchievementResponse, error) {
	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}

	result, err := u.Repo.FindOne(ctx, memberId, achievementId)
	if err != nil {
		if errors.Is(err, repository.ErrMemberAchievementNotFound) {
			return nil, fiber.ErrNotFound
		}
		u.Log.Warnf("Failed get member achievement: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToMemberAchievementResponse(result), nil
}

func (u *memberAchievementUseCaseImpl) Delete(ctx context.Context, memberId, achievementId string, claims *model.Claims) error {
	if claims.Role != "super_admin" {
		return fiber.ErrForbidden
	}

	if err := u.Repo.Delete(ctx, memberId, achievementId); err != nil {
		if errors.Is(err, repository.ErrMemberAchievementNotFound) {
			return fiber.ErrNotFound
		}
		u.Log.Warnf("Failed delete member achievement: %v", err)
		return fiber.ErrInternalServerError
	}

	return nil
}
