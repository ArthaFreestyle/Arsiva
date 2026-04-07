package usecase

import (
"testing"

"github.com/go-playground/validator/v10"
)

func TestNewCeritaUseCase(t *testing.T) {
uc := NewCeritaUseCase(nil, nil, nil, validator.New())
if uc == nil {
t.Fatal("expected usecase instance")
}
}
