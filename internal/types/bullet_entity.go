package types

import "errors"

// BulletEntity представляет пулю
type BulletEntity struct {
	ImageGetter IImageIdGetter
	Position    Position
	Speed       float64
	Direction   Direction
	Owner       *TankEntity
	Altitude    Altitude
}

// GetImageId возвращает ID изображения пули (реализует IImageIdGetter)
func (b *BulletEntity) GetImageId() (string, error) {
	if b.ImageGetter == nil {
		return "", errors.New("ImageGetter is nil")
	}
	return b.ImageGetter.GetImageId()
}

// GetSize возвращает размер пули
func (b *BulletEntity) GetSize() Size {
	return Size{Width: 4, Height: 4}
}

// GetPosition возвращает позицию пули в мире
func (b *BulletEntity) GetPosition() Position {
	return b.Position
}

// GetAltitude возвращает высоту пули
func (b *BulletEntity) GetAltitude() Altitude {
	return b.Altitude
}
