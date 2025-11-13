package tank_use_cases

import (
	"github.com/shpaker/tnk25/internal/types"
	image_providers "github.com/shpaker/tnk25/internal/types/image_providers"
	"github.com/shpaker/tnk25/internal/use_cases"
)

type TankRenderUseCases struct {
	tilesUseCases *use_cases.TilesUseCases
}

func NewTankRenderUseCases(
	tilesUseCases *use_cases.TilesUseCases,
) *TankRenderUseCases {
	return &TankRenderUseCases{
		tilesUseCases: tilesUseCases,
	}
}

func (uc *TankRenderUseCases) IsSpawnAnimationFinished(
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

func (uc *TankRenderUseCases) IsExplosionAnimationFinished(
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

func (uc *TankRenderUseCases) UpdateTankAnimation(
	tank *types.TankEntity,
) {
	if uc.tilesUseCases == nil || tank == nil {
		return
	}

	animationName := tank.GetTankAnimationName()
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

	uc.SyncAnimationWithState(tank)
}

func (uc *TankRenderUseCases) SyncAnimationWithState(
	tank *types.TankEntity,
) {
	if uc.tilesUseCases == nil {
		return
	}

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
