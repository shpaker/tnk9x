package types

import "github.com/hajimehoshi/ebiten/v2"

// BulletEntity представляет пулю
type BulletEntity struct {
	Image         *ebiten.Image
	WorldPosition Position
	Speed         float64
	Direction     Direction
	Owner         *TankEntity
}

// GetSize возвращает размер пули
func (b *BulletEntity) GetSize() Size {
	return Size{Width: 4, Height: 4}
}

// GetWorldPosition возвращает позицию пули в мире
func (b *BulletEntity) GetWorldPosition() Position {
	return b.WorldPosition
}

// GetScreenPosition возвращает позицию пули на экране
func (b *BulletEntity) GetScreenPosition() Position {
	return b.WorldPosition
}
