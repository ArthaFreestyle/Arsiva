package usecase

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/mailer"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
	"ArthaFreestyle/Arsiva/internal/utils"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type AuthUseCase interface {
	Login(ctx context.Context, request *model.LoginRequest) (*model.LoginResponse, error)
	RegisterMember(ctx context.Context, request *model.RegisterRequest) (*model.UserResponse, error)
	RegisterGuru(ctx context.Context, request *model.RegisterRequest) (*model.UserResponse, error)

	// VerifyEmail confirms a freshly-registered account via the OTP mailed to it.
	VerifyEmail(ctx context.Context, request *model.VerifyEmailRequest) error
	// ResendVerificationOTP re-issues a verification OTP for an unverified account.
	ResendVerificationOTP(ctx context.Context, request *model.ResendOTPRequest) error
	// ForgotPassword emails a password-reset link. Always succeeds from the
	// caller's view (anti-enumeration) regardless of whether the email is registered.
	ForgotPassword(ctx context.Context, request *model.ForgotPasswordRequest) error
	// ResetPassword sets a new password after a valid reset token (from the link).
	ResetPassword(ctx context.Context, request *model.ResetPasswordRequest) error
	// OTPPolicy returns the (static) OTP/reset-link timing knobs so controllers can
	// hand them to the FE for countdowns and resend-button throttling.
	OTPPolicy() model.OTPPolicy
}

type AuthUseCaseImpl struct {
	DB               *pgxpool.Pool
	Log              *logrus.Logger
	Validate         *validator.Validate
	Repo             repository.UserRepository
	GuruRepository   repository.GuruRepository
	MemberRepository repository.MemberRepository
	Secret           []byte
	Redis            *redis.Client
	Mailer         mailer.Mailer
	OTPTTL         time.Duration // TTL of a verification OTP (register flow)
	ResetTTL       time.Duration // TTL of a password-reset link — longer, since it travels via email
	OTPMaxAttempts int
	ResendCooldown time.Duration
	// ResetBaseURL is the frontend page the reset link points at, e.g.
	// "https://arsiva.id/reset-password". The token + email are appended as query
	// params when building the link.
	ResetBaseURL string
}

func NewAuthUseCase(repo repository.UserRepository, secret []byte, validate *validator.Validate, log *logrus.Logger, DB *pgxpool.Pool, guruRepo repository.GuruRepository, memberRepo repository.MemberRepository, redisClient *redis.Client, mail mailer.Mailer, otpTTL time.Duration, resetTTL time.Duration, otpMaxAttempts int, resendCooldown time.Duration, resetBaseURL string) AuthUseCase {
	return &AuthUseCaseImpl{
		Repo:             repo,
		Secret:           secret,
		Validate:         validate,
		DB:               DB,
		Log:              log,
		GuruRepository:   guruRepo,
		MemberRepository: memberRepo,
		Redis:            redisClient,
		Mailer:           mail,
		OTPTTL:           otpTTL,
		ResetTTL:         resetTTL,
		OTPMaxAttempts:   otpMaxAttempts,
		ResendCooldown:   resendCooldown,
		ResetBaseURL:     resetBaseURL,
	}
}

// OTPPolicy exposes the static timing knobs (seconds) for the FE.
func (c *AuthUseCaseImpl) OTPPolicy() model.OTPPolicy {
	return model.OTPPolicy{
		OTPExpiresInSeconds:       int(c.OTPTTL.Seconds()),
		ResetLinkExpiresInSeconds: int(c.ResetTTL.Seconds()),
		ResendCooldownSeconds:     int(c.ResendCooldown.Seconds()),
	}
}

// ttlFor returns the secret lifetime for a given purpose: a short OTP for email
// verification, a longer window for the reset link (it round-trips through email).
func (c *AuthUseCaseImpl) ttlFor(purpose string) time.Duration {
	if purpose == otpPurposeReset {
		return c.ResetTTL
	}
	return c.OTPTTL
}

