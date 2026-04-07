package repository

import "testing"

func TestNewCeritaRepository(t *testing.T) {
repo := NewCeritaRepository(nil, nil)
if repo == nil {
t.Fatal("expected repository instance")
}
}
