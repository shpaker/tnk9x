package types

import "errors"

type BulletEntity struct {
	Position  Position
	Size      Size
	Altitude  Altitude
	Image     IImageProvider
	Speed     float64
	Direction Direction
	owner     *TankEntity
}

func (b *BulletEntity) GetImageID() (string, error) {
	if b.Image == nil {
		return "", errors.New("image is nil")
	}
	return b.Image.GetImageID()
}

func (b *BulletEntity) GetSize() Size {
	if b.Size.Width == 0 && b.Size.Height == 0 {
		return Size{Width: 4, Height: 4}
	}
	return b.Size
}

func (b *BulletEntity) GetPosition() Position {
	return b.Position
}

func (b *BulletEntity) GetAltitude() Altitude {
	return b.Altitude
}

func (b *BulletEntity) GetOwner() *TankEntity {
	return b.owner
}

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
