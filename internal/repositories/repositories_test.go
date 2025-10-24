package repositories

import (
	"testing"

	"github.com/shpaker/gonflict/internal/repositories/processed"
)

// TestTilesetRepositoryReimport проверяет, что TilesetRepository корректно реимпортирован
func TestTilesetRepositoryReimport(t *testing.T) {
	// Проверяем, что интерфейс доступен
	var _ processed.ITilesetRepository = (*processed.TilesetDataRepository)(nil)

	// Проверяем, что конструктор доступен (функция всегда не nil)
	_ = processed.NewTilesetDataRepository

	t.Log("TilesetRepository успешно реимпортирован")
}
