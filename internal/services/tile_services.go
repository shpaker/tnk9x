package services

import (
	"fmt"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

var _ interfaces.ITileService = (*TileService)(nil)

type TileService struct {
	tilesetRegistry  interfaces.ITilesetRepositoryRegistry
	tilesetType      types.TilesetType
	enemyTilesetType types.TilesetType
}

func NewTileService(
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	tilesetType types.TilesetType,
) *TileService {
	return &TileService{
		tilesetRegistry: tilesetRegistry,
		tilesetType:     tilesetType,
	}
}

func NewTileServiceWithSpecialRepos(
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	primaryTilesetType types.TilesetType,
	enemyTilesetType types.TilesetType,
	spawnerTilesetType types.TilesetType,
	explosionTilesetType types.TilesetType,
) *TileService {
	return &TileService{
		tilesetRegistry:  tilesetRegistry,
		tilesetType:      primaryTilesetType,
		enemyTilesetType: enemyTilesetType,
	}
}

func (s *TileService) GetTileAnimationFrames(
	id string,
) (types.AnimationData, error) {
	animationData, err := s.getAnimationDataFromTileset(s.tilesetType, id)
	if err == nil {
		return animationData, nil
	}

	if s.enemyTilesetType != "" {
		return s.getAnimationDataFromTileset(s.enemyTilesetType, id)
	}

	return nil, fmt.Errorf("animation '%s' not found", id)
}

func (s *TileService) GetAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	config, err := s.getAnimationConfigFromTileset(s.tilesetType, id)
	if err == nil {
		return config, nil
	}

	if s.enemyTilesetType != "" {
		return s.getAnimationConfigFromTileset(s.enemyTilesetType, id)
	}

	return types.AnimationConfig{}, fmt.Errorf(
		"animation config '%s' not found",
		id,
	)
}

func (s *TileService) getAnimationDataFromTileset(
	tilesetType types.TilesetType,
	id string,
) (types.AnimationData, error) {
	switch tilesetType {
	case types.TilesetTypeBlocks:
		return s.tilesetRegistry.GetBlocksAnimationData(id)
	case types.TilesetTypePlayer:
		return s.tilesetRegistry.GetPlayerAnimationData(id)
	case types.TilesetTypeEnemy:
		return s.tilesetRegistry.GetEnemyAnimationData(id)
	case types.TilesetTypeBullet:
		return s.tilesetRegistry.GetBulletAnimationData(id)
	case types.TilesetTypeSpawner:
		return s.tilesetRegistry.GetSpawnerAnimationData(id)
	case types.TilesetTypeExplosion:
		return s.tilesetRegistry.GetExplosionTankAnimationData(id)
	case types.TilesetTypeHQ:
		return s.tilesetRegistry.GetHQAnimationData(id)
	default:
		return nil, fmt.Errorf("unknown tileset type: %s", tilesetType)
	}
}

func (s *TileService) getAnimationConfigFromTileset(
	tilesetType types.TilesetType,
	id string,
) (types.AnimationConfig, error) {
	switch tilesetType {
	case types.TilesetTypeBlocks:
		return s.tilesetRegistry.GetBlocksAnimationConfig(id)
	case types.TilesetTypePlayer:
		return s.tilesetRegistry.GetPlayerAnimationConfig(id)
	case types.TilesetTypeEnemy:
		return s.tilesetRegistry.GetEnemyAnimationConfig(id)
	case types.TilesetTypeBullet:
		return s.tilesetRegistry.GetBulletAnimationConfig(id)
	case types.TilesetTypeSpawner:
		return s.tilesetRegistry.GetSpawnerAnimationConfig(id)
	case types.TilesetTypeExplosion:
		return s.tilesetRegistry.GetExplosionTankAnimationConfig(id)
	case types.TilesetTypeHQ:
		return s.tilesetRegistry.GetHQAnimationConfig(id)
	default:
		return types.AnimationConfig{}, fmt.Errorf(
			"unknown tileset type: %s",
			tilesetType,
		)
	}
}

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

func (s *TileService) HasOffset(offset [2]float64) bool {
	return offset[0] != 0 || offset[1] != 0
}

func (s *TileService) CreateAnimationTileFromTileset(
	tilesetType string,
	id string,
) (*image_providers.AnimationProvider, error) {
	config, err := s.getAnimationConfigFromTileset(
		types.TilesetType(tilesetType),
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("animation config '%s' not found: %w", id, err)
	}

	animationFrames, err := s.getAnimationDataFromTileset(
		types.TilesetType(tilesetType),
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("animation '%s' not found: %w", id, err)
	}

	return s.CreateAnimationFromConfig(animationFrames, config), nil
}
