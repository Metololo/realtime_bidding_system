package repo

import "testing"

func TestNewInMemoryDevilFruitRepositoryContainsAllDevilFruits(t *testing.T) {
	repository := NewInMemoryDevilFruitRepository()

	devilFruits := repository.List()

	if len(devilFruits) == 0 {
		t.Fatalf("expected devil fruits to be set in the rpository, got 0")
	}
}
