package http

import "testing"

func TestNewQuizCategoryController(t *testing.T) {
controller := NewQuizCategoryController(nil, nil)
if controller == nil {
t.Fatal("expected controller instance")
}
}
