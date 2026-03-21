package model

import (
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserId		string	`json:"user_id"`
	Email		string	`json:"email"`
	Username	string	`json:"username"`
	Role		string	`json:"role"`
	jwt.RegisteredClaims
}