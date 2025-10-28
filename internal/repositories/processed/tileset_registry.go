package processed

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/repositories/raw"
)

// TilesetRepositoryRegistry содержит все репозитории тайлсетов
type TilesetRepositoryRegistry struct {
	blocks    ITilesetRepository
	player    ITilesetRepository
	bullet    ITilesetRepository
	spawner   ITilesetRepository
	explosion ITilesetRepository
}

// NewTilesetRepositoryRegistry создает новый реестр тайлсетов
func NewTilesetRepositoryRegistry(fileRepo raw.IFileRepository) (*TilesetRepositoryRegistry, error) {
	// Создаем репозиторий для блоков
	blocksRepo, err := NewTilesetDataRepository(fileRepo, "tiles/blocks")
	if err != nil {
		return nil, fmt.Errorf("failed to create blocks tileset: %w", err)
	}

	// Создаем репозиторий для игрока
	playerRepo, err := NewTilesetDataRepository(fileRepo, "tiles/player")
	if err != nil {
		return nil, fmt.Errorf("failed to create player tileset: %w", err)
	}

	// Создаем репозиторий для пуль
	bulletRepo, err := NewTilesetDataRepository(fileRepo, "tiles/bullet")
	if err != nil {
		return nil, fmt.Errorf("failed to create bullet tileset: %w", err)
	}

	// Создаем репозиторий для спавна
	spawnerRepo, err := NewTilesetDataRepository(fileRepo, "tiles/spawner")
	if err != nil {
		return nil, fmt.Errorf("failed to create spawner tileset: %w", err)
	}

	// Создаем репозиторий для взрыва
	explosionRepo, err := NewTilesetDataRepository(fileRepo, "tiles/explosion")
	if err != nil {
		return nil, fmt.Errorf("failed to create explosion tileset: %w", err)
	}

	return &TilesetRepositoryRegistry{
		blocks:    blocksRepo,
		player:    playerRepo,
		bullet:    bulletRepo,
		spawner:   spawnerRepo,
		explosion: explosionRepo,
	}, nil
}

// Blocks возвращает репозиторий тайлсетов для блоков
func (tr *TilesetRepositoryRegistry) Blocks() ITilesetRepository {
	return tr.blocks
}

// Player возвращает репозиторий тайлсетов для игрока
func (tr *TilesetRepositoryRegistry) Player() ITilesetRepository {
	return tr.player
}

// Bullet возвращает репозиторий тайлсетов для пуль
func (tr *TilesetRepositoryRegistry) Bullet() ITilesetRepository {
	return tr.bullet
}

// Spawner возвращает репозиторий тайлсетов для спавна
func (tr *TilesetRepositoryRegistry) Spawner() ITilesetRepository {
	return tr.spawner
}

// Explosion возвращает репозиторий тайлсетов для взрыва
func (tr *TilesetRepositoryRegistry) Explosion() ITilesetRepository {
	return tr.explosion
}
