package models

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/types"
)

type BlockData struct {
	Name     types.BlockType
	Position types.Position
}

type BlockProperties struct {
	Collidable bool
}

type Block struct {
	Image         *ebiten.Image
	Data          *BlockData
	Properties    *BlockProperties
	WorldPosition types.Position
}

// GetSize возвращает размер блока
func (b *Block) GetSize() types.Size {
	return types.Size{Width: 8, Height: 8} // Стандартный размер блока (TileMinSize)
}

// GetWorldPosition возвращает позицию блока в мире
func (b *Block) GetWorldPosition() types.Position {
	return b.WorldPosition
}

// GetScreenPosition возвращает позицию блока на экране (в данном случае совпадает с мировой)
func (b *Block) GetScreenPosition() types.Position {
	return b.WorldPosition
}
