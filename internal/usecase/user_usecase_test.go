package usecase

import (
"testing"

"github.com/go-playground/validator/v10"
)

func TestNewUserUseCase(t *testing.T) {
uc := NewUserUseCase(nil, nil, nil, validator.New())
if uc == nil {
t.Fatal("expected usecase instance")
}
}
