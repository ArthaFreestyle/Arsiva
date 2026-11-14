package http

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/delivery/http/middleware"
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/repository"
	"ArthaFreestyle/Arsiva/internal/usecase"
)

type memberListRepo struct {
	repository.MemberRepository
	find func(string, string, int, int) ([]*entity.MemberAccount, int, error)
}

func (r *memberListRepo) FindAll(_ context.Context, search, status string, limit, offset int) ([]*entity.MemberAccount, int, error) {
	return r.find(search, status, limit, offset)
}

func TestMemberListHTTP(t *testing.T) {
	for _, tc := range []struct {
		name, query, role, search, status string
		page, size, offset, code          int
		repoError                         bool
	}{
		{name: "default", role: "super_admin", status: "all", page: 1, size: 10, code: 200},
		{name: "with profile and search", query: "?profile_status=with_profile&search=NIS&page=2&size=1", role: "super_admin", search: "NIS", status: "with_profile", page: 2, size: 1, offset: 1, code: 200},
		{name: "without profile", query: "?profile_status=without_profile", role: "super_admin", status: "without_profile", page: 1, size: 10, code: 200},
		{name: "normalize zero", query: "?page=0&size=0", role: "super_admin", status: "all", page: 1, size: 10, code: 200},
		{name: "normalize negative and large", query: "?page=-3&size=51", role: "super_admin", status: "all", page: 1, size: 10, code: 200},
		{name: "normalize non-numeric", query: "?page=abc&size=abc", role: "super_admin", status: "all", page: 1, size: 10, code: 200},
		{name: "invalid status", query: "?profile_status=complete", role: "super_admin", code: 400},
		{name: "case sensitive status", query: "?profile_status=ALL", role: "super_admin", code: 400},
		{name: "overflow offset", query: "?page=" + strconv.Itoa(math.MaxInt), role: "super_admin", code: 400},
		{name: "member forbidden", role: "member", code: 403},
		{name: "guru forbidden", role: "guru", code: 403},
		{name: "database error", role: "super_admin", status: "all", page: 1, size: 10, code: 500, repoError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			repo := &memberListRepo{find: func(search, status string, limit, offset int) ([]*entity.MemberAccount, int, error) {
				called = true
				if search != tc.search || status != tc.status || limit != tc.size || offset != tc.offset {
					t.Fatalf("repository args: %q %q %d %d", search, status, limit, offset)
				}
				if tc.repoError {
					return nil, 0, errors.New("database unavailable")
				}
				return []*entity.MemberAccount{{UserId: "42", Username: "new_student", Email: "new@example.com"}}, 23, nil
			}}
			uc := usecase.NewMemberUseCase(repo, nil, nil, logrus.New(), nil)
			controller := NewMemberController(uc, logrus.New())
			app := fiber.New()
			// Same per-route role restriction as SetupAuthRoutes.
			app.Get("/v1/member", func(ctx fiber.Ctx) error { ctx.Locals("user", &model.Claims{Role: tc.role}); return ctx.Next() }, middleware.RoleMiddleware("super_admin"), controller.FindAll)
			resp, err := app.Test(httptest.NewRequest("GET", "/v1/member"+tc.query, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.code {
				t.Fatalf("status=%d want %d", resp.StatusCode, tc.code)
			}
			if tc.code == 400 || tc.code == 403 {
				if called {
					t.Fatal("invalid or forbidden request reached repository")
				}
				return
			}
			if !called {
				t.Fatal("repository not called")
			}
			if tc.code != 200 {
				return
			}
			var body struct {
				Data   []map[string]any   `json:"data"`
				Paging model.PageMetaData `json:"paging"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.Paging.Page != tc.page || body.Paging.Size != tc.size || body.Paging.TotalItem != 23 || body.Paging.TotalPage != (23+tc.size-1)/tc.size {
				t.Fatalf("paging=%+v", body.Paging)
			}
			if len(body.Data) != 1 || body.Data[0]["user_id"] != "42" || body.Data[0]["has_profile"] != false {
				t.Fatalf("data=%+v", body.Data)
			}
			for _, key := range []string{"member_id", "sekolah_id", "nis", "total_xp", "level", "foto_profil", "bio", "tanggal_lahir", "jenis_kelamin", "minat"} {
				if value, ok := body.Data[0][key]; !ok || value != nil {
					t.Errorf("%s must be explicit null, got %v (present=%v)", key, value, ok)
				}
			}
		})
	}
}
