package types

import "errors"

// BlockEntity представляет блок карты
type BlockEntity struct {
	Position Position
	Size     Size
	Altitude Altitude
	Image    IImageProvider
	Data     *BlockData
}

// BlockData содержит данные блока
type BlockData struct {
	Name     BlockType
	Position Position
}

// GetImageID возвращает ID изображения блока
func (b *BlockEntity) GetImageID() (string, error) {
	if b.Image == nil {
		return "", errors.New("image is nil")
	}
	return b.Image.GetImageID()
}

// GetSize возвращает размер блока
func (b *BlockEntity) GetSize() Size {
	if b.Size.Width == 0 && b.Size.Height == 0 {
		return Size{
			Width:  8,
			Height: 8,
		} // Стандартный размер блока (TileBaseSize = base_size_px/2)
	}
	return b.Size
}

// GetPosition возвращает позицию блока в мире
func (b *BlockEntity) GetPosition() Position {
	return b.Position
}

// GetAltitude возвращает высоту блока
func (b *BlockEntity) GetAltitude() Altitude {
	return b.Altitude
}

// NewBlockEntity создает новый BlockEntity с указанными параметрами
func NewBlockEntity(
	blockType string,
	positionX,
	positionY float64,
	imageGetter IImageProvider,
) *BlockEntity {
	altitude := SURFACE // По умолчанию блоки на уровне поверхности
	// Деревья (Forest) рисуются выше игрока
	if blockType == string(Forest) {
		altitude = AIR
	}

	return &BlockEntity{
		Position: Position{X: positionX, Y: positionY},
		Size:     Size{Width: 8, Height: 8},
		Altitude: altitude,
		Image:    imageGetter,
		Data: &BlockData{
			Name:     BlockType(blockType),
			Position: Position{X: positionX, Y: positionY},
		},
	}
}
