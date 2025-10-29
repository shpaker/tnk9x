package types

import "errors"

// BulletEntity представляет пулю
type BulletEntity struct {
	ImageGetter IImageIDGetter
	Position    Position
	Speed       float64
	Direction   Direction
	Owner       *TankEntity
	Altitude    Altitude
}

// GetImageID возвращает ID изображения пули (реализует IImageIDGetter)
func (b *BulletEntity) GetImageID() (string, error) {
	if b.ImageGetter == nil {
		return "", errors.New("ImageGetter is nil")
	}
	return b.ImageGetter.GetImageID()
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
