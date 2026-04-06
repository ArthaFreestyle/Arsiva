package repository

import "testing"

func TestNewArticleCategoryRepository(t *testing.T) {
repo := NewArticleCategoryRepository(nil, nil)
if repo == nil {
t.Fatal("expected repository instance")
}
}
