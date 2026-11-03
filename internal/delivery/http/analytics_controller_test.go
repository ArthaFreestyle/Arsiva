package http

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/delivery/http/middleware"
	"ArthaFreestyle/Arsiva/internal/model"
)

type stubAnalyticsUseCase struct{}

func (stubAnalyticsUseCase) GetUserSummary(context.Context) (*model.UserSummaryResponse, error) {
	return &model.UserSummaryResponse{}, nil
}
func (stubAnalyticsUseCase) GetContentSummary(context.Context) (*model.ContentSummaryResponse, error) {
	return &model.ContentSummaryResponse{}, nil
}
func (stubAnalyticsUseCase) GetUserGrowth(context.Context, *model.GrowthRequest) (*model.UserGrowthResponse, error) {
	return &model.UserGrowthResponse{}, nil
}
func (stubAnalyticsUseCase) GetContentGrowth(context.Context, *model.GrowthRequest) (*model.ContentGrowthResponse, error) {
	return &model.ContentGrowthResponse{}, nil
}

func TestNewAnalyticsController(t *testing.T) {
	if NewAnalyticsController(nil, nil) == nil {
		t.Fatal("expected controller instance")
	}
}

// Mirrors the route registration: super_admin only, middleware passed inline.
func TestAnalyticsController_SuperAdminOnly(t *testing.T) {
	c := NewAnalyticsController(stubAnalyticsUseCase{}, logrus.New())
	routes := map[string]fiber.Handler{
		"/v1/analytics/users/summary":   c.GetUserSummary,
		"/v1/analytics/users/growth":    c.GetUserGrowth,
		"/v1/analytics/content/summary": c.GetContentSummary,
		"/v1/analytics/content/growth":  c.GetContentGrowth,
	}

	for _, tc := range []struct {
		role string
		want int
	}{
		{"super_admin", fiber.StatusOK},
		{"guru", fiber.StatusForbidden},
		{"member", fiber.StatusForbidden},
	} {
		app := fiber.New()
		setClaims := func(ctx fiber.Ctx) error {
			ctx.Locals("user", &model.Claims{UserId: "1", Role: tc.role})
			return ctx.Next()
		}
		superadminOnly := middleware.RoleMiddleware("super_admin")
		for path, h := range routes {
			app.Get(path, setClaims, superadminOnly, h)
		}

		for path := range routes {
			resp, err := app.Test(httptest.NewRequest("GET", path, nil))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.want {
				t.Errorf("%s as %s: status %d, want %d", path, tc.role, resp.StatusCode, tc.want)
			}
		}
	}
}
