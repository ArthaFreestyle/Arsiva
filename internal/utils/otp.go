package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"math/big"
)

// GenerateOTP returns a cryptographically-random 6-digit numeric code (000000–999999),
// zero-padded so it is always 6 characters.
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	code := n.Int64()
	// %06d zero-pads; build without fmt to keep it allocation-light.
	buf := []byte("000000")
	for i := 5; i >= 0; i-- {
		buf[i] = byte('0' + code%10)
		code /= 10
	}
	return string(buf), nil
}

// GenerateResetToken returns a cryptographically-random, URL-safe token used in
// password-reset links (e.g. .../reset-password?token=<this>). It carries 32 bytes
// of entropy — far more than a 6-digit OTP — because a reset link is not attempt-
// limited by a human typing it and may sit in an inbox. base64 URL encoding
// (no padding) keeps it safe to drop straight into a query string.
func GenerateResetToken() (string, error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf[:]), nil
}

// HashOTP returns the hex-encoded SHA-256 of the code. We never store the raw OTP
// (or reset token) in Redis — only its hash — so a Redis dump does not leak live
// codes/tokens.
func HashOTP(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

// CheckOTP compares a plaintext code against a stored hash in constant time.
func CheckOTP(code, storedHash string) bool {
	return subtle.ConstantTimeCompare([]byte(HashOTP(code)), []byte(storedHash)) == 1
}
