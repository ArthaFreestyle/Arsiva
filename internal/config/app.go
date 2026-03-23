package config

import (
	"ArthaFreestyle/Arsiva/internal/repository"
	"ArthaFreestyle/Arsiva/internal/usecase"
	"ArthaFreestyle/Arsiva/internal/delivery/http"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/gofiber/fiber/v3"
	"ArthaFreestyle/Arsiva/delivery/http/route"
)

type BootstrapConfig struct{
	DB				*pgxpool.Pool
	App				*fiber.App
	Log				*logrus.Logger
	Validate		*validator.Validate
	Secret			[]byte
	Config			*viper.Viper
}

func Bootstrap(cfg BootstrapConfig) {
	userRepo := repository.NewUserRepository(cfg.DB,cfg.Log)

	AuthUseCase := usecase.NewAuthUseCase(userRepo,cfg.Secret,cfg.Validate,cfg.Log,cfg.DB)
	UserUseCase := usecase.NewUserUseCase(userRepo,cfg.Log,cfg.DB,cfg.Validate)

	AuthController := http.NewAuthController(cfg.Log,AuthUseCase)
	UserController := http.NewUserController(UserUseCase,cfg.Log)

	routeConfig := route.RouteConfig{
		App: cfg.App,
		AuthController : AuthController,
		UserController : UserController,
	}

	routeConfig.SetupRoutes()
	
}