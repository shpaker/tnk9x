package processed

import (
	"github.com/shpaker/gonflict/internal/types"
)

// Константы для репозиториев
const (
	MapBlocksLength = 26
)

// MapCharsBlocksMapping содержит соответствие символов карты типам блоков
var MapCharsBlocksMapping = map[string]types.BlockType{
	"#": types.Brick,
	"@": types.Steel,
	"%": types.Forest,
	"~": types.Water,
	"=": types.Ice,
}
