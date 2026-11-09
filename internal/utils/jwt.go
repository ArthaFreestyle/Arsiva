package utils

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenTTL  = 60 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

var (
	// ErrWrongTokenType is returned when a structurally valid token is presented
	// where a different token type is required (e.g. a refresh token used as a
	// Bearer access token).
	ErrWrongTokenType = errors.New("wrong token type")
	// ErrMissingSubject is returned when a token has no user_id.
	ErrMissingSubject = errors.New("token has no user_id")
)

// GenerateAccessToken mints a short-lived access token carrying the user's
// identity, role and profile details (used by stateless middleware checks).
func GenerateAccessToken(user *entity.User, details any, secret []byte) (string, error) {
	now := time.Now()
	claims := model.Claims{
		UserId:    user.UserId,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Details:   details,
		TokenType: model.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

// GenerateRefreshToken mints a long-lived refresh token. It deliberately carries
// only the user id and a random jti: the jti is what the server stores (and
// deletes on use), and identity/role are re-read from the DB on refresh.
func GenerateRefreshToken(userId string, secret []byte) (token string, jti string, err error) {
	now := time.Now()
	jti = uuid.NewString() // crypto/rand-backed UUIDv4
	claims := model.Claims{
		UserId:    userId,
		TokenType: model.TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenTTL)),
		},
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return "", "", err
	}
	return token, jti, nil
}

// ValidateAccessToken accepts only a valid, unexpired HS256 token with
// token_type "access".
func ValidateAccessToken(tokenString string, secret []byte) (*model.Claims, error) {
	return parseToken(tokenString, secret, model.TokenTypeAccess)
}

// ValidateRefreshToken accepts only a valid, unexpired HS256 token with
// token_type "refresh" and a jti. It does NOT check revocation — callers must
// still consult the refresh-token store.
func ValidateRefreshToken(tokenString string, secret []byte) (*model.Claims, error) {
	claims, err := parseToken(tokenString, secret, model.TokenTypeRefresh)
	if err != nil {
		return nil, err
	}
	if claims.ID == "" {
		return nil, jwt.ErrTokenInvalidId
	}
	return claims, nil
}

func parseToken(tokenString string, secret []byte, expectedType string) (*model.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.Claims{},
		func(token *jwt.Token) (interface{}, error) { return secret, nil },
		// Pin the algorithm so a token can never pick its own verification method.
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*model.Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	if claims.TokenType != expectedType {
		return nil, ErrWrongTokenType
	}
	if claims.UserId == "" {
		return nil, ErrMissingSubject
	}
	return claims, nil
}

func ExtractExpiration(tokenString string) (time.Time, error) {
	// parse JWT tanpa cek signature
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &model.Claims{})
	if err != nil {
		return time.Time{}, err
	}

	claims, ok := token.Claims.(*model.Claims)
	if !ok {
		return time.Time{}, errors.New("cannot parse claims")
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, errors.New("no expiration in token")
	}

	return claims.ExpiresAt.Time, nil
}
