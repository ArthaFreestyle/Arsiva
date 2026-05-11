package usecase

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
)

type AchievementUseCase interface {
	Create(ctx context.Context, req *model.AchievementCreateRequest) (*model.AchievementResponse, error)
	FindById(ctx context.Context, achievementId string) (*model.AchievementResponse, error)
	FindAll(ctx context.Context, search string, tier string, page int, size int) ([]*model.AchievementResponse, int, error)
	Update(ctx context.Context, achievementId string, req *model.AchievementUpdateRequest) (*model.AchievementResponse, error)
	Delete(ctx context.Context, achievementId string) error
}

type achievementUseCaseImpl struct {
	AchievementRepository repository.AchievementRepository
	Log                   *logrus.Logger
	Validator             *validator.Validate
}

func NewAchievementUseCase(
	achievementRepo repository.AchievementRepository,
	log *logrus.Logger,
	validate *validator.Validate,
) AchievementUseCase {
	return &achievementUseCaseImpl{
		AchievementRepository: achievementRepo,
		Log:                   log,
		Validator:             validate,
	}
}

func (u *achievementUseCaseImpl) Create(ctx context.Context, req *model.AchievementCreateRequest) (*model.AchievementResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Invalid request body: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	_, err := u.AchievementRepository.FindByNama(ctx, req.Nama)
	if err == nil {
		return nil, fiber.NewError(fiber.StatusConflict, "nama achievement sudah terdaftar")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		u.Log.Warnf("Failed check achievement nama: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	ach := &entity.Achievement{
		Nama:       req.Nama,
		Deskripsi:  req.Deskripsi,
		BadgeIcon:  req.BadgeIcon,
		XPRequired: req.XPRequired,
		Tier:       entity.TierAchievement(req.Tier),
	}

	result, err := u.AchievementRepository.Create(ctx, ach)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, fiber.NewError(fiber.StatusConflict, "nama achievement sudah terdaftar")
		}
		u.Log.Warnf("Failed create achievement: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToAchievementResponse(result), nil
}

func (u *achievementUseCaseImpl) FindById(ctx context.Context, achievementId string) (*model.AchievementResponse, error) {
	ach, err := u.AchievementRepository.FindById(ctx, achievementId)
	if err != nil {
		u.Log.Warnf("Achievement not found: %v", err)
		return nil, fiber.ErrNotFound
	}
	return converter.ToAchievementResponse(ach), nil
}

func (u *achievementUseCaseImpl) FindAll(ctx context.Context, search string, tier string, page int, size int) ([]*model.AchievementResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	// Tier filter hanya diteruskan jika nilainya valid — nilai lain diabaikan (bukan 400/500)
	validTiers := map[string]bool{"bronze": true, "silver": true, "gold": true, "platinum": true}
	if tier != "" && !validTiers[tier] {
		tier = ""
	}

	offset := (page - 1) * size

	achievements, total, err := u.AchievementRepository.FindAll(ctx, search, tier, size, offset)
	if err != nil {
		u.Log.Warnf("Failed get all achievement: %v", err)
		return nil, 0, fiber.ErrInternalServerError
	}

	return converter.ToAchievementResponses(achievements), total, nil
}

func (u *achievementUseCaseImpl) Update(ctx context.Context, achievementId string, req *model.AchievementUpdateRequest) (*model.AchievementResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Invalid request body: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	existing, err := u.AchievementRepository.FindById(ctx, achievementId)
	if err != nil {
		u.Log.Warnf("Achievement not found: %v", err)
		return nil, fiber.ErrNotFound
	}

	if req.Nama != existing.Nama {
		_, err = u.AchievementRepository.FindByNama(ctx, req.Nama)
		if err == nil {
			return nil, fiber.NewError(fiber.StatusConflict, "nama achievement sudah terdaftar")
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			u.Log.Warnf("Failed check achievement nama: %v", err)
			return nil, fiber.ErrInternalServerError
		}
	}

	ach := &entity.Achievement{
		AchievementId: achievementId,
		Nama:          req.Nama,
		Deskripsi:     req.Deskripsi,
		BadgeIcon:     req.BadgeIcon,
		XPRequired:    req.XPRequired,
		Tier:          entity.TierAchievement(req.Tier),
	}

	result, err := u.AchievementRepository.Update(ctx, ach)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, fiber.NewError(fiber.StatusConflict, "nama achievement sudah terdaftar")
		}
		u.Log.Warnf("Failed update achievement: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToAchievementResponse(result), nil
}

func (u *achievementUseCaseImpl) Delete(ctx context.Context, achievementId string) error {
	_, err := u.AchievementRepository.FindById(ctx, achievementId)
	if err != nil {
		u.Log.Warnf("Achievement not found: %v", err)
		return fiber.ErrNotFound
	}

	count, err := u.AchievementRepository.CountMembersAwarded(ctx, achievementId)
	if err != nil {
		u.Log.Warnf("Failed count members awarded: %v", err)
		return fiber.ErrInternalServerError
	}
	if count > 0 {
		return fiber.NewError(fiber.StatusConflict, "achievement sudah pernah di-unlock oleh member; tidak bisa dihapus")
	}

	err = u.AchievementRepository.Delete(ctx, achievementId)
	if err != nil {
		u.Log.Warnf("Failed delete achievement: %v", err)
		return fiber.ErrInternalServerError
	}

	return nil
}
