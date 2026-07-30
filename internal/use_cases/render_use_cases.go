package use_cases

import (
	"image/color"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

// heavyTankLevel — уровень тяжёлого танка, у которого запас здоровья
// показывается цветным слоем
const heavyTankLevel = 3

// tankHealthOverlayColors — цвет индикации по оставшемуся здоровью
var tankHealthOverlayColors = map[uint]color.NRGBA{
	4: {R: 255, G: 0, B: 0, A: 128},
	3: {R: 255, G: 255, B: 0, A: 128},
	2: {R: 0, G: 255, B: 0, A: 128},
}

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

// IsTankVisible — false в выключенной фазе мигания вражеского танка
// с бонусом; такой танк в этом кадре не отрисовывается
func (uc *RenderUseCases) IsTankVisible(tank *types.TankEntity) bool {
	if tank == nil {
		return false
	}
	return !(tank.IsEnemy() && tank.GetWithBonus() && !tank.GetBlinkFlag())
}

// TankHealthOverlay возвращает цвет слоя здоровья тяжёлого танка;
// ok=false — слой не отрисовывается
func (uc *RenderUseCases) TankHealthOverlay(
	tank *types.TankEntity,
) (color.NRGBA, bool) {
	if tank == nil || !tank.IsEnemy() || tank.GetSpecs() == nil ||
		tank.GetSpecs().GetLevel() != heavyTankLevel {
		return color.NRGBA{}, false
	}

	overlayColor, exists := tankHealthOverlayColors[tank.GetHitPoints()]
	return overlayColor, exists
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
