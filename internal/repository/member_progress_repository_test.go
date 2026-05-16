package repository

import "testing"

func TestNewMemberProgressRepository(t *testing.T) {
	repo := NewMemberProgressRepository(nil, nil)
	if repo == nil {
		t.Fatal("expected repository instance")
	}
}