func (c *AuthUseCaseImpl) Login(ctx context.Context, request *model.LoginRequest) (*model.LoginResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("Invalid request body  : %+v", err)
		return nil, fiber.ErrBadRequest
	}

	request.Email = strings.ToLower(request.Email)

	user, err := c.Repo.FindByEmail(ctx, request.Email)
	if err != nil {
		c.Log.Warnf("Failed find user by email : %+v", err)
		return nil, fiber.ErrUnauthorized
	}

	if !utils.CheckPasswordHash(request.Password, user.PasswordHash) {
		c.Log.Warnf("Invalid password : %+v", err)
		return nil, fiber.ErrUnauthorized
	}

	// Verification check AFTER password verification so an attacker without the
	// correct password cannot learn whether an email is registered/verified.
	if !user.IsVerified {
		c.Log.Warnf("Login blocked: email not verified for user %s", user.UserId)
		return nil, fiber.NewError(fiber.StatusForbidden, "email belum diverifikasi, cek email untuk kode verifikasi")
	}

	// Role check AFTER password verification so a wrong expected_role
	// cannot be used to enumerate roles on accounts with bad passwords.
	if request.ExpectedRole != "" && request.ExpectedRole != user.Role {
		c.Log.Warnf("Login role mismatch: expected_role=%s actual=%s", request.ExpectedRole, user.Role)
		return nil, fiber.NewError(fiber.StatusForbidden, "wrong login page for this account")
	}

	// Embed the profile details into the JWT so downstream stateless checks
	// (RequireProfileComplete, member_id extraction) work without DB hits.
	// A missing profile row (pgx.ErrNoRows) is legitimate — the user just hasn't
	// completed onboarding, so leave details nil and let ProfileCompleteMiddleware
	// gate them. Any OTHER error must NOT be swallowed: doing so mints a token
	// with empty Details that permanently 403s every member_id/profile-based
	// endpoint with no way to diagnose it.
	var details any
	switch user.Role {
	case "guru":
		if c.GuruRepository != nil {
			guru, err := c.GuruRepository.FindByUserId(ctx, user.UserId)
			switch {
			case err == nil:
				details = model.GuruDetails{
					GuruId:     guru.GuruId,
					NIP:        guru.NIP,
					BidangAjar: guru.BidangAjar,
					SekolahId:  guru.SekolahId,
				}
			case errors.Is(err, pgx.ErrNoRows):
				c.Log.Infof("Guru profile not yet created for user %s; issuing token without details", user.UserId)
			default:
				c.Log.Errorf("Failed to load guru profile for user %s: %+v", user.UserId, err)
				return nil, fiber.ErrInternalServerError
			}
		}
	case "member":
		if c.MemberRepository != nil {
			member, err := c.MemberRepository.FindByUserId(ctx, user.UserId)
			switch {
			case err == nil:
				details = model.MemberDetails{
					MemberId:  member.MemberId,
					NIS:       member.NIS,
					SekolahId: member.SekolahId,
					Level:     member.Level,
				}
			case errors.Is(err, pgx.ErrNoRows):
				c.Log.Infof("Member profile not yet created for user %s; issuing token without details", user.UserId)
			default:
				c.Log.Errorf("Failed to load member profile for user %s: %+v", user.UserId, err)
				return nil, fiber.ErrInternalServerError
			}
		}
	}

	access, refresh, err := utils.GenerateToken(user, details, c.Secret)
	if err != nil {
		c.Log.Warnf("Failed generate token : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	AuthResponse := converter.ToUserResponse(user)

	return &model.LoginResponse{
		User:         *AuthResponse,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (c *AuthUseCaseImpl) RegisterMember(ctx context.Context, request *model.RegisterRequest) (*model.UserResponse, error) {
	return c.register(ctx, request, "member")
}

func (c *AuthUseCaseImpl) RegisterGuru(ctx context.Context, request *model.RegisterRequest) (*model.UserResponse, error) {
	return c.register(ctx, request, "guru")
}

func (c *AuthUseCaseImpl) register(ctx context.Context, request *model.RegisterRequest, role string) (*model.UserResponse, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("Invalid register request body : %+v", err)
		return nil, fiber.ErrBadRequest
	}

	request.Email = strings.ToLower(request.Email)

	hashedPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		c.Log.Warnf("Failed to hash password : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	user := &entity.User{
		Username:     request.Username,
		Email:        request.Email,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	created, err := c.Repo.CreateUser(ctx, user)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return nil, fiber.NewError(fiber.StatusConflict, "email or username already in use")
		}
		c.Log.Warnf("Failed to create user : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	// Mail the verification OTP. A mail failure must NOT fail registration — the
	// account exists and the user can request a fresh code via ResendVerificationOTP.
	// (This also keeps registration working in local dev where no relay is reachable.)
	if err := c.issueOTP(ctx, otpPurposeVerify, created.Email); err != nil {
		c.Log.Warnf("Failed to send verification OTP to %s (registration still succeeded): %+v", created.Email, err)
	}

	return converter.ToUserResponse(created), nil
}

// ─── OTP flows (email verification + password reset) ─────────────────────────

const (
	otpPurposeVerify = "verify"
	otpPurposeReset  = "reset"
)

func otpKey(purpose, email string) string      { return fmt.Sprintf("otp:%s:%s", purpose, email) }
func otpCooldownKey(purpose, email string) string {
	return fmt.Sprintf("otp:cooldown:%s:%s", purpose, email)
}

// storeSecret persists the hash of a single-use secret (OTP code or reset token)
// in Redis with a TTL, resets its attempt counter, and arms a resend cooldown.
// Returns a 429 fiber error if a cooldown is still active. Shared by issueOTP and
// issueResetToken so both flows have identical throttling/storage semantics.
func (c *AuthUseCaseImpl) storeSecret(ctx context.Context, purpose, email, secretHash string) error {
	cooldownKey := otpCooldownKey(purpose, email)
	if n, _ := c.Redis.Exists(ctx, cooldownKey).Result(); n > 0 {
		return fiber.NewError(fiber.StatusTooManyRequests, "tunggu sebentar sebelum meminta kode baru")
	}

	key := otpKey(purpose, email)
	pipe := c.Redis.Pipeline()
	pipe.Del(ctx, key) // drop any previous secret so attempt counters reset
	pipe.HSet(ctx, key, map[string]any{"code_hash": secretHash, "attempts": 0})
	pipe.Expire(ctx, key, c.ttlFor(purpose))
	pipe.Set(ctx, cooldownKey, "1", c.ResendCooldown)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("store secret: %w", err)
	}
	return nil
}

// issueOTP generates a fresh OTP, stores its hash in Redis (with TTL + a resend
// cooldown), and emails the plaintext code. Returns an error on cooldown, Redis
// failure, or mail failure — callers decide whether to surface or swallow it.
// Used by the email-verification flow (password reset uses issueResetToken).
func (c *AuthUseCaseImpl) issueOTP(ctx context.Context, purpose, email string) error {
	if c.Redis == nil || c.Mailer == nil {
		return fmt.Errorf("otp infrastructure not configured")
	}

	code, err := utils.GenerateOTP()
	if err != nil {
		return fmt.Errorf("generate otp: %w", err)
	}

	if err := c.storeSecret(ctx, purpose, email, utils.HashOTP(code)); err != nil {
		return err
	}

	subject, htmlBody, textBody := c.otpEmailContent(code)
	if err := c.Mailer.SendHTML(email, subject, htmlBody, textBody); err != nil {
		return fmt.Errorf("send otp mail: %w", err)
	}
	return nil
}

// issueResetToken generates a fresh password-reset token, stores its hash in Redis
// (reusing the OTP storage/cooldown semantics under the "reset" purpose), and emails
// a clickable reset link pointing at ResetBaseURL. Like issueOTP, callers decide
// whether to surface or swallow the returned error.
func (c *AuthUseCaseImpl) issueResetToken(ctx context.Context, email string) error {
	if c.Redis == nil || c.Mailer == nil {
		return fmt.Errorf("otp infrastructure not configured")
	}

	token, err := utils.GenerateResetToken()
	if err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}

	if err := c.storeSecret(ctx, otpPurposeReset, email, utils.HashOTP(token)); err != nil {
		return err
	}

	subject, htmlBody, textBody := c.resetEmailContent(token, email)
	if err := c.Mailer.SendHTML(email, subject, htmlBody, textBody); err != nil {
		return fmt.Errorf("send reset mail: %w", err)
	}
	return nil
}

// buildResetURL appends the token + email as query params to the configured reset
// page URL, e.g. https://arsiva.id/reset-password?token=<t>&email=<e>.
func (c *AuthUseCaseImpl) buildResetURL(token, email string) string {
	sep := "?"
	if strings.Contains(c.ResetBaseURL, "?") {
		sep = "&"
	}
	return fmt.Sprintf("%s%stoken=%s&email=%s", c.ResetBaseURL, sep, url.QueryEscape(token), url.QueryEscape(email))
}

// consumeOTP validates a code against the stored hash. On success it deletes the
// key (single-use). On a wrong code it increments the attempt counter and, once
// OTPMaxAttempts is reached, invalidates the code entirely.
func (c *AuthUseCaseImpl) consumeOTP(ctx context.Context, purpose, email, code string) error {
	if c.Redis == nil {
		return fiber.ErrInternalServerError
	}

	key := otpKey(purpose, email)
	data, err := c.Redis.HGetAll(ctx, key).Result()
	if err != nil {
		c.Log.Errorf("consumeOTP: redis error for %s: %+v", key, err)
		return fiber.ErrInternalServerError
	}
	if len(data) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "kode tidak valid atau sudah kedaluwarsa")
	}

	attempts, _ := strconv.Atoi(data["attempts"])
	if attempts >= c.OTPMaxAttempts {
		c.Redis.Del(ctx, key)
		return fiber.NewError(fiber.StatusBadRequest, "terlalu banyak percobaan, minta kode baru")
	}

	if !utils.CheckOTP(code, data["code_hash"]) {
		c.Redis.HIncrBy(ctx, key, "attempts", 1)
		return fiber.NewError(fiber.StatusBadRequest, "kode salah")
	}

	c.Redis.Del(ctx, key)
	return nil
}

