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

// NewTileServiceWithEnemyFallback — сервис, который при отсутствии
// анимации в основном тайлсете ищет её в тайлсете врагов
func NewTileServiceWithEnemyFallback(
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	primaryTilesetType types.TilesetType,
	enemyTilesetType types.TilesetType,
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
	animationData, err := s.tilesetRegistry.GetAnimationData(s.tilesetType, id)
	if err == nil {
		return animationData, nil
	}

	if s.enemyTilesetType != "" {
		return s.tilesetRegistry.GetAnimationData(s.enemyTilesetType, id)
	}

	return nil, fmt.Errorf("animation '%s' not found", id)
}

func (s *TileService) GetAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	config, err := s.tilesetRegistry.GetAnimationConfig(s.tilesetType, id)
	if err == nil {
		return config, nil
	}

	if s.enemyTilesetType != "" {
		return s.tilesetRegistry.GetAnimationConfig(s.enemyTilesetType, id)
	}

	return types.AnimationConfig{}, fmt.Errorf(
		"animation config '%s' not found",
		id,
	)
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
	tilesetType types.TilesetType,
	id string,
) (*image_providers.AnimationProvider, error) {
	config, err := s.tilesetRegistry.GetAnimationConfig(tilesetType, id)
	if err != nil {
		return nil, fmt.Errorf("animation config '%s' not found: %w", id, err)
	}

	animationFrames, err := s.tilesetRegistry.GetAnimationData(tilesetType, id)
	if err != nil {
		return nil, fmt.Errorf("animation '%s' not found: %w", id, err)
	}

	return s.CreateAnimationFromConfig(animationFrames, config), nil
}
