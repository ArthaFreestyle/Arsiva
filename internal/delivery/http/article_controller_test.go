package http

import "testing"

func TestNewArticleController(t *testing.T) {
controller := NewArticleController(nil, nil)
if controller == nil {
t.Fatal("expected controller instance")
}
}
