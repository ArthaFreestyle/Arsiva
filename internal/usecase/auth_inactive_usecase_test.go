package usecase

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/utils"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-playground/validator/v10"
)

// Tests for issue #52: a deactivated (soft-deleted, is_active=false) account must
// be locked out of every auth flow, without leaking its deactivated state to
// anyone who does not know the password.

// countingMailer records how many emails were sent.
type countingMailer struct{ sent int }

func (m *countingMailer) Send(to, subject, body string) error { m.sent++; return nil }
func (m *countingMailer) SendHTML(to, subject, htmlBody, textBody string) error {
	m.sent++
	return nil
}

type inactiveFixture struct {
	uc            AuthUseCase
	mr            *miniredis.Miniredis
	mail          *countingMailer
	user          *entity.User
	passwordCalls int
	verifyCalls   int
}

func newInactiveFixture(t *testing.T, active, verified bool) *inactiveFixture {
	t.Helper()
	hash, _ := utils.HashPassword("Rahasia123!")
	f := &inactiveFixture{
		mail: &countingMailer{},
		user: &entity.User{
			UserId: "1", Username: "siswa", Email: "siswa@example.com", Role: "member",
			PasswordHash: hash, IsActive: active, IsVerified: verified,
		},
	}
	repo := &mockUserRepo{
		findByEmailFn: func(ctx context.Context, email string) (*entity.User, error) {
			u := *f.user
			return &u, nil
		},
		updatePasswordFn: func(ctx context.Context, userId, passwordHash string) error {
			f.passwordCalls++
			return nil
		},
	}
	rdb, mr := newTestRedis(t)
	f.mr = mr
	f.uc = NewAuthUseCase(repo, []byte("secret"), validator.New(), discardLogger(), nil, nil, nil, rdb, f.mail, 5*time.Minute, 15*time.Minute, 5, time.Minute, "https://arsiva.id/reset-password")
	return f
}

func TestLogin_Inactive_CorrectPassword_Returns403_NoTokens(t *testing.T) {
	f := newInactiveFixture(t, false, true)

	resp, err := f.uc.Login(context.Background(), &model.LoginRequest{Email: f.user.Email, Password: "Rahasia123!"})
	assertFiberCode(t, err, http.StatusForbidden)
	if resp != nil {
		t.Error("expected no tokens for a deactivated account")
	}
	if keys := f.mr.Keys(); len(keys) != 0 {
		t.Errorf("expected nothing stored in Redis, got %v", keys)
	}
}

func TestLogin_Inactive_WrongPassword_Returns401(t *testing.T) {
	// Must be 401, not 403: a wrong-password caller must not learn the account
	// exists but is deactivated.
	f := newInactiveFixture(t, false, true)
	_, err := f.uc.Login(context.Background(), &model.LoginRequest{Email: f.user.Email, Password: "SalahPassword!"})
	assertFiberCode(t, err, http.StatusUnauthorized)
}

func TestLogin_Reactivated_Succeeds(t *testing.T) {
	f := newInactiveFixture(t, false, true)
	_, err := f.uc.Login(context.Background(), &model.LoginRequest{Email: f.user.Email, Password: "Rahasia123!"})
	assertFiberCode(t, err, http.StatusForbidden)

	f.user.IsActive = true // RestoreUser
	if _, err := f.uc.Login(context.Background(), &model.LoginRequest{Email: f.user.Email, Password: "Rahasia123!"}); err != nil {
		t.Fatalf("expected login to succeed after restore, got %v", err)
	}
}

func TestForgotPassword_Inactive_GenericSuccess_NoMail(t *testing.T) {
	f := newInactiveFixture(t, false, true)
	if err := f.uc.ForgotPassword(context.Background(), &model.ForgotPasswordRequest{Email: f.user.Email}); err != nil {
		t.Fatalf("expected generic success, got %v", err)
	}
	if f.mail.sent != 0 {
		t.Errorf("expected no reset mail for a deactivated account, got %d", f.mail.sent)
	}
}

func TestForgotPassword_Active_SendsMail(t *testing.T) {
	// Control case: proves the fixture would detect a sent mail.
	f := newInactiveFixture(t, true, true)
	if err := f.uc.ForgotPassword(context.Background(), &model.ForgotPasswordRequest{Email: f.user.Email}); err != nil {
		t.Fatalf("forgot password: %v", err)
	}
	if f.mail.sent != 1 {
		t.Errorf("expected 1 reset mail, got %d", f.mail.sent)
	}
}

func TestResetPassword_Inactive_Returns400_NoUpdate(t *testing.T) {
	// Reset link mailed before the admin deactivated the account.
	f := newInactiveFixture(t, false, true)
	f.mr.HSet(otpKey(otpPurposeReset, f.user.Email), "code_hash", utils.HashOTP("reset-token"), "attempts", "0")

	err := f.uc.ResetPassword(context.Background(), &model.ResetPasswordRequest{
		Email: f.user.Email, Token: "reset-token", NewPassword: "PasswordBaru123!",
	})
	assertFiberCode(t, err, http.StatusBadRequest)
	if f.passwordCalls != 0 {
		t.Errorf("UpdatePassword must not be called for a deactivated account, called %d times", f.passwordCalls)
	}
}

func TestResendOTP_Inactive_NoMail(t *testing.T) {
	f := newInactiveFixture(t, false, false)
	if err := f.uc.ResendVerificationOTP(context.Background(), &model.ResendOTPRequest{Email: f.user.Email}); err != nil {
		t.Fatalf("expected generic success, got %v", err)
	}
	if f.mail.sent != 0 {
		t.Errorf("expected no OTP mail for a deactivated account, got %d", f.mail.sent)
	}
}

func TestVerifyEmail_Inactive_Returns400(t *testing.T) {
	f := newInactiveFixture(t, false, false)
	f.mr.HSet(otpKey(otpPurposeVerify, f.user.Email), "code_hash", utils.HashOTP("123456"), "attempts", "0")

	err := f.uc.VerifyEmail(context.Background(), &model.VerifyEmailRequest{Email: f.user.Email, Code: "123456"})
	assertFiberCode(t, err, http.StatusBadRequest)
}
