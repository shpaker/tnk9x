package services

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
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
) *types.TileAnimationEntity {
	hasOffset := s.HasOffset(config.Offset)
	hasRepeats := config.Repeats != nil

	switch {
	case hasRepeats && hasOffset:
		return types.NewTileAnimationEntityWithLoopsAndOffset(
			animationFrames,
			*config.Repeats,
			config.Offset,
		)
	case hasRepeats:
		return types.NewTileAnimationEntityWithLoops(
			animationFrames,
			*config.Repeats,
		)
	case hasOffset:
		return types.NewTileAnimationEntityWithOffset(
			animationFrames,
			config.Offset,
		)
	default:
		return types.NewTileAnimationEntity(animationFrames)
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
) (*types.TileAnimationEntity, error) {
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
