package http

import "testing"

func TestNewUploadController(t *testing.T) {
controller := NewUploadController(nil, "./uploads", nil)
if controller == nil {
t.Fatal("expected controller instance")
}
}
