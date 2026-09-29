package types

import "errors"

type BulletEntity struct {
	Position  Position
	Size      Size
	Altitude  Altitude
	Image     IImageProvider
	Direction Direction
	specs     *SpecsEntity // Спецификации танка, из которого выпущена пуля
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

func (b *BulletEntity) GetSpecs() *SpecsEntity {
	if b == nil {
		return nil
	}
	return b.specs
}

func (b *BulletEntity) IsReinforced() bool {
	if b == nil || b.specs == nil {
		return false
	}
	return b.specs.GetBulletsReinforced()
}

func (b *BulletEntity) GetSpeed() float64 {
	if b == nil || b.specs == nil {
		return 120.0 // Значение по умолчанию
	}
	return b.specs.GetBulletsSpeed()
}

func NewBulletEntity(
	position Position,
	size Size,
	altitude Altitude,
	image IImageProvider,
	direction Direction,
	specs *SpecsEntity,
	owner *TankEntity,
) *BulletEntity {
	return &BulletEntity{
		Position:  position,
		Size:      size,
		Altitude:  altitude,
		Image:     image,
		Direction: direction,
		specs:     specs,
		owner:     owner,
	}
}