// otpEmailContent builds the subject plus HTML and plain-text bodies for a
// verification OTP email (the only remaining OTP purpose — password reset uses a
// link, see resetEmailContent). The HTML is rendered from the shared Arsiva OTP
// template; the text body is the fallback for non-HTML clients. If HTML rendering
// ever fails it degrades gracefully to a text-only body reused for both parts.
func (c *AuthUseCaseImpl) otpEmailContent(code string) (subject, htmlBody, textBody string) {
	mins := int(c.OTPTTL.Minutes())

	data := mailer.OTPEmail{
		Eyebrow:      "Keamanan Akun",
		Code:         code,
		ExpiryMins:   mins,
		Heading:      "Verifikasi email kamu",
		Intro:        "Masukkan kode di bawah ini untuk menyelesaikan verifikasi email akun Arsiva kamu.",
		CodeLabel:    "Kode verifikasi kamu",
		SecurityNote: "Tidak merasa mendaftar di Arsiva? Kamu bisa mengabaikan email ini dengan aman.",
		Preheader:    fmt.Sprintf("Kode verifikasi Arsiva kamu adalah %s. Berlaku %d menit.", code, mins),
	}
	subj := "Kode Verifikasi Email Arsiva"

	text := mailer.RenderOTPText(data)
	html, err := mailer.RenderOTPHTML(data)
	if err != nil {
		c.Log.Warnf("otpEmailContent: failed to render HTML email, falling back to text: %+v", err)
		html = text
	}
	return subj, html, text
}

