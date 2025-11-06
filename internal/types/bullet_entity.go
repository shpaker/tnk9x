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
	owner     *TankEntity
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

// GetOwner возвращает владельца пули
func (b *BulletEntity) GetOwner() *TankEntity {
	return b.owner
}

// NewBulletEntity создает новую пулю
func NewBulletEntity(
	position Position,
	size Size,
	altitude Altitude,
	image IImageProvider,
	speed float64,
	direction Direction,
	owner *TankEntity,
) *BulletEntity {
	return &BulletEntity{
		Position:  position,
		Size:      size,
		Altitude:  altitude,
		Image:     image,
		Speed:     speed,
		Direction: direction,
		owner:     owner,
	}
}
