package types

import "errors"

// BlockEntity представляет блок карты
type BlockEntity struct {
	ImageGetter   IImageIdGetter
	Data          *BlockData
	Properties    *BlockProperties
	WorldPosition Position
}

// BlockData содержит данные блока
type BlockData struct {
	Name     BlockType
	Position Position
}

// BlockProperties содержит свойства блока
type BlockProperties struct {
	Collidable bool
}

// GetSize возвращает размер блока
func (b *BlockEntity) GetSize() Size {
	return Size{Width: 8, Height: 8} // Стандартный размер блока (TileMinSize)
}

// GetWorldPosition возвращает позицию блока в мире
func (b *BlockEntity) GetWorldPosition() Position {
	return b.WorldPosition
}

// GetScreenPosition возвращает позицию блока на экране
func (b *BlockEntity) GetScreenPosition() Position {
	return b.WorldPosition
}

// GetImageId возвращает ID изображения блока
func (b *BlockEntity) GetImageId() (string, error) {
	if b.ImageGetter == nil {
		return "", errors.New("ImageGetter is nil")
	}
	return b.ImageGetter.GetImageId()
}

// NewBlockEntity создает новый BlockEntity с указанными параметрами
func NewBlockEntity(
	blockType string,
	positionX,
	positionY float64,
	imageGetter IImageIdGetter,
) *BlockEntity {
	return &BlockEntity{
		ImageGetter: imageGetter,
		Data: &BlockData{
			Name:     BlockType(blockType),
			Position: Position{X: positionX, Y: positionY},
		},
		Properties: &BlockProperties{
			Collidable: true, // По умолчанию блоки коллизибельны
		},
		WorldPosition: Position{X: positionX, Y: positionY},
	}
}
