package repository

import "testing"

func TestNewLeaderboardRepository(t *testing.T) {
	repo := NewLeaderboardRepository(nil, nil)
	if repo == nil {
		t.Fatal("expected repository instance")
	}
}
