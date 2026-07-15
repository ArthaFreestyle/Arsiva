package utils

import "testing"

func TestGenerateOTP_SixDigits(t *testing.T) {
	for i := 0; i < 1000; i++ {
		code, err := GenerateOTP()
		if err != nil {
			t.Fatalf("GenerateOTP error: %v", err)
		}
		if len(code) != 6 {
			t.Fatalf("expected 6-char code, got %q (len %d)", code, len(code))
		}
		for _, r := range code {
			if r < '0' || r > '9' {
				t.Fatalf("expected only digits, got %q", code)
			}
		}
	}
}

func TestHashOTP_DeterministicAndOpaque(t *testing.T) {
	h1 := HashOTP("123456")
	h2 := HashOTP("123456")
	if h1 != h2 {
		t.Error("HashOTP must be deterministic for the same input")
	}
	if h1 == "123456" {
		t.Error("HashOTP must not return the plaintext code")
	}
	if HashOTP("123456") == HashOTP("654321") {
		t.Error("different codes must hash differently")
	}
}

func TestCheckOTP(t *testing.T) {
	hash := HashOTP("246810")
	if !CheckOTP("246810", hash) {
		t.Error("CheckOTP should accept the matching code")
	}
	if CheckOTP("000000", hash) {
		t.Error("CheckOTP should reject a non-matching code")
	}
}
