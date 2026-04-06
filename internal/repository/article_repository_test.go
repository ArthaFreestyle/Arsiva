package repository

import "testing"

func TestNewArticleRepository(t *testing.T) {
repo := NewArticleRepository(nil, nil)
if repo == nil {
t.Fatal("expected repository instance")
}
}
