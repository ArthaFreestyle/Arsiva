package middleware

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/utils"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

var mwSecret = []byte("mw-secret")

func runAuthMiddleware(t *testing.T, authHeader string) int {
	t.Helper()
	log := logrus.New()
	log.SetOutput(io.Discard)

	app := fiber.New()
	app.Get("/v1/protected", NewAuthMiddleware(mwSecret, log), func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/v1/protected", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	return resp.StatusCode
}

func TestAuthMiddleware_AccessToken_Allowed(t *testing.T) {
	access, _ := utils.GenerateAccessToken(&entity.User{UserId: "1", Role: "member"}, nil, mwSecret)
	if code := runAuthMiddleware(t, "Bearer "+access); code != fiber.StatusOK {
		t.Errorf("expected 200 for access token, got %d", code)
	}
}

func TestAuthMiddleware_RefreshToken_Rejected(t *testing.T) {
	refresh, _, _ := utils.GenerateRefreshToken("1", mwSecret)
	if code := runAuthMiddleware(t, "Bearer "+refresh); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401 when a refresh token is used as Bearer, got %d", code)
	}
}

func TestAuthMiddleware_MissingHeader_Rejected(t *testing.T) {
	if code := runAuthMiddleware(t, ""); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401 without Authorization header, got %d", code)
	}
}
