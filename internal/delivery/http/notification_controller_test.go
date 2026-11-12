package http

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"ArthaFreestyle/Arsiva/internal/model"
)

func TestNotificationAuthorization(t *testing.T) {
	for _, tc := range []struct {
		name   string
		claims *model.Claims
		status int
	}{{"anonymous", nil, 401}, {"teacher", &model.Claims{Role: "guru"}, 403}, {"admin", &model.Claims{Role: "super_admin"}, 403}, {"missing-profile", &model.Claims{Role: "member"}, 403}} {
		t.Run(tc.name, func(t *testing.T) {
			for _, method := range []string{"GET", "PATCH"} {
				app := fiber.New()
				c := &NotificationController{}
				app.Use(func(ctx fiber.Ctx) error {
					if tc.claims != nil {
						ctx.Locals("user", tc.claims)
					}
					return ctx.Next()
				})
				app.Get("/notifications", c.List)
				app.Patch("/notifications", c.Read)
				resp, err := app.Test(httptest.NewRequest(method, "/notifications", nil))
				if err != nil {
					t.Fatal(err)
				}
				resp.Body.Close()
				if resp.StatusCode != tc.status {
					t.Fatalf("%s got %d want %d", method, resp.StatusCode, tc.status)
				}
			}
		})
	}
}
