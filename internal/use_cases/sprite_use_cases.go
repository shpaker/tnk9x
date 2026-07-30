package use_cases

import (
	"image"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ISpriteUseCases = (*SpriteUseCases)(nil)

// SpriteUseCases выдаёт рендеру изображения спрайтов по типу тайлсета
type SpriteUseCases struct {
	tilesetRegistry interfaces.ITilesetRepositoryRegistry
}

func NewSpriteUseCases(
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
) *SpriteUseCases {
	return &SpriteUseCases{tilesetRegistry: tilesetRegistry}
}

func (uc *SpriteUseCases) GetImage(
	tilesetType types.TilesetType,
	id string,
) (image.Image, error) {
	return uc.tilesetRegistry.GetImageData(tilesetType, id)
}

func (uc *SpriteUseCases) GetImageIDs(
	tilesetType types.TilesetType,
) []string {
	return uc.tilesetRegistry.GetImageIDs(tilesetType)
}
