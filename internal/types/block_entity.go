package types

import "errors"

// BlockEntity представляет блок карты
type BlockEntity struct {
	ImageGetter IImageIdGetter
	Data        *BlockData
	Properties  *BlockProperties
	Position    Position
	Altitude    Altitude
}

// BlockData содержит данные блока
type BlockData struct {
	Name     BlockType
	Position Position
}

// BlockProperties содержит свойства блока
type BlockProperties struct {
}

// GetImageId возвращает ID изображения блока
func (b *BlockEntity) GetImageId() (string, error) {
	if b.ImageGetter == nil {
		return "", errors.New("ImageGetter is nil")
	}
	return b.ImageGetter.GetImageId()
}

// GetScreenPosition возвращает позицию блока на экране
func (b *BlockEntity) GetScreenPosition() Position {
	return b.Position
}

// GetSize возвращает размер блока
func (b *BlockEntity) GetSize() Size {
	return Size{Width: 8, Height: 8} // Стандартный размер блока (TileMinSize)
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
	imageGetter IImageIdGetter,
) *BlockEntity {
	altitude := SURFACE // По умолчанию блоки на уровне поверхности
	// Деревья (Forest) рисуются выше игрока
	if blockType == string(Forest) {
		altitude = AIR
	}

	return &BlockEntity{
		ImageGetter: imageGetter,
		Data: &BlockData{
			Name:     BlockType(blockType),
			Position: Position{X: positionX, Y: positionY},
		},
		Properties: &BlockProperties{},
		Position:   Position{X: positionX, Y: positionY},
		Altitude:   altitude,
	}
}
