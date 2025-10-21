package models

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/types"
)

// type TankVersion struct {
// 	AnimationFrames [2]*ebiten.Image
// }
// type TankVariation = [4]TankVersion

type Tank struct {
	// TankVariation
	Image *ebiten.Image
	// Level         uint8
	SpawnPosition types.Position
	WorldPosition types.Position
	Speed         float64
	Direction     types.Direction
}

// GetSize возвращает размер танка
func (t *Tank) GetSize() types.Size {
	return types.Size{Width: 16, Height: 16} // Стандартный размер танка (TankSpriteSize)
}

// GetWorldPosition возвращает позицию танка в мире
func (t *Tank) GetWorldPosition() types.Position {
	return t.WorldPosition
}

// GetScreenPosition возвращает позицию танка на экране (в данном случае совпадает с мировой)
func (t *Tank) GetScreenPosition() types.Position {
	return t.WorldPosition
}
