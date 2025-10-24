package use_cases

import (
	"fmt"
	"image"

	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
)

// TilesUseCases содержит бизнес-логику для работы с тайлами
type TilesUseCases struct {
	tilesRepository processed.ITilesetRepository
}

// NewTilesUseCases создает новый экземпляр TilesUseCases
func NewTilesUseCases(
	tilesRepository processed.ITilesetRepository,
) *TilesUseCases {
	return &TilesUseCases{
		tilesRepository: tilesRepository,
	}
}

// GetImage возвращает изображение по ID
func (tuc *TilesUseCases) GetImage(id string) (image.Image, error) {
	return tuc.tilesRepository.GetImage(id)
}

// GetTileAnimationFrames возвращает данные анимации по ID
func (tuc *TilesUseCases) GetTileAnimationFrames(id string) (types.AnimationData, error) {
	return tuc.tilesRepository.GetAnimationData(id)
}

// CreateStaticTile создает статический тайл по ID изображения
func (tuc *TilesUseCases) CreateStaticTile(
	id string,
) (types.IImageIdGetter, error) {
	// Проверяем, что изображение существует
	_, err := tuc.tilesRepository.GetImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}

	return &types.TileStaticEntity{
		ImageId: id,
	}, nil
}

// CreateAnimationTile создает анимированный тайл по ID анимации
func (tuc *TilesUseCases) CreateAnimationTile(id string) (*types.TileAnimationEntity, error) {
	// Получаем данные анимации
	animationFrames, err := tuc.tilesRepository.GetAnimationData(id)
	if err != nil {
		return nil, fmt.Errorf("animation '%s' not found: %w", id, err)
	}

	return types.NewTileAnimationEntity(animationFrames), nil
}