// resetEmailContent builds the subject plus HTML/text bodies for the password-reset
// link email. Mirrors otpEmailContent's graceful-degradation behaviour.
func (c *AuthUseCaseImpl) resetEmailContent(token, email string) (subject, htmlBody, textBody string) {
	mins := int(c.ResetTTL.Minutes())

	data := mailer.ResetLinkEmail{
		Heading:      "Atur ulang password kamu",
		Intro:        "Kami menerima permintaan untuk mengatur ulang password akun Arsiva kamu. Klik tombol di bawah ini untuk membuat password baru.",
		ButtonLabel:  "Reset Password",
		ResetURL:     c.buildResetURL(token, email),
		ExpiryMins:   mins,
		SecurityNote: "Tidak meminta reset password? Kamu bisa mengabaikan email ini dengan aman dan password kamu tetap tidak berubah.",
		Preheader:    fmt.Sprintf("Atur ulang password Arsiva kamu. Tautan berlaku %d menit.", mins),
	}

	text := mailer.RenderResetLinkText(data)
	html, err := mailer.RenderResetLinkHTML(data)
	if err != nil {
		c.Log.Warnf("resetEmailContent: failed to render HTML email, falling back to text: %+v", err)
		html = text
	}
	return "Reset Password Arsiva", html, text
}

