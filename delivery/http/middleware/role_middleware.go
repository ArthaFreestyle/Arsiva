package middleware

import (
	"ArthaFreestyle/Arsiva/internal/model"

	"github.com/gofiber/fiber/v3"
)

// RoleMiddleware returns a Fiber handler that restricts access to the given roles.
// The allowed role values are: "member", "guru", "superadmin".
// This middleware must be used AFTER AuthMiddleware so that c.Locals("user") is set.
func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		claims, ok := ctx.Locals("user").(*model.Claims)
		if !ok || claims == nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(model.WebResponse[any]{
				Errors: "unauthorized",
			})
		}

		for _, role := range allowedRoles {
			if claims.Role == role {
				return ctx.Next()
			}
		}


		return ctx.Status(fiber.StatusForbidden).JSON(model.WebResponse[any]{
			Errors: "forbidden: insufficient role",
		})
	}
}
