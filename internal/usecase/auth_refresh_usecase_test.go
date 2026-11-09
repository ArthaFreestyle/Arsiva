package usecase

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/utils"
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

// refreshFixture wires an AuthUseCase against miniredis with a user whose state
// (active/verified/role) each test can tweak through the returned pointer.
type refreshFixture struct {
	uc         AuthUseCase
	mr         *miniredis.Miniredis
	user       *entity.User
	memberRepo *mockMemberRepo
	userRepo   *mockUserRepo
}

func newRefreshFixture(t *testing.T) *refreshFixture {
	t.Helper()
	hash, _ := utils.HashPassword("Rahasia123!")
	f := &refreshFixture{
		user: &entity.User{
			UserId: "1", Username: "siswa", Email: "siswa@example.com", Role: "member",
			PasswordHash: hash, IsActive: true, IsVerified: true,
		},
		memberRepo: &mockMemberRepo{},
	}
	f.userRepo = &mockUserRepo{
		findByEmailFn: func(ctx context.Context, email string) (*entity.User, error) {
			u := *f.user
			return &u, nil
		},
		getUserByIdFn: func(ctx context.Context, userId string) (*entity.User, error) {
			if userId != f.user.UserId {
				return nil, pgx.ErrNoRows
			}
			u := *f.user
			u.PasswordHash = "" // GetUserById does not select password_hash
			return &u, nil
		},
	}
	rdb, mr := newTestRedis(t)
	f.mr = mr
	f.uc = NewAuthUseCase(f.userRepo, []byte("secret"), validator.New(), discardLogger(), nil, nil, f.memberRepo, rdb, nil, 5*time.Minute, 15*time.Minute, 5, time.Minute, "https://arsiva.id/reset-password")
	return f
}

