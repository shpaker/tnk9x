package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

var _ interfaces.IHQUseCases = (*HQUseCases)(nil)

type HQUseCases struct {
	tilesUseCases interfaces.ITilesUseCases
	hq            *types.HQEntity
}

func NewHQUseCases(
	tilesUseCases interfaces.ITilesUseCases,
	hq *types.HQEntity,
) *HQUseCases {
	return &HQUseCases{
		tilesUseCases: tilesUseCases,
		hq:            hq,
	}
}

func (uc *HQUseCases) Explode(hq *types.HQEntity) error {
	if hq == nil || hq.State == types.HQStateExploding ||
		hq.IsDestroyed() {
		return nil
	}

	explosionAnim, err := uc.tilesUseCases.CreateExplosionAnimation()
	if err != nil {
		return err
	}

	hq.Image = explosionAnim
	hq.State = types.HQStateExploding

	uc.tilesUseCases.StartAnimation(explosionAnim)
	return nil
}

func (uc *HQUseCases) IsExplosionAnimationFinished(hq *types.HQEntity) bool {
	if hq == nil || hq.Image == nil {
		return true
	}

	if tileAnim, ok := hq.Image.(*image_providers.AnimationProvider); ok {
		return tileAnim.IsFinished()
	}

	return true
}

func (uc *HQUseCases) IsExplosionFinished(hq *types.HQEntity) {
	if hq != nil && hq.State == types.HQStateExploding {
		if uc.IsExplosionAnimationFinished(hq) {

			destroyedImage, err := uc.tilesUseCases.CreateStaticTile(
				"hq_destroyed",
			)
			if err == nil {
				hq.Image = destroyedImage
			}

			hq.State = types.HQStateDestroyed
			hq.Altitude = types.SURFACE
		}
	}
}

func (uc *HQUseCases) GetHQ() *types.HQEntity {
	return uc.hq
}

func (uc *HQUseCases) IsDestroyed() bool {
	if uc == nil || uc.hq == nil {
		return false
	}

	return uc.hq.IsDestroyed()
}
