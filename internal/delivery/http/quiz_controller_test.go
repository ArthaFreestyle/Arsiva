package http

import "testing"

func TestNewQuizController(t *testing.T) {
controller := NewQuizController(nil, nil)
if controller == nil {
t.Fatal("expected controller instance")
}
}
