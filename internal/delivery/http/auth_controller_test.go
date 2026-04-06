package http

import "testing"

func TestNewAuthController(t *testing.T) {
controller := NewAuthController(nil, nil)
if controller == nil {
t.Fatal("expected controller instance")
}
}
