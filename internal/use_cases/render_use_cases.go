package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

var _ interfaces.IRenderUseCases = (*RenderUseCases)(nil)

type RenderUseCases struct {
	tilesUseCases interfaces.ITilesUseCases
}

func NewRenderUseCases(
	tilesUseCases interfaces.ITilesUseCases,
) *RenderUseCases {
	return &RenderUseCases{
		tilesUseCases: tilesUseCases,
	}
}

func (uc *RenderUseCases) IsTankSpawnAnimationFinished(
	tank *types.TankEntity,
) bool {
	if tank.Image == nil {
		return false
	}
	if anim, ok := tank.Image.(*image_providers.AnimationProvider); ok {
		return anim.IsFinished()
	}
	return false
}

func (uc *RenderUseCases) IsTankExplosionAnimationFinished(
	tank *types.TankEntity,
) bool {
	if tank.Image == nil {
		return false
	}
	if anim, ok := tank.Image.(*image_providers.AnimationProvider); ok {
		return anim.IsFinished()
	}
	return false
}

func (uc *RenderUseCases) UpdateTankAnimation(
	tank *types.TankEntity,
) {
	if tank == nil {
		return
	}

	animationName := tank.AnimationName()
	tankAnimation, err := uc.tilesUseCases.CreateTankAnimationTile(
		animationName,
		tank.IsEnemy(),
	)
	if err != nil {
		return
	}

	if tank.Image != nil {
		if anim, ok := tank.Image.(*image_providers.AnimationProvider); ok {
			uc.tilesUseCases.StopAnimation(anim)
		}
	}

	tank.Image = tankAnimation
	uc.tilesUseCases.AddAnimation(tankAnimation)

	uc.SyncTankAnimationWithState(tank)
}

func (uc *RenderUseCases) SyncTankAnimationWithState(
	tank *types.TankEntity,
) {
	if tank.Image == nil {
		return
	}
	anim, ok := tank.Image.(*image_providers.AnimationProvider)
	if !ok {
		return
	}

	if tank.State == types.TankStateStopped {
		if anim.IsAnimating {
			uc.tilesUseCases.StopAnimation(anim)
		}
		return
	}

	shouldAnimate := tank.State == types.TankStateMoving ||
		tank.State == types.TankStateBraking

	if shouldAnimate && !anim.IsAnimating {
		uc.tilesUseCases.StartAnimation(anim)
	} else if !shouldAnimate && anim.IsAnimating {
		uc.tilesUseCases.StopAnimation(anim)
	}
}

func (uc *RenderUseCases) UpdateBlink(blinkObjects []types.IBlink) {
	if blinkObjects == nil {
		return
	}

	for _, blinkObj := range blinkObjects {
		if blinkObj != nil {
			blinkObj.UpdateBlink()
		}
	}
}
