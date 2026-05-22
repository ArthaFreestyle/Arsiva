package usecase

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
	"ArthaFreestyle/Arsiva/internal/utils"
	"context"
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type AuthUseCase interface {
	Login(ctx context.Context, request *model.LoginRequest) (*model.LoginResponse, error)
	RegisterMember(ctx context.Context, request *model.RegisterRequest) (*model.UserResponse, error)
	RegisterGuru(ctx context.Context, request *model.RegisterRequest) (*model.UserResponse, error)
}

type AuthUseCaseImpl struct {
	DB               *pgxpool.Pool
	Log              *logrus.Logger
	Validate         *validator.Validate
	Repo             repository.UserRepository
	GuruRepository   repository.GuruRepository
	MemberRepository repository.MemberRepository
	Secret           []byte
}

func NewAuthUseCase(repo repository.UserRepository, secret []byte, validate *validator.Validate, log *logrus.Logger, DB *pgxpool.Pool, guruRepo repository.GuruRepository, memberRepo repository.MemberRepository) AuthUseCase {
	return &AuthUseCaseImpl{
		Repo:             repo,
		Secret:           secret,
		Validate:         validate,
		DB:               DB,
		Log:              log,
		GuruRepository:   guruRepo,
		MemberRepository: memberRepo,
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

	// Role check AFTER password verification so a wrong expected_role
	// cannot be used to enumerate roles on accounts with bad passwords.
	if request.ExpectedRole != "" && request.ExpectedRole != user.Role {
		c.Log.Warnf("Login role mismatch: expected_role=%s actual=%s", request.ExpectedRole, user.Role)
		return nil, fiber.NewError(fiber.StatusForbidden, "wrong login page for this account")
	}

	// Embed the profile details into the JWT so downstream stateless checks
	// (RequireProfileComplete, member_id extraction) work without DB hits.
	// A missing profile row (pgx.ErrNoRows) is legitimate — the user just hasn't
	// completed onboarding, so leave details nil and let ProfileCompleteMiddleware
	// gate them. Any OTHER error must NOT be swallowed: doing so mints a token
	// with empty Details that permanently 403s every member_id/profile-based
	// endpoint with no way to diagnose it.
	var details any
	switch user.Role {
	case "guru":
		if c.GuruRepository != nil {
			guru, err := c.GuruRepository.FindByUserId(ctx, user.UserId)
			switch {
			case err == nil:
				details = model.GuruDetails{
					GuruId:     guru.GuruId,
					NIP:        guru.NIP,
					BidangAjar: guru.BidangAjar,
					SekolahId:  guru.SekolahId,
				}
			case errors.Is(err, pgx.ErrNoRows):
				c.Log.Infof("Guru profile not yet created for user %s; issuing token without details", user.UserId)
			default:
				c.Log.Errorf("Failed to load guru profile for user %s: %+v", user.UserId, err)
				return nil, fiber.ErrInternalServerError
			}
		}
	case "member":
		if c.MemberRepository != nil {
			member, err := c.MemberRepository.FindByUserId(ctx, user.UserId)
			switch {
			case err == nil:
				details = model.MemberDetails{
					MemberId:  member.MemberId,
					NIS:       member.NIS,
					SekolahId: member.SekolahId,
					Level:     member.Level,
				}
			case errors.Is(err, pgx.ErrNoRows):
				c.Log.Infof("Member profile not yet created for user %s; issuing token without details", user.UserId)
			default:
				c.Log.Errorf("Failed to load member profile for user %s: %+v", user.UserId, err)
				return nil, fiber.ErrInternalServerError
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

func (c *AuthUseCaseImpl) RegisterMember(ctx context.Context, request *model.RegisterRequest) (*model.UserResponse, error) {
	return c.register(ctx, request, "member")
}

func (c *AuthUseCaseImpl) RegisterGuru(ctx context.Context, request *model.RegisterRequest) (*model.UserResponse, error) {
	return c.register(ctx, request, "guru")
}

func (c *AuthUseCaseImpl) register(ctx context.Context, request *model.RegisterRequest, role string) (*model.UserResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("Invalid register request body : %+v", err)
		return nil, fiber.ErrBadRequest
	}

	request.Email = strings.ToLower(request.Email)

	hashedPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		c.Log.Warnf("Failed to hash password : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	user := &entity.User{
		Username:     request.Username,
		Email:        request.Email,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	created, err := c.Repo.CreateUser(ctx, user)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return nil, fiber.NewError(fiber.StatusConflict, "email or username already in use")
		}
		c.Log.Warnf("Failed to create user : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToUserResponse(created), nil
}