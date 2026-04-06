package usecase

import "testing"

func TestNewAssetUsecase(t *testing.T) {
uc := NewAssetUsecase(nil, nil, "./uploads")
if uc == nil {
t.Fatal("expected usecase instance")
}
}
