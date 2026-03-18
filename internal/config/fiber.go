package config

import (
	"github.com/gofiber/fiber/v3"
	"github.com/spf13/viper"
)

func NewFiber(config *viper.Viper) *fiber.App {
	var app = fiber.New(
		fiber.Config{
			AppName : config.GetString("app.name"),
		},
	)

	return app
	
}

func NewErrorHandler() fiber.ErrorHandler{
	return func(c fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		return c.Status(code).JSON(fiber.Map{
			"error" : err.Error(),
		})
	}
}