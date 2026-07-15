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

// ForgotPasswordRequest starts the password-reset flow by emailing an OTP.
type ForgotPasswordRequest struct {
	Email	string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest completes the password-reset flow.
type ResetPasswordRequest struct {
	Email		string `json:"email"        validate:"required,email"`
	Code		string `json:"code"         validate:"required,len=6,numeric"`
	NewPassword	string `json:"new_password" validate:"required,min=8"`
}

// MessageResponse is a generic acknowledgement for flows that return no entity
// (verify, resend, forgot, reset). The message is intentionally generic for the
// forgot-password flow so it cannot be used to probe which emails are registered.
type MessageResponse struct {
	Message string `json:"message"`
}