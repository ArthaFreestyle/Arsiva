package http

import "testing"

func TestNewUserController(t *testing.T) {
	controller := NewUserController(nil, nil)
	if controller == nil {
		t.Fatal("expected controller instance")
	}
}

func TestUserControllerImplementsGetDeletedUsersAndRestoreUser(t *testing.T) {
	// compile-time check: UserControllerImpl must satisfy UserController
	var _ UserController = &UserControllerImpl{}
}
