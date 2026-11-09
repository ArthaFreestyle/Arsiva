package model

import (
	"github.com/golang-jwt/jwt/v5"
)

// Token types carried in Claims.TokenType. AuthMiddleware accepts ONLY access
// tokens; POST /v1/refresh and /v1/logout accept ONLY refresh tokens. Any other
// JWT signed with the same secret (e.g. a group_invite token, or a token minted
// before this claim existed) has an empty/foreign type and is rejected by both.
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Claims is shared by access and refresh tokens. A refresh token carries only
// user_id + token_type + the registered claims (jti/exp/iat); identity fields are
// omitempty so they are absent from it — the refresh endpoint re-reads role and
// profile details from the DB instead of trusting a 7-day-old snapshot.
type Claims struct {
	UserId		string	`json:"user_id"`
	Email		string	`json:"email,omitempty"`
	Username	string	`json:"username,omitempty"`
	Role		string	`json:"role,omitempty"`
	Details		any		`json:"details,omitempty"`
	TokenType	string	`json:"token_type"`
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