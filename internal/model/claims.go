package model

import (
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserId		string	`json:"user_id"`
	Email		string	`json:"email"`
	Username	string	`json:"username"`
	Role		string	`json:"role"`
	Details		any		`json:"details,omitempty"`
	jwt.RegisteredClaims
}

type GuruDetails struct {
	GuruId		string	`json:"guru_id"`
	NIP			string	`json:"nip,omitempty"`
	BidangAjar	string	`json:"bidang_ajar,omitempty"`
	SekolahId	string	`json:"sekolah_id,omitempty"`
}

type MemberDetails struct {
	MemberId	string	`json:"member_id"`
	NIS			string	`json:"nis,omitempty"`
	SekolahId	string	`json:"sekolah_id,omitempty"`
	Level		int		`json:"level,omitempty"`
}