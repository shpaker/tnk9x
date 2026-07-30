package processed

import (
	"sort"

	"github.com/shpaker/tnk9x/internal/types"
)

// RequiredSprites перечисляет спрайты блоков, на которые ссылаются
// карты уровней: статичные блоки рисуются одноимёнными изображениями,
// вода — анимацией
func RequiredSprites() types.SpriteManifest {
	var blockImageIDs []string
	for _, blockType := range MapCharsBlocksMapping {
		if blockType == types.Water {
			continue
		}
		blockImageIDs = append(blockImageIDs, string(blockType))
	}
	sort.Strings(blockImageIDs)

	return types.SpriteManifest{
		Images: map[types.TilesetType][]string{
			types.TilesetTypeBlocks: blockImageIDs,
		},
		Animations: map[types.TilesetType][]string{
			types.TilesetTypeBlocks: {string(types.Water)},
		},
	}
}
