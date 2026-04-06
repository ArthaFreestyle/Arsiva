package usecase

import (
"testing"

"github.com/go-playground/validator/v10"
)

func TestNewQuizCategoryUseCase(t *testing.T) {
uc := NewQuizCategoryUseCase(nil, nil, validator.New())
if uc == nil {
t.Fatal("expected usecase instance")
}
}
