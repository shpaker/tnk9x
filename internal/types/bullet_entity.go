package types

import "errors"

// BulletEntity представляет пулю
type BulletEntity struct {
	ImageGetter   IImageIdGetter
	WorldPosition Position
	Speed         float64
	Direction     Direction
	Owner         *TankEntity
}

// GetSize возвращает размер пули
func (b *BulletEntity) GetSize() Size {
	return Size{Width: 4, Height: 4}
}

// GetWorldPosition возвращает позицию пули в мире
func (b *BulletEntity) GetWorldPosition() Position {
	return b.WorldPosition
}

// GetImageId возвращает ID изображения пули (реализует IImageIdGetter)
func (b *BulletEntity) GetImageId() (string, error) {
	if b.ImageGetter == nil {
		return "", errors.New("ImageGetter is nil")
	}
	return b.ImageGetter.GetImageId()
}

// GetScreenPosition возвращает позицию пули на экране
func (b *BulletEntity) GetScreenPosition() Position {
	return b.WorldPosition
}
