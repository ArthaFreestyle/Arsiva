package http

import "testing"

func TestNewCeritaController(t *testing.T) {
controller := NewCeritaController(nil, nil)
if controller == nil {
t.Fatal("expected controller instance")
}
}
