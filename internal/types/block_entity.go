package types

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// IImageIdGetter определяет интерфейс для получения ID изображения
type IImageIdGetter interface {
	GetImageId() string
}

// ITilesetRepository определяет интерфейс для работы с тайлсетами
type ITilesetRepository interface {
	// GetImage возвращает изображение по ID из тайлсета
	GetImage(id string) (*ebiten.Image, error)
	// GetAnimationData возвращает данные анимации по ID
	GetAnimationData(id string) (AnimationData, error)
}

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
func (b *BlockEntity) GetImageId() string {
	if b.ImageGetter == nil {
		return ""
	}
	return b.ImageGetter.GetImageId()
}

// GetImage возвращает изображение блока из репозитория
func (b *BlockEntity) GetImage(repo ITilesetRepository) (*ebiten.Image, error) {
	if b.ImageGetter == nil {
		return nil, fmt.Errorf("no image getter available")
	}

	imageId := b.ImageGetter.GetImageId()
	if imageId == "" {
		return nil, fmt.Errorf("empty image ID")
	}

	return repo.GetImage(imageId)
}

// SetImageGetter устанавливает новый ImageGetter
func (b *BlockEntity) SetImageGetter(imageGetter IImageIdGetter) {
	b.ImageGetter = imageGetter
}
