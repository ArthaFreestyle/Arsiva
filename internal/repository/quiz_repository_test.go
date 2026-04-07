package repository

import "testing"

func TestNewQuizRepository(t *testing.T) {
repo := NewQuizRepository(nil, nil)
if repo == nil {
t.Fatal("expected repository instance")
}
}
