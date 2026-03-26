package usecase

import (
	"ArthaFreestyle/Arsiva/internal/entity"
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

type UserUseCase interface {
	GetAllUsers(ctx context.Context) ([]*model.UserResponse, error)
	GetUserById(ctx context.Context,UserId string) (*model.UserResponse, error)
	CreateUser(ctx context.Context,user *model.UserRequest) (*model.UserResponse, error)
	UpdateUser(ctx context.Context,user *model.UserRequest, UserId string) (*model.UserResponse, error)
	DeleteUser(ctx context.Context,UserId string) (error)
}

type UserUseCaseImpl struct {
	UserRepository repository.UserRepository
	Log *logrus.Logger
	Validate *validator.Validate
	DB *pgxpool.Pool
}

func NewUserUseCase(userRepository repository.UserRepository,log *logrus.Logger,db *pgxpool.Pool,val *validator.Validate) UserUseCase {
	return &UserUseCaseImpl{
		UserRepository: userRepository,
		Log: log,
		Validate: val,
		DB: db,
	}
}


func (c *UserUseCaseImpl) GetAllUsers(ctx context.Context) ([]*model.UserResponse, error) {
	users,err := c.UserRepository.GetAllUsers(ctx)
	if err != nil {
		return nil,err
	}

	res := converter.ToUsersResponse(users)

	return res,nil
}

func (c *UserUseCaseImpl) GetUserById(ctx context.Context,UserId string) (*model.UserResponse, error) {
	user,err := c.UserRepository.GetUserById(ctx,UserId)
	if err != nil {
		return nil,err
	}

	res := converter.ToUserResponse(user)

	return res,nil
}

func (c *UserUseCaseImpl) CreateUser(ctx context.Context,user *model.UserRequest) (*model.UserResponse, error) {
	if err := c.Validate.Struct(user); err != nil {
		c.Log.Warnf("Validation error: %v", err)
		return nil,fiber.ErrBadRequest
	}
	
	HashedPassword,err := utils.HashPassword(user.Password)
	if err != nil {
		c.Log.Warnf("Error hashing password: %v", err)
		return nil,fiber.ErrInternalServerError
	}

	NewUser := &entity.User{
		Username: user.Username,
		Email: strings.ToLower(user.Email),
		PasswordHash: HashedPassword,
		Role: user.Role,
	}

	createdUser,err := c.UserRepository.CreateUser(ctx,NewUser)
	if err != nil {
		c.Log.Warnf("Error creating user: %v", err)
		return nil,fiber.ErrInternalServerError
	}

	return converter.ToUserResponse(createdUser),nil
}

func (c *UserUseCaseImpl) UpdateUser(ctx context.Context,user *model.UserRequest, UserId string) (*model.UserResponse, error) {
	if err := c.Validate.Struct(user); err != nil {
		c.Log.Warnf("Validation error: %v", err)
		return nil,fiber.ErrBadRequest
	}
	
	HashedPassword,err := utils.HashPassword(user.Password)
	if err != nil {
		c.Log.Warnf("Error hashing password: %v", err)
		return nil,fiber.ErrInternalServerError
	}

	UpdatedUser := &entity.User{
		UserId: UserId,
		Username: user.Username,
		Email: strings.ToLower(user.Email),
		PasswordHash: HashedPassword,
		Role: user.Role,
	}

	updatedUser,err := c.UserRepository.UpdateUser(ctx,UpdatedUser)
	if err != nil {
		c.Log.Warnf("Error updating user: %v", err)
		return nil,fiber.ErrInternalServerError
	}

	return converter.ToUserResponse(updatedUser),nil
}

func (c *UserUseCaseImpl) DeleteUser(ctx context.Context,UserId string) (error) {
	user,err := c.UserRepository.GetUserById(ctx,UserId)
	if err != nil {
		c.Log.Warnf("Error getting user: %v", err)
		return fiber.ErrInternalServerError
	}

	err = c.UserRepository.DeleteUser(ctx,user)
	if err != nil {
		c.Log.Warnf("Error deleting user: %v", err)
		return fiber.ErrInternalServerError
	}

	return nil
}