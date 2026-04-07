package usecase

import (
"testing"

"github.com/go-playground/validator/v10"
)

func TestNewQuizUseCase(t *testing.T) {
uc := NewQuizUseCase(nil, nil, nil, validator.New())
if uc == nil {
t.Fatal("expected usecase instance")
}
}