// VerifyEmail confirms an account via its verification OTP.
func (c *AuthUseCaseImpl) VerifyEmail(ctx context.Context, request *model.VerifyEmailRequest) error {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("VerifyEmail: invalid request: %+v", err)
		return fiber.ErrBadRequest
	}
	email := strings.ToLower(request.Email)

	if err := c.consumeOTP(ctx, otpPurposeVerify, email, request.Code); err != nil {
		return err
	}

	user, err := c.Repo.FindByEmail(ctx, email)
	if err != nil {
		// OTP matched but the account vanished — treat as invalid rather than 500.
		c.Log.Warnf("VerifyEmail: user not found after OTP match for %s: %+v", email, err)
		return fiber.NewError(fiber.StatusBadRequest, "kode tidak valid atau sudah kedaluwarsa")
	}
	if user.IsVerified {
		return nil
	}

	if err := c.Repo.MarkVerified(ctx, user.UserId); err != nil {
		c.Log.Errorf("VerifyEmail: failed to mark verified for %s: %+v", user.UserId, err)
		return fiber.ErrInternalServerError
	}
	return nil
}

// ResendVerificationOTP re-issues a verification OTP. Responses are kept generic
// (the controller returns a fixed message): a non-existent or already-verified
// account is silently a no-op so this endpoint cannot enumerate accounts.
func (c *AuthUseCaseImpl) ResendVerificationOTP(ctx context.Context, request *model.ResendOTPRequest) error {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("ResendVerificationOTP: invalid request: %+v", err)
		return fiber.ErrBadRequest
	}
	email := strings.ToLower(request.Email)

	user, err := c.Repo.FindByEmail(ctx, email)
	if err != nil || user.IsVerified {
		return nil // no-op; do not reveal existence/verification state
	}

	if err := c.issueOTP(ctx, otpPurposeVerify, email); err != nil {
		// Surface only the rate-limit signal; swallow mail/redis errors as generic success.
		var fe *fiber.Error
		if errors.As(err, &fe) && fe.Code == fiber.StatusTooManyRequests {
			return fe
		}
		c.Log.Warnf("ResendVerificationOTP: issueOTP failed for %s: %+v", email, err)
	}
	return nil
}

// ForgotPassword emails a password-reset link. From the caller's perspective it
// always succeeds regardless of whether the email exists (anti-enumeration); all
// internal failures are logged and swallowed.
func (c *AuthUseCaseImpl) ForgotPassword(ctx context.Context, request *model.ForgotPasswordRequest) error {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("ForgotPassword: invalid request: %+v", err)
		return fiber.ErrBadRequest
	}
	email := strings.ToLower(request.Email)

	user, err := c.Repo.FindByEmail(ctx, email)
	if err != nil {
		c.Log.Infof("ForgotPassword: no account for %s (returning generic success)", email)
		return nil
	}

	if err := c.issueResetToken(ctx, user.Email); err != nil {
		// Swallow everything (including cooldown) so this endpoint stays uniform.
		c.Log.Warnf("ForgotPassword: issueResetToken failed for %s: %+v", email, err)
	}
	return nil
}

// ResetPassword sets a new password after a valid reset token from the emailed link.
func (c *AuthUseCaseImpl) ResetPassword(ctx context.Context, request *model.ResetPasswordRequest) error {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("ResetPassword: invalid request: %+v", err)
		return fiber.ErrBadRequest
	}
	email := strings.ToLower(request.Email)

	// consumeOTP validates the token the same way it validates an OTP code: it
	// compares SHA-256 hashes, so it works for the reset token too.
	if err := c.consumeOTP(ctx, otpPurposeReset, email, request.Token); err != nil {
		return err
	}

	user, err := c.Repo.FindByEmail(ctx, email)
	if err != nil {
		c.Log.Warnf("ResetPassword: user not found after OTP match for %s: %+v", email, err)
		return fiber.NewError(fiber.StatusBadRequest, "kode tidak valid atau sudah kedaluwarsa")
	}

	hashed, err := utils.HashPassword(request.NewPassword)
	if err != nil {
		c.Log.Errorf("ResetPassword: hash failed: %+v", err)
		return fiber.ErrInternalServerError
	}

	if err := c.Repo.UpdatePassword(ctx, user.UserId, hashed); err != nil {
		c.Log.Errorf("ResetPassword: update failed for %s: %+v", user.UserId, err)
		return fiber.ErrInternalServerError
	}
	return nil
}