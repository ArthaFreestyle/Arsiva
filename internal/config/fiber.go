package config

import (
	"ArthaFreestyle/Arsiva/internal/model"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/contrib/v3/swaggerui"
	"github.com/spf13/viper"
)

func NewFiber(config *viper.Viper) *fiber.App {
	var app = fiber.New(
		fiber.Config{
			AppName : config.GetString("app.name"),
			ErrorHandler: NewErrorHandler(),
			
		},
	)

	app.Get("/docs/*", static.New("./docs"))

	cfg := swaggerui.Config{
		BasePath: "/",                   
		FilePath: "/docs/openapi.yaml", 
		Path:     "/",            
		Title:    "Dokumentasi API",     
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins: config.GetStringSlice("app.allowance"),
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
	}))

	app.Use(swaggerui.New(cfg))

	return app
	
}

func NewErrorHandler() fiber.ErrorHandler{
	return func(c fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		res := model.WebResponse[string]{
			Errors: err.Error(),
		}

		return c.Status(code).JSON(res)
	}
}