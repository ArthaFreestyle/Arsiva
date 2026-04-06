package http

import "testing"

func TestNewPuzzleController(t *testing.T) {
controller := NewPuzzleController(nil, nil)
if controller == nil {
t.Fatal("expected controller instance")
}
}
