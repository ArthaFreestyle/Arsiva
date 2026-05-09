package usecase

import (
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
	"ArthaFreestyle/Arsiva/internal/utils"
	"context"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type AuthUseCase interface {
	Login(ctx context.Context, request *model.LoginRequest) (*model.LoginResponse, error)
}

type AuthUseCaseImpl struct {
	DB             *pgxpool.Pool
	Log            *logrus.Logger
	Validate       *validator.Validate
	Repo           repository.UserRepository
	GuruRepository repository.GuruRepository
	Secret         []byte
}

func NewAuthUseCase(repo repository.UserRepository, secret []byte, validate *validator.Validate, log *logrus.Logger, DB *pgxpool.Pool, guruRepo repository.GuruRepository) AuthUseCase {
	return &AuthUseCaseImpl{
		Repo:           repo,
		Secret:         secret,
		Validate:       validate,
		DB:             DB,
		Log:            log,
		GuruRepository: guruRepo,
	}
}

func (c *AuthUseCaseImpl) Login(ctx context.Context, request *model.LoginRequest) (*model.LoginResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("Invalid request body  : %+v", err)
		return nil, fiber.ErrBadRequest
	}

	request.Email = strings.ToLower(request.Email)

	user, err := c.Repo.FindByEmail(ctx, request.Email)
	if err != nil {
		c.Log.Warnf("Failed find user by email : %+v", err)
		return nil, fiber.ErrUnauthorized
	}

	if !utils.CheckPasswordHash(request.Password, user.PasswordHash) {
		c.Log.Warnf("Invalid password : %+v", err)
		return nil, fiber.ErrUnauthorized
	}

	var details any
	if user.Role == "guru" && c.GuruRepository != nil {
		guru, err := c.GuruRepository.FindByUserId(ctx, user.UserId)
		if err == nil {
			details = model.GuruDetails{
				GuruId:     guru.GuruId,
				NIP:        guru.NIP,
				BidangAjar: guru.BidangAjar,
				SekolahId:  guru.SekolahId,
			}
		}
	}

	access, refresh, err := utils.GenerateToken(user, details, c.Secret)
	if err != nil {
		c.Log.Warnf("Failed generate token : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	AuthResponse := converter.ToUserResponse(user)

	return &model.LoginResponse{
		User:         *AuthResponse,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}