package utils

import (
	"errors"
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(User *entity.User, jwtSecret []byte) (string, string, error) {
	var AccessExpiration = time.Now().Add(60 * time.Minute)
	AccessClaims := model.Claims{
		UserId:      User.UserId,
		Username:    User.Username,
		Email:    User.Email,
		Role:        User.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(AccessExpiration),
		},
	}
	AccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, AccessClaims)
	accessString, err := AccessToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}
	RefreshExpiration := time.Now().Add(7 * 24 * time.Hour)
	RefreshClaims := model.Claims{
		UserId:      User.UserId,
		Username:    User.Username,
		Email:    User.Email,
		Role:        User.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(RefreshExpiration),
		},
	}
	RefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, RefreshClaims)
	refreshString, err := RefreshToken.SignedString([]byte(jwtSecret))

	if err != nil {
		return "", "", err
	}

	return accessString, refreshString, nil
}

func ValidateToken(tokenString string, jwtSecret []byte) (*model.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*model.Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrInvalidKey
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