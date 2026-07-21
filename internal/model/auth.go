package model

type LoginRequest struct {
	Email        string `json:"email"          validate:"required,email"`
	Password     string `json:"password"       validate:"required"`
	ExpectedRole string `json:"expected_role"  validate:"omitempty,oneof=member guru"`
}

type LoginResponse struct {
	User			UserResponse	`json:"user"`
	AccessToken		string			`json:"access_token"`
	RefreshToken	string			`json:"refresh_token"`
}

type RegisterRequest struct {
	Username	string `json:"username" validate:"required,min=3,max=50"`
	Email		string `json:"email"    validate:"required,email,max=100"`
	Password	string `json:"password" validate:"required,min=8"`
}

// VerifyEmailRequest confirms ownership of an email after registration.
type VerifyEmailRequest struct {
	Email	string `json:"email" validate:"required,email"`
	Code	string `json:"code"  validate:"required,len=6,numeric"`
}

// ResendOTPRequest asks for a fresh verification OTP (register flow).
type ResendOTPRequest struct {
	Email	string `json:"email" validate:"required,email"`
}

// ForgotPasswordRequest starts the password-reset flow by emailing a reset link.
type ForgotPasswordRequest struct {
	Email	string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest completes the password-reset flow. The email + token come
// from the reset link the user clicked (…/reset-password?token=…&email=…); the FE
// posts them back alongside the new password.
type ResetPasswordRequest struct {
	Email		string `json:"email"        validate:"required,email"`
	Token		string `json:"token"        validate:"required"`
	NewPassword	string `json:"new_password" validate:"required,min=8"`
}

// MessageResponse is a generic acknowledgement for flows that return no entity
// (verify, reset). The message is intentionally generic for the forgot-password
// flow so it cannot be used to probe which emails are registered.
type MessageResponse struct {
	Message string `json:"message"`
}

// OTPPolicy carries the (static, non-user-specific) OTP/reset-link timing knobs so
// the FE can render countdowns and disable the "resend" button for the right
// duration. All values come from server config; exposing them leaks nothing about
// whether a given email is registered, so it is safe on anti-enumeration endpoints.
type OTPPolicy struct {
	OTPExpiresInSeconds       int `json:"otp_expires_in_seconds"`
	ResetLinkExpiresInSeconds int `json:"reset_link_expires_in_seconds"`
	ResendCooldownSeconds     int `json:"resend_cooldown_seconds"`
}

// RegisterResponse is the register payload: the created user plus the verification
// timing knobs the FE needs to drive the "check your email" screen. UserResponse is
// embedded so its fields stay at the same JSON level as before (id, username, …).
type RegisterResponse struct {
	UserResponse
	OTPExpiresInSeconds   int `json:"otp_expires_in_seconds"`
	ResendCooldownSeconds int `json:"resend_cooldown_seconds"`
}

// OTPActionResponse is the generic acknowledgement for resend/forgot, enriched with
// timing knobs. Fields are omitempty so each endpoint only emits the ones it means:
// resend sends otp_expires_in_seconds, forgot sends reset_link_expires_in_seconds.
type OTPActionResponse struct {
	Message                   string `json:"message"`
	OTPExpiresInSeconds       int    `json:"otp_expires_in_seconds,omitempty"`
	ResetLinkExpiresInSeconds int    `json:"reset_link_expires_in_seconds,omitempty"`
	ResendCooldownSeconds     int    `json:"resend_cooldown_seconds,omitempty"`
}