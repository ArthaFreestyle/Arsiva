package usecase

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
)

type MemberSocialLinkUseCase interface {
	Create(ctx context.Context, req *model.MemberSocialLinkCreateRequest, claims *model.Claims) (*model.MemberSocialLinkResponse, error)
	FindAllMine(ctx context.Context, claims *model.Claims) ([]*model.MemberSocialLinkResponse, error)
	FindById(ctx context.Context, socialId string, claims *model.Claims) (*model.MemberSocialLinkResponse, error)
	Update(ctx context.Context, socialId string, req *model.MemberSocialLinkUpdateRequest, claims *model.Claims) (*model.MemberSocialLinkResponse, error)
	Delete(ctx context.Context, socialId string, claims *model.Claims) error
}

type memberSocialLinkUseCaseImpl struct {
	SocialLinkRepository repository.MemberSocialLinkRepository
	Log                  *logrus.Logger
	Validator            *validator.Validate
}

func NewMemberSocialLinkUseCase(socialLinkRepo repository.MemberSocialLinkRepository, log *logrus.Logger, validate *validator.Validate) MemberSocialLinkUseCase {
	return &memberSocialLinkUseCaseImpl{
		SocialLinkRepository: socialLinkRepo,
		Log:                  log,
		Validator:            validate,
	}
}

func (u *memberSocialLinkUseCaseImpl) Create(ctx context.Context, req *model.MemberSocialLinkCreateRequest, claims *model.Claims) (*model.MemberSocialLinkResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Invalid request body: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}

	exists, err := u.SocialLinkRepository.ExistsByMemberAndPlatform(ctx, memberId, req.Platform, "")
	if err != nil {
		u.Log.Warnf("Failed check platform exists: %v", err)
		return nil, fiber.ErrInternalServerError
	}
	if exists {
		return nil, fiber.NewError(fiber.StatusConflict, "platform sudah terdaftar")
	}

	link := &entity.MemberSocialLink{
		MemberId: memberId,
		Platform: entity.Platform(req.Platform),
		URL:      req.URL,
	}

	result, err := u.SocialLinkRepository.Create(ctx, link)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case "23505":
				return nil, fiber.NewError(fiber.StatusConflict, "platform sudah terdaftar")
			case "23503":
				u.Log.Errorf("FK violation creating social link: %v", err)
				return nil, fiber.ErrInternalServerError
			}
		}
		u.Log.Warnf("Failed create social link: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToMemberSocialLinkResponse(result), nil
}

func (u *memberSocialLinkUseCaseImpl) FindAllMine(ctx context.Context, claims *model.Claims) ([]*model.MemberSocialLinkResponse, error) {
	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}

	links, err := u.SocialLinkRepository.FindAllByMemberId(ctx, memberId)
	if err != nil {
		u.Log.Warnf("Failed get social links: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToMemberSocialLinkResponses(links), nil
}

func (u *memberSocialLinkUseCaseImpl) FindById(ctx context.Context, socialId string, claims *model.Claims) (*model.MemberSocialLinkResponse, error) {
	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}

	link, err := u.SocialLinkRepository.FindById(ctx, socialId)
	if err != nil {
		return nil, fiber.ErrNotFound
	}

	if link.MemberId != memberId {
		return nil, fiber.ErrForbidden
	}

	return converter.ToMemberSocialLinkResponse(link), nil
}

func (u *memberSocialLinkUseCaseImpl) Update(ctx context.Context, socialId string, req *model.MemberSocialLinkUpdateRequest, claims *model.Claims) (*model.MemberSocialLinkResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Invalid request body: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}

	link, err := u.SocialLinkRepository.FindById(ctx, socialId)
	if err != nil {
		return nil, fiber.ErrNotFound
	}

	if link.MemberId != memberId {
		return nil, fiber.ErrForbidden
	}

	exists, err := u.SocialLinkRepository.ExistsByMemberAndPlatform(ctx, memberId, req.Platform, socialId)
	if err != nil {
		u.Log.Warnf("Failed check platform exists: %v", err)
		return nil, fiber.ErrInternalServerError
	}
	if exists {
		return nil, fiber.NewError(fiber.StatusConflict, "platform sudah terdaftar")
	}

	link.Platform = entity.Platform(req.Platform)
	link.URL = req.URL

	result, err := u.SocialLinkRepository.Update(ctx, link)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, fiber.NewError(fiber.StatusConflict, "platform sudah terdaftar")
		}
		u.Log.Warnf("Failed update social link: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToMemberSocialLinkResponse(result), nil
}

func (u *memberSocialLinkUseCaseImpl) Delete(ctx context.Context, socialId string, claims *model.Claims) error {
	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return fiber.ErrForbidden
	}

	link, err := u.SocialLinkRepository.FindById(ctx, socialId)
	if err != nil {
		return fiber.ErrNotFound
	}

	if link.MemberId != memberId {
		return fiber.ErrForbidden
	}

	if err = u.SocialLinkRepository.Delete(ctx, socialId); err != nil {
		u.Log.Warnf("Failed delete social link: %v", err)
		return fiber.ErrInternalServerError
	}

	return nil
}
