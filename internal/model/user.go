package model

import "time"

type UserResponse struct {
	ID			string 		`json:"id"`
	Username	string 		`json:"username,omitempty"`
	Email		string 		`json:"email,omitempty"`
	Role		string 		`json:"role,omitempty"`
	CreatedAt	*time.Time 	`json:"created_at,omitempty"`
}

type UserRequest struct {
	Username	string `json:"username" validate:"required"`
	Email		string `json:"email" validate:"required,email"`
	Password	string `json:"password" validate:"required"`
	Role		string `json:"role" validate:"required"`
}