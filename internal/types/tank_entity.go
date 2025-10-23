package types

import "github.com/hajimehoshi/ebiten/v2"

// TankEntity представляет танк (игрока или врага)
type TankEntity struct {
	Image         *ebiten.Image
	SpawnPosition Position
	WorldPosition Position
	Speed         float64
	Direction     Direction
}

// GetSize возвращает размер танка
func (t *TankEntity) GetSize() Size {
	return Size{Width: 16, Height: 16} // Стандартный размер танка
}

// GetWorldPosition возвращает позицию танка в мире
func (t *TankEntity) GetWorldPosition() Position {
	return t.WorldPosition
}

// GetScreenPosition возвращает позицию танка на экране
func (t *TankEntity) GetScreenPosition() Position {
	return t.WorldPosition
}
