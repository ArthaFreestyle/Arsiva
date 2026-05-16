package http

import "testing"

func TestNewLeaderboardController(t *testing.T) {
	controller := NewLeaderboardController(nil, nil)
	if controller == nil {
		t.Fatal("expected controller instance")
	}
}
