package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
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

// HashOTP returns the hex-encoded SHA-256 of the code. We never store the raw OTP
// in Redis — only its hash — so a Redis dump does not leak live codes.
func HashOTP(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

// CheckOTP compares a plaintext code against a stored hash in constant time.
func CheckOTP(code, storedHash string) bool {
	return subtle.ConstantTimeCompare([]byte(HashOTP(code)), []byte(storedHash)) == 1
}
