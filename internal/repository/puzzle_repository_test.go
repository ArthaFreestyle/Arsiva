package repository

import "testing"

func TestNewPuzzleRepository(t *testing.T) {
repo := NewPuzzleRepository(nil, nil)
if repo == nil {
t.Fatal("expected repository instance")
}
}