func (f *refreshFixture) login(t *testing.T) *model.LoginResponse {
	t.Helper()
	resp, err := f.uc.Login(context.Background(), &model.LoginRequest{Email: f.user.Email, Password: "Rahasia123!"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	return resp
}

func (f *refreshFixture) refresh(token string) (*model.LoginResponse, error) {
	return f.uc.Refresh(context.Background(), &model.RefreshTokenRequest{RefreshToken: token})
}

func assertFiberCode(t *testing.T, err error, code int) {
	t.Helper()
	var fe *fiber.Error
	if !errors.As(err, &fe) || fe.Code != code {
		t.Fatalf("expected fiber error %d, got %v", code, err)
	}
}

func jtiOf(t *testing.T, token string) string {
	t.Helper()
	c, err := utils.ValidateRefreshToken(token, []byte("secret"))
	if err != nil {
		t.Fatalf("parse refresh token: %v", err)
	}
	return c.ID
}

func TestLogin_StoresRefreshJti(t *testing.T) {
	f := newRefreshFixture(t)
	resp := f.login(t)

	if !f.mr.Exists(refreshTokenKey(jtiOf(t, resp.RefreshToken))) {
		t.Error("expected refresh jti to be stored in Redis after login")
	}
	if resp.ExpiresIn != int(utils.AccessTokenTTL.Seconds()) {
		t.Errorf("expected expires_in %d, got %d", int(utils.AccessTokenTTL.Seconds()), resp.ExpiresIn)
	}
}

func TestLogin_RedisDown_Returns500(t *testing.T) {
	f := newRefreshFixture(t)
	f.mr.Close()
	_, err := f.uc.Login(context.Background(), &model.LoginRequest{Email: f.user.Email, Password: "Rahasia123!"})
	assertFiberCode(t, err, http.StatusInternalServerError)
}

func TestRefresh_HappyPath_RotatesToken(t *testing.T) {
	f := newRefreshFixture(t)
	first := f.login(t)
	oldJti := jtiOf(t, first.RefreshToken)

	second, err := f.refresh(first.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if second.AccessToken == "" || second.RefreshToken == "" || second.RefreshToken == first.RefreshToken {
		t.Fatal("expected a brand-new token pair")
	}
	if second.ExpiresIn <= 0 {
		t.Error("expected expires_in to be set")
	}
	if f.mr.Exists(refreshTokenKey(oldJti)) {
		t.Error("old refresh jti must be deleted after use")
	}
	if !f.mr.Exists(refreshTokenKey(jtiOf(t, second.RefreshToken))) {
		t.Error("new refresh jti must be stored")
	}
	if _, err := utils.ValidateAccessToken(second.AccessToken, []byte("secret")); err != nil {
		t.Errorf("new access token must be a valid access token: %v", err)
	}
}

func TestRefresh_ReuseRejected(t *testing.T) {
	f := newRefreshFixture(t)
	first := f.login(t)

	if _, err := f.refresh(first.RefreshToken); err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	_, err := f.refresh(first.RefreshToken)
	assertFiberCode(t, err, http.StatusUnauthorized)
}

func TestRefresh_AccessTokenRejected(t *testing.T) {
	f := newRefreshFixture(t)
	first := f.login(t)
	_, err := f.refresh(first.AccessToken)
	assertFiberCode(t, err, http.StatusUnauthorized)
}

func TestRefresh_GarbageToken_Returns401(t *testing.T) {
	f := newRefreshFixture(t)
	_, err := f.refresh("not-a-jwt")
	assertFiberCode(t, err, http.StatusUnauthorized)
}

func TestRefresh_EmptyBody_Returns400(t *testing.T) {
	f := newRefreshFixture(t)
	_, err := f.refresh("")
	assertFiberCode(t, err, http.StatusBadRequest)
}

func TestRefresh_UnknownJti_Returns401(t *testing.T) {
	// Validly signed refresh token that was never stored (e.g. minted before the
	// deploy that introduced rotation).
	f := newRefreshFixture(t)
	tok, _, _ := utils.GenerateRefreshToken(f.user.UserId, []byte("secret"))
	_, err := f.refresh(tok)
	assertFiberCode(t, err, http.StatusUnauthorized)
}

func TestRefresh_InactiveUser_Returns401AndRevokesAll(t *testing.T) {
	f := newRefreshFixture(t)
	a := f.login(t)
	b := f.login(t) // second device

	f.user.IsActive = false
	_, err := f.refresh(a.RefreshToken)
	assertFiberCode(t, err, http.StatusUnauthorized)

	if f.mr.Exists(refreshTokenKey(jtiOf(t, b.RefreshToken))) {
		t.Error("all refresh tokens of an inactive user must be revoked")
	}
}

func TestRefresh_UnverifiedUser_Returns401(t *testing.T) {
	f := newRefreshFixture(t)
	a := f.login(t)
	f.user.IsVerified = false
	_, err := f.refresh(a.RefreshToken)
	assertFiberCode(t, err, http.StatusUnauthorized)
}

func TestRefresh_DeletedUser_Returns401(t *testing.T) {
	f := newRefreshFixture(t)
	a := f.login(t)
	f.user.UserId = "gone" // GetUserById("1") now returns ErrNoRows
	_, err := f.refresh(a.RefreshToken)
	assertFiberCode(t, err, http.StatusUnauthorized)
}

func TestRefresh_DBError_Returns500(t *testing.T) {
	f := newRefreshFixture(t)
	a := f.login(t)
	f.userRepo.getUserByIdFn = func(ctx context.Context, userId string) (*entity.User, error) {
		return nil, errors.New("connection refused")
	}
	_, err := f.refresh(a.RefreshToken)
	assertFiberCode(t, err, http.StatusInternalServerError)
}

func TestRefresh_RedisDown_Returns500(t *testing.T) {
	f := newRefreshFixture(t)
	a := f.login(t)
	f.mr.Close()
	_, err := f.refresh(a.RefreshToken)
	assertFiberCode(t, err, http.StatusInternalServerError)
}

func TestRefresh_PicksUpFreshRoleAndDetails(t *testing.T) {
	// Logged in before completing the profile → no details. After onboarding, a
	// refresh must embed the member details read fresh from the DB.
	f := newRefreshFixture(t)
	first := f.login(t)

	f.memberRepo.findByUserIdFn = func(ctx context.Context, userId string) (*entity.Member, error) {
		return &entity.Member{MemberId: "42"}, nil
	}
	second, err := f.refresh(first.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	claims, err := utils.ValidateAccessToken(second.AccessToken, []byte("secret"))
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	details, ok := claims.Details.(map[string]any)
	if !ok || details["member_id"] != "42" {
		t.Errorf("expected refreshed token to carry member_id 42, got %#v", claims.Details)
	}
}

func TestLogout_RevokesToken(t *testing.T) {
	f := newRefreshFixture(t)
	a := f.login(t)

	if err := f.uc.Logout(context.Background(), &model.RefreshTokenRequest{RefreshToken: a.RefreshToken}); err != nil {
		t.Fatalf("logout: %v", err)
	}
	_, err := f.refresh(a.RefreshToken)
	assertFiberCode(t, err, http.StatusUnauthorized)
}

func TestLogout_InvalidToken_IsNoOp(t *testing.T) {
	f := newRefreshFixture(t)
	if err := f.uc.Logout(context.Background(), &model.RefreshTokenRequest{RefreshToken: "garbage"}); err != nil {
		t.Errorf("logout with an invalid token must succeed (idempotent), got %v", err)
	}
}

func TestLogout_RedisDown_Returns500(t *testing.T) {
	f := newRefreshFixture(t)
	a := f.login(t)
	f.mr.Close()
	err := f.uc.Logout(context.Background(), &model.RefreshTokenRequest{RefreshToken: a.RefreshToken})
	assertFiberCode(t, err, http.StatusInternalServerError)
}

func TestResetPassword_RevokesAllRefreshTokens(t *testing.T) {
	f := newRefreshFixture(t)
	a := f.login(t)
	b := f.login(t)

	// Plant a valid reset token the way issueResetToken stores it.
	f.mr.HSet(otpKey(otpPurposeReset, f.user.Email), "code_hash", utils.HashOTP("reset-token"), "attempts", "0")

	err := f.uc.ResetPassword(context.Background(), &model.ResetPasswordRequest{
		Email: f.user.Email, Token: "reset-token", NewPassword: "PasswordBaru123!",
	})
	if err != nil {
		t.Fatalf("reset password: %v", err)
	}
	for _, tok := range []string{a.RefreshToken, b.RefreshToken} {
		_, err := f.refresh(tok)
		assertFiberCode(t, err, http.StatusUnauthorized)
	}
}
