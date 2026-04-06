package usecase

import (
"testing"

"github.com/go-playground/validator/v10"
)

func TestNewArticleCategoryUseCase(t *testing.T) {
uc := NewArticleCategoryUseCase(nil, nil, validator.New())
if uc == nil {
t.Fatal("expected usecase instance")
}
}
