package http

import (
	"ArthaFreestyle/Arsiva/internal/usecase"
	"ArthaFreestyle/Arsiva/internal/model"
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type AuthController interface {
	Login(ctx fiber.Ctx) error
	RegisterMember(ctx fiber.Ctx) error
	RegisterGuru(ctx fiber.Ctx) error
	VerifyEmail(ctx fiber.Ctx) error
	ResendOTP(ctx fiber.Ctx) error
	ForgotPassword(ctx fiber.Ctx) error
	ResetPassword(ctx fiber.Ctx) error
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

func (c *AuthControllerImpl) Login(ctx fiber.Ctx) error {
	var request model.LoginRequest
	if err := ctx.Bind().Body(&request); err != nil {
		c.Log.Warnf("Invalid request body  : %+v", err)
		return fiber.ErrBadRequest
	}

	result, err := c.UseCase.Login(ctx, &request)
	if err != nil {
		c.Log.Warnf("Failed login : %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[model.LoginResponse]{
		Data: *result,
	})
}

func (c *AuthControllerImpl) RegisterMember(ctx fiber.Ctx) error {
	var request model.RegisterRequest
	if err := ctx.Bind().Body(&request); err != nil {
		c.Log.Warnf("Invalid request body  : %+v", err)
		return fiber.ErrBadRequest
	}

	result, err := c.UseCase.RegisterMember(ctx, &request)
	if err != nil {
		c.Log.Warnf("Failed register member : %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[model.UserResponse]{
		Data: *result,
	})
}

func (c *AuthControllerImpl) RegisterGuru(ctx fiber.Ctx) error {
	var request model.RegisterRequest
	if err := ctx.Bind().Body(&request); err != nil {
		c.Log.Warnf("Invalid request body  : %+v", err)
		return fiber.ErrBadRequest
	}

	result, err := c.UseCase.RegisterGuru(ctx, &request)
	if err != nil {
		c.Log.Warnf("Failed register guru : %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[model.UserResponse]{
		Data: *result,
	})
}

func (c *AuthControllerImpl) VerifyEmail(ctx fiber.Ctx) error {
	var request model.VerifyEmailRequest
	if err := ctx.Bind().Body(&request); err != nil {
		c.Log.Warnf("Invalid request body  : %+v", err)
		return fiber.ErrBadRequest
	}

	if err := c.UseCase.VerifyEmail(ctx, &request); err != nil {
		c.Log.Warnf("Failed verify email : %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[model.MessageResponse]{
		Data: model.MessageResponse{Message: "email berhasil diverifikasi, silakan login"},
	})
}

func (c *AuthControllerImpl) ResendOTP(ctx fiber.Ctx) error {
	var request model.ResendOTPRequest
	if err := ctx.Bind().Body(&request); err != nil {
		c.Log.Warnf("Invalid request body  : %+v", err)
		return fiber.ErrBadRequest
	}

	if err := c.UseCase.ResendVerificationOTP(ctx, &request); err != nil {
		c.Log.Warnf("Failed resend OTP : %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[model.MessageResponse]{
		Data: model.MessageResponse{Message: "jika email terdaftar dan belum terverifikasi, kode baru telah dikirim"},
	})
}

func (c *AuthControllerImpl) ForgotPassword(ctx fiber.Ctx) error {
	var request model.ForgotPasswordRequest
	if err := ctx.Bind().Body(&request); err != nil {
		c.Log.Warnf("Invalid request body  : %+v", err)
		return fiber.ErrBadRequest
	}

	if err := c.UseCase.ForgotPassword(ctx, &request); err != nil {
		c.Log.Warnf("Failed forgot password : %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[model.MessageResponse]{
		Data: model.MessageResponse{Message: "jika email terdaftar, kode reset password telah dikirim"},
	})
}

func (c *AuthControllerImpl) ResetPassword(ctx fiber.Ctx) error {
	var request model.ResetPasswordRequest
	if err := ctx.Bind().Body(&request); err != nil {
		c.Log.Warnf("Invalid request body  : %+v", err)
		return fiber.ErrBadRequest
	}

	if err := c.UseCase.ResetPassword(ctx, &request); err != nil {
		c.Log.Warnf("Failed reset password : %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[model.MessageResponse]{
		Data: model.MessageResponse{Message: "password berhasil diperbarui, silakan login"},
	})
}