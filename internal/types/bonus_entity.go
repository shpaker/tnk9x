package types

import (
	"errors"
)

type BonusType string

const (
	BonusTypeHelmet  BonusType = "helmet"  // Защита
	BonusTypeTimer   BonusType = "timer"   // Заморозка врагов
	BonusTypeShovel  BonusType = "shovel"  // Укрепление базы
	BonusTypeStar    BonusType = "star"    // Улучшение танка
	BonusTypeGrenade BonusType = "grenade" // Уничтожение всех врагов
	BonusTypeTank    BonusType = "tank"    // Дополнительная жизнь
)

type BonusEntity struct {
	position     Position
	size         Size
	altitude     Altitude
	image        IImageProvider
	bonusType    BonusType
	owner        *TankEntity
	blinkCounter int  // Счетчик тиков для мигания
	blinkFlag    bool // Флаг видимости
}

func (b *BonusEntity) GetImageID() (string, error) {
	if b.image == nil {
		return "", errors.New("image is nil")
	}
	return b.image.GetImageID()
}

func (b *BonusEntity) GetSize() Size {
	if b.size.Width == 0 && b.size.Height == 0 {
		return Size{
			Width:  16,
			Height: 16,
		}
	}
	return b.size
}

func (b *BonusEntity) GetPosition() Position {
	return b.position
}

func (b *BonusEntity) GetAltitude() Altitude {
	return b.altitude
}

func (b *BonusEntity) GetImage() IImageProvider {
	return b.image
}

func (b *BonusEntity) GetType() BonusType {
	return b.bonusType
}

func (b *BonusEntity) GetOwner() *TankEntity {
	return b.owner
}

func (b *BonusEntity) SetOwner(owner *TankEntity) {
	b.owner = owner
}

func (b *BonusEntity) GetBlinkFlag() bool {
	return b.blinkFlag
}

func (b *BonusEntity) UpdateBlink() {
	b.blinkCounter++
	if b.blinkCounter >= 10 {
		b.blinkCounter = 0
		b.blinkFlag = !b.blinkFlag
	}
}

// Реализация интерфейса IBlink
var _ IBlink = (*BonusEntity)(nil)

func NewBonusEntity(
	bonusType BonusType,
	position Position,
	size Size,
	imageGetter IImageProvider,
) *BonusEntity {
	return &BonusEntity{
		position:     position,
		size:         size,
		altitude:     SURFACE,
		image:        imageGetter,
		bonusType:    bonusType,
		owner:        nil,
		blinkCounter: 0,
		blinkFlag:    true, // Начинаем с видимым
	}
}
