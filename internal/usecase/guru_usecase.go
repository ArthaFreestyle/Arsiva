package usecase

import (
	"context"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
)

type GuruUseCase interface {
	Create(ctx context.Context, req *model.GuruCreateRequest, claims *model.Claims) (*model.GuruResponse, error)
	FindById(ctx context.Context, guruId string, claims *model.Claims) (*model.GuruDetailResponse, error)
	FindAll(ctx context.Context, search string, page int, size int) ([]*model.GuruResponse, int, error)
	Update(ctx context.Context, guruId string, req *model.GuruUpdateRequest, claims *model.Claims) (*model.GuruResponse, error)
	Delete(ctx context.Context, guruId string) error
	GetMe(ctx context.Context, claims *model.Claims) (*model.GuruDetailResponse, error)
}

type guruUseCaseImpl struct {
	GuruRepository repository.GuruRepository
	Log            *logrus.Logger
	Validator      *validator.Validate
}

func NewGuruUseCase(guruRepo repository.GuruRepository, log *logrus.Logger, validate *validator.Validate) GuruUseCase {
	return &guruUseCaseImpl{
		GuruRepository: guruRepo,
		Log:            log,
		Validator:      validate,
	}
}

func (u *guruUseCaseImpl) Create(ctx context.Context, req *model.GuruCreateRequest, claims *model.Claims) (*model.GuruResponse, error) {
	if claims.Role == "guru" {
		req.UserId = claims.UserId
	}

	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Invalid request body: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	guru := &entity.Guru{
		UserId:     req.UserId,
		SekolahId:  req.SekolahId,
		NIP:        req.NIP,
		BidangAjar: req.BidangAjar,
	}

	result, err := u.GuruRepository.Create(ctx, guru)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "nip") {
				return nil, fiber.NewError(fiber.StatusConflict, "nip sudah digunakan")
			}
			return nil, fiber.NewError(fiber.StatusConflict, "user_id sudah terdaftar sebagai guru")
		}
		u.Log.Warnf("Failed create guru: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToGuruResponse(result), nil
}

func (u *guruUseCaseImpl) FindById(ctx context.Context, guruId string, claims *model.Claims) (*model.GuruDetailResponse, error) {
	if claims.Role == "guru" {
		claimsGuruId := extractGuruIdFromClaims(claims)
		if claimsGuruId == "" || claimsGuruId != guruId {
			return nil, fiber.ErrForbidden
		}
	}

	guru, err := u.GuruRepository.FindById(ctx, guruId)
	if err != nil {
		u.Log.Warnf("Guru not found: %v", err)
		return nil, fiber.ErrNotFound
	}

	var sekolah *entity.Sekolah
	if guru.SekolahId != "" {
		sekolah, _ = u.GuruRepository.FindSekolahByGuruId(ctx, guruId)
	}

	groups, err := u.GuruRepository.FindGroupsByGuruId(ctx, guruId)
	if err != nil {
		u.Log.Warnf("Failed get groups for guru: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToGuruDetailResponse(guru, sekolah, groups), nil
}

func (u *guruUseCaseImpl) FindAll(ctx context.Context, search string, page int, size int) ([]*model.GuruResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	offset := (page - 1) * size

	gurus, total, err := u.GuruRepository.FindAll(ctx, search, size, offset)
	if err != nil {
		u.Log.Warnf("Failed get all guru: %v", err)
		return nil, 0, fiber.ErrInternalServerError
	}

	return converter.ToGuruResponses(gurus), total, nil
}

func (u *guruUseCaseImpl) Update(ctx context.Context, guruId string, req *model.GuruUpdateRequest, claims *model.Claims) (*model.GuruResponse, error) {
	if claims.Role == "guru" {
		claimsGuruId := extractGuruIdFromClaims(claims)
		if claimsGuruId == "" || claimsGuruId != guruId {
			return nil, fiber.ErrForbidden
		}
	}

	_, err := u.GuruRepository.FindById(ctx, guruId)
	if err != nil {
		u.Log.Warnf("Guru not found: %v", err)
		return nil, fiber.ErrNotFound
	}

	guru := &entity.Guru{
		GuruId:     guruId,
		SekolahId:  req.SekolahId,
		NIP:        req.NIP,
		BidangAjar: req.BidangAjar,
	}

	result, err := u.GuruRepository.Update(ctx, guru)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "nip") {
				return nil, fiber.NewError(fiber.StatusConflict, "nip sudah digunakan")
			}
		}
		u.Log.Warnf("Failed update guru: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToGuruResponse(result), nil
}

func (u *guruUseCaseImpl) Delete(ctx context.Context, guruId string) error {
	_, err := u.GuruRepository.FindById(ctx, guruId)
	if err != nil {
		u.Log.Warnf("Guru not found: %v", err)
		return fiber.ErrNotFound
	}

	err = u.GuruRepository.Delete(ctx, guruId)
	if err != nil {
		u.Log.Warnf("Failed delete guru: %v", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func (u *guruUseCaseImpl) GetMe(ctx context.Context, claims *model.Claims) (*model.GuruDetailResponse, error) {
	guruId := extractGuruIdFromClaims(claims)
	if guruId == "" {
		return nil, fiber.ErrForbidden
	}
	return u.FindById(ctx, guruId, claims)
}

func extractGuruIdFromClaims(claims *model.Claims) string {
	if claims.Details == nil {
		return ""
	}
	detailsMap, ok := claims.Details.(map[string]interface{})
	if !ok {
		return ""
	}
	guruId, _ := detailsMap["guru_id"].(string)
	return guruId
}
