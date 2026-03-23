package route

import (
	"github.com/gofiber/fiber/v3"
	"ArthaFreestyle/Arsiva/internal/delivery/http"
)

type RouteConfig struct {
	App *fiber.App
	AuthController http.AuthController
	UserController http.UserController
}

func (c *RouteConfig) SetupRoutes() {
	c.SetupGuestRoutes()
	c.SetupAuthRoutes()
}

func (c *RouteConfig) SetupGuestRoutes() {
	c.App.Post("/api/v1/login",c.AuthController.Login)
}

func (c *RouteConfig) SetupAuthRoutes() {
	c.App.Get("/api/v1/users",c.UserController.GetAllUsers)
	c.App.Get("/api/v1/users/:id",c.UserController.GetUserById)
	c.App.Post("/api/v1/users",c.UserController.CreateUser)
	c.App.Put("/api/v1/users/:id",c.UserController.UpdateUser)
	c.App.Delete("/api/v1/users/:id",c.UserController.DeleteUser)
}