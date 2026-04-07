package repository

import "testing"

func TestNewAssetRepository(t *testing.T) {
repo := NewAssetRepository(nil, nil)
if repo == nil {
t.Fatal("expected repository instance")
}
}
