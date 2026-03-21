package route

import (
	"github.com/gofiber/fiber/v3"
	"ArthaFreestyle/Arsiva/internal/delivery/http"
)

type RouteConfig struct {
	App *fiber.App
	AuthController http.AuthController
}

func (c *RouteConfig) SetupRoutes() {
	c.SetupGuestRoutes()
}

func (c *RouteConfig) SetupGuestRoutes() {
	c.App.Post("/api/v1/login",c.AuthController.Login)
}