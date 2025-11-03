package services

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

// TileService предоставляет логику создания тайлов и анимаций
type TileService struct {
	tilesRepository      interfaces.ITilesetRepository
	spawnerTilesetRepo   interfaces.ITilesetRepository
	explosionTilesetRepo interfaces.ITilesetRepository
}

// NewTileService создает новый сервис тайлов
func NewTileService(
	tilesRepository interfaces.ITilesetRepository,
) *TileService {
	return &TileService{
		tilesRepository: tilesRepository,
	}
}

// NewTileServiceWithSpecialRepos создает новый сервис тайлов с репозиториями для специальных анимаций
func NewTileServiceWithSpecialRepos(
	tilesRepository interfaces.ITilesetRepository,
	spawnerTilesetRepo interfaces.ITilesetRepository,
	explosionTilesetRepo interfaces.ITilesetRepository,
) *TileService {
	return &TileService{
		tilesRepository:      tilesRepository,
		spawnerTilesetRepo:   spawnerTilesetRepo,
		explosionTilesetRepo: explosionTilesetRepo,
	}
}

// GetTileAnimationFrames возвращает данные анимации по ID
func (s *TileService) GetTileAnimationFrames(
	id string,
) (types.AnimationData, error) {
	return s.tilesRepository.GetAnimationData(id)
}

// GetAnimationConfig получает конфигурацию анимации по ID
func (s *TileService) GetAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	config, err := s.tilesRepository.GetAnimationConfig(id)
	if err != nil {
		return types.AnimationConfig{}, fmt.Errorf(
			"animation config '%s' not found: %w",
			id,
			err,
		)
	}
	return config, nil
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

// CreateAnimationTileFromRepo создает анимированный тайл из указанного репозитория
func (s *TileService) CreateAnimationTileFromRepo(
	repo interfaces.ITilesetRepository,
	id string,
) (*image_providers.AnimationProvider, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository is nil")
	}

	config, err := repo.GetAnimationConfig(id)
	if err != nil {
		return nil, fmt.Errorf("animation config '%s' not found: %w", id, err)
	}

	animationFrames, err := repo.GetAnimationData(id)
	if err != nil {
		return nil, fmt.Errorf("animation '%s' not found: %w", id, err)
	}

	return s.CreateAnimationFromConfig(animationFrames, config), nil
}
