package repository

import "testing"

func TestNewUserRepository(t *testing.T) {
repo := NewUserRepository(nil, nil)
if repo == nil {
t.Fatal("expected repository instance")
}
}
