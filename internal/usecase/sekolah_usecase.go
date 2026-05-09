package usecase

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
)

type SekolahUseCase interface {
	Create(ctx context.Context, req *model.SekolahCreateRequest) (*model.SekolahResponse, error)
	FindById(ctx context.Context, sekolahId string) (*model.SekolahDetailResponse, error)
	FindAll(ctx context.Context, search string, page int, size int) ([]*model.SekolahResponse, int, error)
	Update(ctx context.Context, sekolahId string, req *model.SekolahUpdateRequest) (*model.SekolahResponse, error)
	Delete(ctx context.Context, sekolahId string) error
}

type sekolahUseCaseImpl struct {
	SekolahRepository repository.SekolahRepository
	Log               *logrus.Logger
	Validator         *validator.Validate
}

func NewSekolahUseCase(sekolahRepo repository.SekolahRepository, log *logrus.Logger, validate *validator.Validate) SekolahUseCase {
	return &sekolahUseCaseImpl{
		SekolahRepository: sekolahRepo,
		Log:               log,
		Validator:         validate,
	}
}

func (u *sekolahUseCaseImpl) Create(ctx context.Context, req *model.SekolahCreateRequest) (*model.SekolahResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Invalid request body: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	sekolah := &entity.Sekolah{
		NamaSekolah:   req.NamaSekolah,
		AlamatSekolah: req.AlamatSekolah,
	}

	result, err := u.SekolahRepository.Create(ctx, sekolah)
	if err != nil {
		u.Log.Warnf("Failed create sekolah: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToSekolahResponse(result), nil
}

func (u *sekolahUseCaseImpl) FindById(ctx context.Context, sekolahId string) (*model.SekolahDetailResponse, error) {
	sekolah, err := u.SekolahRepository.FindById(ctx, sekolahId)
	if err != nil {
		u.Log.Warnf("Sekolah not found: %v", err)
		return nil, fiber.ErrNotFound
	}

	gurus, err := u.SekolahRepository.FindGurusBySekolahId(ctx, sekolahId)
	if err != nil {
		u.Log.Warnf("Failed get gurus for sekolah: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToSekolahDetailResponse(sekolah, gurus), nil
}

func (u *sekolahUseCaseImpl) FindAll(ctx context.Context, search string, page int, size int) ([]*model.SekolahResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	offset := (page - 1) * size

	sekolahs, total, err := u.SekolahRepository.FindAll(ctx, search, size, offset)
	if err != nil {
		u.Log.Warnf("Failed get all sekolah: %v", err)
		return nil, 0, fiber.ErrInternalServerError
	}

	return converter.ToSekolahResponses(sekolahs), total, nil
}

func (u *sekolahUseCaseImpl) Update(ctx context.Context, sekolahId string, req *model.SekolahUpdateRequest) (*model.SekolahResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Invalid request body: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	_, err := u.SekolahRepository.FindById(ctx, sekolahId)
	if err != nil {
		u.Log.Warnf("Sekolah not found: %v", err)
		return nil, fiber.ErrNotFound
	}

	sekolah := &entity.Sekolah{
		SekolahId:     sekolahId,
		NamaSekolah:   req.NamaSekolah,
		AlamatSekolah: req.AlamatSekolah,
	}

	result, err := u.SekolahRepository.Update(ctx, sekolah)
	if err != nil {
		u.Log.Warnf("Failed update sekolah: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToSekolahResponse(result), nil
}

func (u *sekolahUseCaseImpl) Delete(ctx context.Context, sekolahId string) error {
	_, err := u.SekolahRepository.FindById(ctx, sekolahId)
	if err != nil {
		u.Log.Warnf("Sekolah not found: %v", err)
		return fiber.ErrNotFound
	}

	count, err := u.SekolahRepository.CountGurusBySekolahId(ctx, sekolahId)
	if err != nil {
		u.Log.Warnf("Failed count gurus for sekolah: %v", err)
		return fiber.ErrInternalServerError
	}
	if count > 0 {
		return fiber.NewError(fiber.StatusConflict, "cannot delete sekolah: guru records still reference it")
	}

	err = u.SekolahRepository.Delete(ctx, sekolahId)
	if err != nil {
		u.Log.Warnf("Failed delete sekolah: %v", err)
		return fiber.ErrInternalServerError
	}

	return nil
}
