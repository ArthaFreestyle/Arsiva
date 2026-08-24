package middleware

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// AuthLimiterKey buckets the auth rate limiter per (account, route) instead of
// per source IP alone. Many legitimate users can share one public IP (school/
// office wifi, NAT) — a pure c.IP() key, combined with one limiter instance
// reused across every auth route, would let one student's login attempts (or
// even just one student clicking "resend OTP") exhaust the shared quota for the
// whole class on every auth endpoint. Keying by the email in the request body
// plus the route path keeps each account's — and each endpoint's — quota
// independent, while still throttling brute-force attempts against one target
// account. A missing/unparsable email (malformed body, or a route with no email
// field) falls back to IP+path so unauthenticated flooding is still capped.
func AuthLimiterKey(c fiber.Ctx) string {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(c.Body(), &body); err == nil && body.Email != "" {
		return strings.ToLower(body.Email) + "|" + c.Path()
	}
	return c.IP() + "|" + c.Path()
}
