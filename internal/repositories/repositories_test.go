package repositories

import (
	"testing"
)

// TestTilesetRepositoryReimport проверяет, что TilesetRepository корректно реимпортирован
func TestTilesetRepositoryReimport(t *testing.T) {
	// Проверяем, что интерфейс доступен
	var _ ITilesetRepository = (*TilesetRepository)(nil)

	// Проверяем, что конструктор доступен
	if NewTilesetRepository == nil {
		t.Fatal("NewTilesetRepository не экспортирован")
	}

	t.Log("TilesetRepository успешно реимпортирован")
}
