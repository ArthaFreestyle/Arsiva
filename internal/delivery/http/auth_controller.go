package http

import (
	"ArthaFreestyle/Arsiva/internal/usecase"
	"ArthaFreestyle/Arsiva/internal/model"
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type AuthController interface {
	Login(ctx fiber.Ctx)  error
}

type AuthControllerImpl struct {
	Log	*logrus.Logger
	UseCase	usecase.AuthUseCase
}


func NewAuthController(log *logrus.Logger,usecase usecase.AuthUseCase) AuthController {
	return &AuthControllerImpl{
		Log: log,
		UseCase: usecase,
	}
}

func (c *AuthControllerImpl) Login(ctx fiber.Ctx)  error {
	var request model.LoginRequest
	if err := ctx.Bind().Body(&request); err != nil {
		c.Log.Warnf("Invalid request body  : %+v", err)
		return fiber.ErrBadRequest
	}

	result,err := c.UseCase.Login(ctx,&request)
	if err != nil {
		c.Log.Warnf("Failed login : %+v", err)
		return err
	}

	return ctx.JSON(model.WebResponse[model.LoginResponse]{
		Status: "success",
		Data: *result,
	})
}