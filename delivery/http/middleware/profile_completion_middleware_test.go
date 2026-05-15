package middleware

import (
	"ArthaFreestyle/Arsiva/internal/model"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func runProfileMiddleware(claims *model.Claims) int {
	app := fiber.New()
	mw := RequireProfileComplete()

	app.Get("/test", func(ctx fiber.Ctx) error {
		ctx.Locals("user", claims)
		return ctx.Next()
	}, mw, func(ctx fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)
	return resp.StatusCode
}

func TestRequireProfileComplete_SuperAdmin_AlwaysPasses(t *testing.T) {
	status := runProfileMiddleware(&model.Claims{UserId: "1", Role: "super_admin"})
	if status != fiber.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestRequireProfileComplete_Guru_WithDetails_Passes(t *testing.T) {
	claims := &model.Claims{
		UserId:  "2",
		Role:    "guru",
		Details: model.GuruDetails{GuruId: "10"},
	}
	status := runProfileMiddleware(claims)
	if status != fiber.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestRequireProfileComplete_Guru_NoDetails_Returns409(t *testing.T) {
	status := runProfileMiddleware(&model.Claims{UserId: "3", Role: "guru"})
	if status != fiber.StatusConflict {
		t.Errorf("expected 409, got %d", status)
	}
}

func TestRequireProfileComplete_Member_WithDetails_Passes(t *testing.T) {
	claims := &model.Claims{
		UserId:  "4",
		Role:    "member",
		Details: model.MemberDetails{MemberId: "20"},
	}
	status := runProfileMiddleware(claims)
	if status != fiber.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestRequireProfileComplete_Member_NoDetails_Returns409(t *testing.T) {
	status := runProfileMiddleware(&model.Claims{UserId: "5", Role: "member"})
	if status != fiber.StatusConflict {
		t.Errorf("expected 409, got %d", status)
	}
}

func TestRequireProfileComplete_MissingClaims_Returns401(t *testing.T) {
	app := fiber.New()
	mw := RequireProfileComplete()

	app.Get("/test", mw, func(ctx fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}
