package http

import "testing"

func TestNewProgressController(t *testing.T) {
	controller := NewProgressController(nil, nil)
	if controller == nil {
		t.Fatal("expected controller instance")
	}
}
