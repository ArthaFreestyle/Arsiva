package model

type LoginRequest struct {
	Email	string	`json:"email" validate:"required,email"`
	Password	string	`json:"password" validate:"required"`
}

type LoginResponse struct {
	User	UserResponse	`json:"user"`
	AccessToken	string		`json:"access_token"`
	RefreshToken	string	`json:"refresh_token"`
}