package utils

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var testSecret = []byte("test-secret")

func testUser() *entity.User {
	return &entity.User{UserId: "7", Username: "siswa", Email: "siswa@example.com", Role: "member"}
}

func TestAccessToken_OnlyValidAsAccess(t *testing.T) {
	access, err := GenerateAccessToken(testUser(), model.MemberDetails{MemberId: "42"}, testSecret)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	claims, err := ValidateAccessToken(access, testSecret)
	if err != nil {
		t.Fatalf("expected access token to validate, got %v", err)
	}
	if claims.UserId != "7" || claims.Role != "member" || claims.TokenType != model.TokenTypeAccess {
		t.Errorf("unexpected claims: %+v", claims)
	}
	if claims.IssuedAt == nil {
		t.Error("expected iat to be set")
	}

	if _, err := ValidateRefreshToken(access, testSecret); !errors.Is(err, ErrWrongTokenType) {
		t.Errorf("access token must be rejected as refresh token, got %v", err)
	}
}

func TestRefreshToken_OnlyValidAsRefresh_AndMinimalClaims(t *testing.T) {
	refresh, jti, err := GenerateRefreshToken("7", testSecret)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if jti == "" {
		t.Fatal("expected a jti")
	}

	claims, err := ValidateRefreshToken(refresh, testSecret)
	if err != nil {
		t.Fatalf("expected refresh token to validate, got %v", err)
	}
	if claims.ID != jti || claims.UserId != "7" || claims.IssuedAt == nil {
		t.Errorf("unexpected claims: %+v", claims)
	}
	if claims.Role != "" || claims.Email != "" || claims.Username != "" || claims.Details != nil {
		t.Errorf("refresh token must not carry identity/role/details, got %+v", claims)
	}

	if _, err := ValidateAccessToken(refresh, testSecret); !errors.Is(err, ErrWrongTokenType) {
		t.Errorf("refresh token must be rejected as access token, got %v", err)
	}
}

func TestRefreshToken_JtiIsUnique(t *testing.T) {
	_, a, _ := GenerateRefreshToken("7", testSecret)
	_, b, _ := GenerateRefreshToken("7", testSecret)
	if a == b {
		t.Errorf("expected unique jtis, got %q twice", a)
	}
}

func TestValidate_RejectsWrongSecret(t *testing.T) {
	access, _ := GenerateAccessToken(testUser(), nil, testSecret)
	if _, err := ValidateAccessToken(access, []byte("other-secret")); err == nil {
		t.Error("expected token signed with another secret to be rejected")
	}
}

func TestValidate_RejectsExpired(t *testing.T) {
	claims := model.Claims{
		UserId:    "7",
		TokenType: model.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(testSecret)
	if _, err := ValidateAccessToken(tok, testSecret); !errors.Is(err, jwt.ErrTokenExpired) {
		t.Errorf("expected expired error, got %v", err)
	}
}

func TestValidate_RejectsMissingExpiry(t *testing.T) {
	claims := model.Claims{UserId: "7", TokenType: model.TokenTypeAccess}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(testSecret)
	if _, err := ValidateAccessToken(tok, testSecret); err == nil {
		t.Error("expected token without exp to be rejected")
	}
}

func TestValidate_RejectsOtherAlgorithm(t *testing.T) {
	claims := model.Claims{
		UserId:    "7",
		TokenType: model.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(testSecret)
	if _, err := ValidateAccessToken(tok, testSecret); err == nil {
		t.Error("expected HS512 token to be rejected (only HS256 allowed)")
	}
}

func TestValidate_RejectsTokenWithoutType(t *testing.T) {
	// Shape of every token minted before token_type existed.
	claims := model.Claims{
		UserId: "7",
		Role:   "member",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(testSecret)
	if _, err := ValidateAccessToken(tok, testSecret); !errors.Is(err, ErrWrongTokenType) {
		t.Errorf("expected legacy token without token_type to be rejected, got %v", err)
	}
}

func TestValidate_RejectsGroupInviteToken(t *testing.T) {
	// Same shape as GroupUseCase.buildGroupInviteToken, signed with the same secret.
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"group_id": "g1",
		"guru_id":  3,
		"type":     "group_invite",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}).SignedString(testSecret)
	if _, err := ValidateAccessToken(tok, testSecret); err == nil {
		t.Error("group_invite token must not be accepted as an access token")
	}
}
