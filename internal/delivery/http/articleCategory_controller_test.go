package http

import "testing"

func TestNewArticleCategoryController(t *testing.T) {
controller := NewArticleCategoryController(nil, nil)
if controller == nil {
t.Fatal("expected controller instance")
}
}
