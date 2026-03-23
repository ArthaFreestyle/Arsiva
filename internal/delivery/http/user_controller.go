package http

import (
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type UserController interface {
	GetAllUsers(ctx fiber.Ctx) error
	GetUserById(ctx fiber.Ctx) error
	CreateUser(ctx fiber.Ctx) error
	UpdateUser(ctx fiber.Ctx) error
	DeleteUser(ctx fiber.Ctx) error
}

type UserControllerImpl struct {
	UserUseCase usecase.UserUseCase
	Log *logrus.Logger
}

func NewUserController(userUseCase usecase.UserUseCase,log *logrus.Logger) UserController {
	return &UserControllerImpl{
		UserUseCase: userUseCase,
		Log: log,
	}
}

func (u *UserControllerImpl) GetAllUsers(ctx fiber.Ctx) error {
	users,err := u.UserUseCase.GetAllUsers(ctx)
	if err != nil {
		return err
	}

	res := model.WebResponse[[]*model.UserResponse]{
		Data: users,
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (u *UserControllerImpl) GetUserById(ctx fiber.Ctx) error {
	userId := ctx.Params("id")
	user,err := u.UserUseCase.GetUserById(ctx,userId)
	if err != nil {
		u.Log.Warnf("Failed get user by id : %+v", err)
		return err
	}

	res := model.WebResponse[*model.UserResponse]{
		Data: user,
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (u *UserControllerImpl) CreateUser(ctx fiber.Ctx) error {
	var user model.UserRequest
	if err := ctx.Bind().Body(&user); err != nil {
		u.Log.Warnf("Invalid request body  : %+v", err)
		return err
	}

	createdUser,err := u.UserUseCase.CreateUser(ctx,&user)
	if err != nil {
		u.Log.Warnf("Failed create user : %+v", err)
		return err
	}

	res := model.WebResponse[*model.UserResponse]{
		Data: createdUser,
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (u *UserControllerImpl) UpdateUser(ctx fiber.Ctx) error {
	userId := ctx.Params("id")
	var user model.UserRequest
	if err := ctx.Bind().Body(&user); err != nil {
		return err
	}

	updatedUser,err := u.UserUseCase.UpdateUser(ctx,&user,userId)
	if err != nil {
		return err
	}

	res := model.WebResponse[*model.UserResponse]{
		Data: updatedUser,
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (u *UserControllerImpl) DeleteUser(ctx fiber.Ctx) error {
	userId := ctx.Params("id")
	err := u.UserUseCase.DeleteUser(ctx,userId)
	if err != nil {
		return err
	}

	res := model.WebResponse[string]{
		Data: "User deleted successfully",
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}