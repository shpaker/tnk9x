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

// type TankVersion struct {
// 	AnimationFrames [2]*ebiten.Image
// }
// type TankVariation = [4]TankVersion

type Tank struct {
	// TankVariation
	Image *ebiten.Image
	// Level         uint8
	SpawnPosition Position
	WorldPosition Position
	Speed         float64
	Direction     types.Direction
}

type Level []Block

type Bullet struct {
	Image         *ebiten.Image
	WorldPosition Position
	Speed         float64
	Direction     types.Direction
	Owner         *Tank
}
