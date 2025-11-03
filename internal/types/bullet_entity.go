package types

import "errors"

// BulletEntity представляет пулю
type BulletEntity struct {
	Position  Position
	Size      Size
	Altitude  Altitude
	Image     IImageProvider
	Speed     float64
	Direction Direction
	Owner     *TankEntity
}

// GetImageID возвращает ID изображения пули (реализует IImageProvider)
func (b *BulletEntity) GetImageID() (string, error) {
	if b.Image == nil {
		return "", errors.New("image is nil")
	}
	return b.Image.GetImageID()
}

// GetSize возвращает размер пули
func (b *BulletEntity) GetSize() Size {
	if b.Size.Width == 0 && b.Size.Height == 0 {
		return Size{Width: 4, Height: 4}
	}
	return b.Size
}

// GetPosition возвращает позицию пули в мире
func (b *BulletEntity) GetPosition() Position {
	return b.Position
}

// GetAltitude возвращает высоту пули
func (b *BulletEntity) GetAltitude() Altitude {
	return b.Altitude
}
