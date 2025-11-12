package services

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

// TileService предоставляет логику создания тайлов и анимаций
type TileService struct {
	tilesetRegistry  interfaces.ITilesetRepositoryRegistry
	tilesetType      processed.TilesetType
	enemyTilesetType processed.TilesetType // Для специальных случаев, когда нужен fallback на enemy
}

// NewTileService создает новый сервис тайлов
func NewTileService(
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	tilesetType processed.TilesetType,
) *TileService {
	return &TileService{
		tilesetRegistry: tilesetRegistry,
		tilesetType:     tilesetType,
	}
}

// NewTileServiceWithSpecialRepos создает новый сервис тайлов с репозиториями для специальных анимаций
func NewTileServiceWithSpecialRepos(
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	primaryTilesetType processed.TilesetType,
	enemyTilesetType processed.TilesetType,
	spawnerTilesetType processed.TilesetType,
	explosionTilesetType processed.TilesetType,
) *TileService {
	return &TileService{
		tilesetRegistry:  tilesetRegistry,
		tilesetType:      primaryTilesetType,
		enemyTilesetType: enemyTilesetType,
	}
}

// GetTileAnimationFrames возвращает данные анимации по ID
func (s *TileService) GetTileAnimationFrames(
	id string,
) (types.AnimationData, error) {
	// Пробуем получить из основного тайлсета
	animationData, err := s.getAnimationDataFromTileset(s.tilesetType, id)
	if err == nil {
		return animationData, nil
	}

	// Если не найдено и есть fallback на enemy, пробуем его
	if s.enemyTilesetType != "" {
		return s.getAnimationDataFromTileset(s.enemyTilesetType, id)
	}

	return nil, fmt.Errorf("animation '%s' not found", id)
}

// GetAnimationConfig получает конфигурацию анимации по ID
func (s *TileService) GetAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	// Пробуем получить из основного тайлсета
	config, err := s.getAnimationConfigFromTileset(s.tilesetType, id)
	if err == nil {
		return config, nil
	}

	// Если не найдено и есть fallback на enemy, пробуем его
	if s.enemyTilesetType != "" {
		return s.getAnimationConfigFromTileset(s.enemyTilesetType, id)
	}

	return types.AnimationConfig{}, fmt.Errorf(
		"animation config '%s' not found",
		id,
	)
}

// getAnimationDataFromTileset получает данные анимации из указанного тайлсета
func (s *TileService) getAnimationDataFromTileset(
	tilesetType processed.TilesetType,
	id string,
) (types.AnimationData, error) {
	switch tilesetType {
	case processed.TilesetTypeBlocks:
		return s.tilesetRegistry.GetBlocksAnimationData(id)
	case processed.TilesetTypePlayer:
		return s.tilesetRegistry.GetPlayerAnimationData(id)
	case processed.TilesetTypeEnemy:
		return s.tilesetRegistry.GetEnemyAnimationData(id)
	case processed.TilesetTypeBullet:
		return s.tilesetRegistry.GetBulletAnimationData(id)
	case processed.TilesetTypeSpawner:
		return s.tilesetRegistry.GetSpawnerAnimationData(id)
	case processed.TilesetTypeExplosion:
		return s.tilesetRegistry.GetExplosionAnimationData(id)
	case processed.TilesetTypeHQ:
		return s.tilesetRegistry.GetHQAnimationData(id)
	default:
		return nil, fmt.Errorf("unknown tileset type: %s", tilesetType)
	}
}

// getAnimationConfigFromTileset получает конфигурацию анимации из указанного тайлсета
func (s *TileService) getAnimationConfigFromTileset(
	tilesetType processed.TilesetType,
	id string,
) (types.AnimationConfig, error) {
	switch tilesetType {
	case processed.TilesetTypeBlocks:
		return s.tilesetRegistry.GetBlocksAnimationConfig(id)
	case processed.TilesetTypePlayer:
		return s.tilesetRegistry.GetPlayerAnimationConfig(id)
	case processed.TilesetTypeEnemy:
		return s.tilesetRegistry.GetEnemyAnimationConfig(id)
	case processed.TilesetTypeBullet:
		return s.tilesetRegistry.GetBulletAnimationConfig(id)
	case processed.TilesetTypeSpawner:
		return s.tilesetRegistry.GetSpawnerAnimationConfig(id)
	case processed.TilesetTypeExplosion:
		return s.tilesetRegistry.GetExplosionAnimationConfig(id)
	case processed.TilesetTypeHQ:
		return s.tilesetRegistry.GetHQAnimationConfig(id)
	default:
		return types.AnimationConfig{}, fmt.Errorf(
			"unknown tileset type: %s",
			tilesetType,
		)
	}
}

// CreateAnimationFromConfig создает анимацию на основе конфигурации и данных кадров
func (s *TileService) CreateAnimationFromConfig(
	animationFrames types.AnimationData,
	config types.AnimationConfig,
) *image_providers.AnimationProvider {
	hasOffset := s.HasOffset(config.Offset)
	hasRepeats := config.Repeats != nil

	switch {
	case hasRepeats && hasOffset:
		return image_providers.NewAnimationProviderWithLoopsAndOffset(
			animationFrames,
			*config.Repeats,
			config.Offset,
		)
	case hasRepeats:
		return image_providers.NewAnimationProviderWithLoops(
			animationFrames,
			*config.Repeats,
		)
	case hasOffset:
		return image_providers.NewAnimationProviderWithOffset(
			animationFrames,
			config.Offset,
		)
	default:
		return image_providers.NewAnimationProvider(animationFrames)
	}
}

// HasOffset проверяет, есть ли непустое смещение
func (s *TileService) HasOffset(offset [2]float64) bool {
	return offset[0] != 0 || offset[1] != 0
}

// CreateAnimationTileFromTileset создает анимированный тайл из указанного тайлсета
func (s *TileService) CreateAnimationTileFromTileset(
	tilesetType string,
	id string,
) (*image_providers.AnimationProvider, error) {
	config, err := s.getAnimationConfigFromTileset(
		processed.TilesetType(tilesetType),
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("animation config '%s' not found: %w", id, err)
	}

	animationFrames, err := s.getAnimationDataFromTileset(
		processed.TilesetType(tilesetType),
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("animation '%s' not found: %w", id, err)
	}

	return s.CreateAnimationFromConfig(animationFrames, config), nil
}
