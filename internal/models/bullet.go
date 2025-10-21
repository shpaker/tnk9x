package models

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/types"
)

type Bullet struct {
	Image         *ebiten.Image
	WorldPosition types.Position
	Speed         float64
	Direction     types.Direction
	Owner         *Tank
}

// GetSize возвращает размер пули
func (b *Bullet) GetSize() types.Size {
	return types.Size{Width: 4, Height: 4} // Размер пули
}

// GetWorldPosition возвращает позицию пули в мире
func (b *Bullet) GetWorldPosition() types.Position {
	return b.WorldPosition
}

// GetScreenPosition возвращает позицию пули на экране (в данном случае совпадает с мировой)
func (b *Bullet) GetScreenPosition() types.Position {
	return b.WorldPosition
}
