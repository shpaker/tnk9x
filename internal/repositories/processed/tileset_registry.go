package processed

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/interfaces"
)

// TilesetRepositoryRegistry содержит все репозитории тайлсетов
type TilesetRepositoryRegistry struct {
	blocks    interfaces.ITilesetRepository
	player    interfaces.ITilesetRepository
	enemy     interfaces.ITilesetRepository
	bullet    interfaces.ITilesetRepository
	spawner   interfaces.ITilesetRepository
	explosion interfaces.ITilesetRepository
	hq        interfaces.ITilesetRepository
}

// NewTilesetRepositoryRegistry создает новый реестр тайлсетов
func NewTilesetRepositoryRegistry(
	fileRepo interfaces.IFileRepository,
) (*TilesetRepositoryRegistry, error) {
	// Создаем репозиторий для блоков
	blocksRepo, err := NewTilesetDataRepository(fileRepo, "tiles/blocks")
	if err != nil {
		return nil, fmt.Errorf("failed to create blocks tileset: %w", err)
	}

	// Создаем репозиторий для танков игроков
	playerRepo, err := NewTilesetDataRepository(fileRepo, "tiles/tanks_players")
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create tanks_players tileset: %w",
			err,
		)
	}

	// Создаем репозиторий для танков врагов
	enemyRepo, err := NewTilesetDataRepository(fileRepo, "tiles/tanks_enemies")
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create tanks_enemies tileset: %w",
			err,
		)
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

	// Создаем репозиторий для базы
	hqRepo, err := NewTilesetDataRepository(fileRepo, "tiles/hq")
	if err != nil {
		return nil, fmt.Errorf("failed to create hq tileset: %w", err)
	}

	return &TilesetRepositoryRegistry{
		blocks:    blocksRepo,
		player:    playerRepo,
		enemy:     enemyRepo,
		bullet:    bulletRepo,
		spawner:   spawnerRepo,
		explosion: explosionRepo,
		hq:        hqRepo,
	}, nil
}

// Blocks возвращает репозиторий тайлсетов для блоков
func (tr *TilesetRepositoryRegistry) Blocks() interfaces.ITilesetRepository {
	return tr.blocks
}

// Player возвращает репозиторий тайлсетов для игрока
func (tr *TilesetRepositoryRegistry) Player() interfaces.ITilesetRepository {
	return tr.player
}

// Enemy возвращает репозиторий тайлсетов для врагов
func (tr *TilesetRepositoryRegistry) Enemy() interfaces.ITilesetRepository {
	return tr.enemy
}

// Bullet возвращает репозиторий тайлсетов для пуль
func (tr *TilesetRepositoryRegistry) Bullet() interfaces.ITilesetRepository {
	return tr.bullet
}

// Spawner возвращает репозиторий тайлсетов для спавна
func (tr *TilesetRepositoryRegistry) Spawner() interfaces.ITilesetRepository {
	return tr.spawner
}

// Explosion возвращает репозиторий тайлсетов для взрыва
func (tr *TilesetRepositoryRegistry) Explosion() interfaces.ITilesetRepository {
	return tr.explosion
}

// HQ возвращает репозиторий тайлсетов для базы
func (tr *TilesetRepositoryRegistry) HQ() interfaces.ITilesetRepository {
	return tr.hq
}
