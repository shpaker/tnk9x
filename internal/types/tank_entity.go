package types

import "errors"

// TankState представляет состояние танка
type TankState int

const (
	TankStateSpawning  TankState = iota // Танк спавнится
	TankStateMoving                     // Танк движется
	TankStateStopped                    // Танк остановлен
	TankStateExploding                  // Танк взрывается
)

// TankEntity представляет танк (игрока или врага)
type TankEntity struct {
	AnimationGetter IImageIdGetter
	Position        Position
	Speed           float64
	Direction       Direction
	State           TankState // Состояние танка
	SpawnedAt       float64   // Время спавна танка
	Altitude        Altitude
}

// GetImageId возвращает ID изображения танка (реализует IImageIdGetter)
func (t *TankEntity) GetImageId() (string, error) {
	if t.AnimationGetter == nil {
		return "", errors.New("ImageGetter is nil")
	}
	return t.AnimationGetter.GetImageId()
}

// GetScreenPosition возвращает позицию танка на экране
func (t *TankEntity) GetScreenPosition() Position {
	return t.Position
}

// GetSize возвращает размер танка
func (t *TankEntity) GetSize() Size {
	return Size{Width: 16, Height: 16} // Стандартный размер танка
}

// GetPosition возвращает позицию танка в мире
func (t *TankEntity) GetPosition() Position {
	return t.Position
}

// GetAltitude возвращает высоту танка
func (t *TankEntity) GetAltitude() Altitude {
	// Если танк взрывается, показываем выше всего
	if t.State == TankStateExploding {
		return AIR
	}
	return t.Altitude
}
