package usecase

import (
"testing"

"github.com/go-playground/validator/v10"
)

func TestNewPuzzleUseCase(t *testing.T) {
uc := NewPuzzleUseCase(nil, nil, nil, validator.New())
if uc == nil {
t.Fatal("expected usecase instance")
}
}
