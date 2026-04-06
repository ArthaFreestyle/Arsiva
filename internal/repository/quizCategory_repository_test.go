package repository

import "testing"

func TestNewQuizCategoryRepository(t *testing.T) {
repo := NewQuizCategoryRepository(nil, nil)
if repo == nil {
t.Fatal("expected repository instance")
}
}
