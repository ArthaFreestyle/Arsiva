package middleware

import (
	"ArthaFreestyle/Arsiva/internal/model"

	"github.com/gofiber/fiber/v3"
)

type ProfileIncompleteResponse struct {
	Errors     string `json:"errors"`
	Code       string `json:"code"`
	NextAction string `json:"next_action"`
}

// RequireProfileComplete blocks guru/member users whose JWT has no Details,
// meaning they registered but never completed their profile. It reads only
// the JWT claims — no database queries — so JWT stays stateless.
// Must be placed after AuthMiddleware so ctx.Locals("user") is populated.
func RequireProfileComplete() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		claims, ok := ctx.Locals("user").(*model.Claims)
		if !ok || claims == nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(model.WebResponse[any]{
				Errors: "unauthorized",
			})
		}

		switch claims.Role {
		case "super_admin":
			return ctx.Next()

		case "guru":
			if claims.Details != nil {
				return ctx.Next()
			}
			return ctx.Status(fiber.StatusConflict).JSON(ProfileIncompleteResponse{
				Errors:     "profile incomplete",
				Code:       "PROFILE_INCOMPLETE",
				NextAction: "POST /v1/guru",
			})

		case "member":
			if claims.Details != nil {
				return ctx.Next()
			}
			return ctx.Status(fiber.StatusConflict).JSON(ProfileIncompleteResponse{
				Errors:     "profile incomplete",
				Code:       "PROFILE_INCOMPLETE",
				NextAction: "POST /v1/member",
			})

		default:
			return ctx.Next()
		}
	}
}
