package repository

import "testing"

func TestNewUserRepository(t *testing.T) {
	repo := NewUserRepository(nil, nil)
	if repo == nil {
		t.Fatal("expected repository instance")
	}
}

func TestGetDeletedUsersAndRestoreUserInInterface(t *testing.T) {
	// verifies that UserRepositoryImpl satisfies the full interface
	// (compile-time check — if the methods are missing this file won't build)
	var _ UserRepository = &UserRepositoryImpl{}
}
