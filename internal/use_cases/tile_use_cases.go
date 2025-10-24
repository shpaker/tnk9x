package use_cases

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/types"
)

// TileUseCases содержит бизнес-логику для работы с тайлами
type TileUseCases struct {
	tilesRepository types.ITilesetRepository
}

// NewTileUseCases создает новый экземпляр TileUseCases
func NewTileUseCases(tilesRepository types.ITilesetRepository) *TileUseCases {
	return &TileUseCases{
		tilesRepository: tilesRepository,
	}
}

// CreateStaticTile создает статический тайл по ID изображения
func (tuc *TileUseCases) CreateStaticTile(id string) (types.IImageIdGetter, error) {
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
func (tuc *TileUseCases) CreateAnimationTile(id string) (*types.TileAnimationEntity, error) {
	// Получаем данные анимации
	animationFrames, err := tuc.tilesRepository.GetAnimationData(id)
	if err != nil {
		return nil, fmt.Errorf("animation '%s' not found: %w", id, err)
	}

	return types.NewTileAnimationEntity(animationFrames), nil
}

// CreateBlockEntity создает BlockEntity с TileStaticEntity
func (tuc *TileUseCases) CreateBlockEntity(blockType string, positionX, positionY float64) (*types.BlockEntity, error) {
	// Создаем TileStaticEntity для блока
	tileEntity, err := tuc.CreateStaticTile(blockType)
	if err != nil {
		return nil, fmt.Errorf("failed to create tile entity for block type '%s': %w", blockType, err)
	}

	// Создаем BlockEntity
	blockEntity := &types.BlockEntity{
		ImageGetter: tileEntity,
		Data: &types.BlockData{
			Name:     types.BlockType(blockType),
			Position: types.Position{X: positionX, Y: positionY},
		},
		Properties: &types.BlockProperties{
			Collidable: true, // По умолчанию блоки коллизибельны
		},
		WorldPosition: types.Position{X: positionX, Y: positionY},
	}

	return blockEntity, nil
}
