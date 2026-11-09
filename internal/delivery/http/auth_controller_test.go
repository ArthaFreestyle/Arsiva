package http

import (
	"ArthaFreestyle/Arsiva/internal/usecase"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

func TestNewAuthController(t *testing.T) {
	controller := NewAuthController(nil, nil)
	if controller == nil {
		t.Fatal("expected controller instance")
	}
}

func newRefreshTestApp() *fiber.App {
	log := logrus.New()
	log.SetOutput(io.Discard)
	// Real usecase with no repo/Redis: validation fails before either is touched.
	uc := usecase.NewAuthUseCase(nil, []byte("secret"), validator.New(), log, nil, nil, nil, nil, nil, 5*time.Minute, 15*time.Minute, 5, time.Minute, "")
	ctrl := NewAuthController(log, uc)
	app := fiber.New()
	app.Post("/v1/refresh", ctrl.Refresh)
	app.Post("/v1/logout", ctrl.Logout)
	return app
}

func TestAuthController_RefreshAndLogout_BadBody_Returns400(t *testing.T) {
	app := newRefreshTestApp()
	cases := []struct{ name, path, body string }{
		{"refresh malformed json", "/v1/refresh", `{not json`},
		{"refresh missing token", "/v1/refresh", `{}`},
		{"logout malformed json", "/v1/logout", `{not json`},
		{"logout missing token", "/v1/logout", `{"refresh_token":""}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			if resp.StatusCode != fiber.StatusBadRequest {
				t.Errorf("expected 400, got %d", resp.StatusCode)
			}
		})
	}
}
