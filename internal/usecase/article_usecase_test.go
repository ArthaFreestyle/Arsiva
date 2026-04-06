package usecase

import (
"testing"

"github.com/go-playground/validator/v10"
)

func TestNewArticleUseCase(t *testing.T) {
uc := NewArticleUseCase(nil, nil, nil, validator.New())
if uc == nil {
t.Fatal("expected usecase instance")
}
}
