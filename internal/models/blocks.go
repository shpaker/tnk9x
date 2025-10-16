package models

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/types"
)

type BlockData struct {
	Name types.BlockType
	Position
}

type BlockProperties struct {
	Collidable bool
}

type Block struct {
	Image         *ebiten.Image
	Data          *BlockData
	Properties    *BlockProperties
	WorldPosition Position
}

type Level []Block
