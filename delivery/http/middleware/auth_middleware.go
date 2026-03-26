package middleware

import (
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/utils"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

func NewAuthMiddleware(jwtSecret []byte, log *logrus.Logger) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			log.Warn("Missing Authorization header")
			return ctx.Status(fiber.StatusUnauthorized).JSON(model.WebResponse[any]{
				Errors: "missing authorization header",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			log.Warn("Invalid Authorization header format")
			return ctx.Status(fiber.StatusUnauthorized).JSON(model.WebResponse[any]{
				Errors: "invalid authorization header format",
			})
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			log.Warnf("Invalid or expired token: %+v", err)
			return ctx.Status(fiber.StatusUnauthorized).JSON(model.WebResponse[any]{
				Errors: "invalid or expired token",
			})
		}

		// Store claims and userId in Locals for downstream handlers
		ctx.Locals("user", claims)
		ctx.Locals("userId", claims.UserId)

		return ctx.Next()
	}
}
