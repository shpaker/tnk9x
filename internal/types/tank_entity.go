package types

import "errors"

// TankEntity представляет танк (игрока или врага)
type TankEntity struct {
	AnimationGetter IImageIdGetter
	SpawnPosition   Position
	WorldPosition   Position
	Speed           float64
	Direction       Direction
	IsSpawned       bool    // Флаг спавна танка (по умолчанию false)
	SpawnedAt       float64 // Время спавна танка
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
	return t.WorldPosition
}

// GetSize возвращает размер танка
func (t *TankEntity) GetSize() Size {
	return Size{Width: 16, Height: 16} // Стандартный размер танка
}

// GetWorldPosition возвращает позицию танка в мире
func (t *TankEntity) GetWorldPosition() Position {
	return t.WorldPosition
}
