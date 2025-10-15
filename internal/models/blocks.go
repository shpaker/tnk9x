package models

import (
	"github.com/shpaker/gonflict/internal/types"
)

type Block struct {
	Position
	Name types.BlockType
}

type Level []Block
