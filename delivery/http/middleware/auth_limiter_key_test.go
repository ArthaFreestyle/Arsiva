package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func runAuthLimiterKey(path, body string) string {
	app := fiber.New()
	var captured string

	app.Post(path, func(ctx fiber.Ctx) error {
		captured = AuthLimiterKey(ctx)
		return ctx.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	app.Test(req)
	return captured
}

func TestAuthLimiterKey_SameEmailDifferentRoutes_ProducesDifferentKeys(t *testing.T) {
	loginKey := runAuthLimiterKey("/v1/login", `{"email":"student@school.id","password":"x"}`)
	forgotKey := runAuthLimiterKey("/v1/forgot-password", `{"email":"student@school.id"}`)

	if loginKey == forgotKey {
		t.Errorf("expected different keys per route for the same account, got identical key %q", loginKey)
	}
}

func TestAuthLimiterKey_DifferentEmailsSameIP_ProducesDifferentKeys(t *testing.T) {
	keyA := runAuthLimiterKey("/v1/login", `{"email":"siswa-a@school.id","password":"x"}`)
	keyB := runAuthLimiterKey("/v1/login", `{"email":"siswa-b@school.id","password":"x"}`)

	if keyA == keyB {
		t.Errorf("expected different keys for different accounts (even behind shared wifi/IP), got identical key %q", keyA)
	}
}

func TestAuthLimiterKey_MalformedBody_FallsBackToIPAndPath(t *testing.T) {
	key := runAuthLimiterKey("/v1/login", `not json`)

	if !strings.HasSuffix(key, "|/v1/login") {
		t.Errorf("expected fallback key to end with route path, got %q", key)
	}
	if strings.Contains(key, "@") {
		t.Errorf("expected fallback key to not contain an email, got %q", key)
	}
}
