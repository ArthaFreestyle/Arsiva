package usecase

import (
"testing"

"github.com/go-playground/validator/v10"
)

func TestNewAuthUseCase(t *testing.T) {
uc := NewAuthUseCase(nil, []byte("secret"), validator.New(), nil, nil)
if uc == nil {
t.Fatal("expected usecase instance")
}
}
