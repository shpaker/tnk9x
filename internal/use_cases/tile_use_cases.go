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
	// Получаем конфигурацию анимации
	config, err := tuc.tilesRepository.GetAnimationConfig(id)
	if err != nil {
		return nil, fmt.Errorf("animation config '%s' not found: %w", id, err)
	}

	// Получаем данные анимации
	animationFrames, err := tuc.tilesRepository.GetAnimationData(id)
	if err != nil {
		return nil, fmt.Errorf("animation '%s' not found: %w", id, err)
	}

	// Создаем анимацию с учетом конфигурации
	var animation *types.TileAnimationEntity

	// Проверяем, есть ли offset в конфиге
	hasOffset := config.Offset[0] != 0 || config.Offset[1] != 0

	if config.Repeats == nil {
		// Бесконечная анимация
		if hasOffset {
			animation = types.NewTileAnimationEntityWithOffset(animationFrames, config.Offset)
		} else {
			animation = types.NewTileAnimationEntity(animationFrames)
		}
	} else {
		// Анимация с ограниченным количеством повторений
		if hasOffset {
			animation = types.NewTileAnimationEntityWithLoopsAndOffset(animationFrames, *config.Repeats, config.Offset)
		} else {
			animation = types.NewTileAnimationEntityWithLoops(animationFrames, *config.Repeats)
		}
	}

	return animation, nil
}
