package http

import "testing"

func TestNewUserController(t *testing.T) {
controller := NewUserController(nil, nil)
if controller == nil {
t.Fatal("expected controller instance")
}
}
