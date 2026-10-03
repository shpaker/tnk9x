package use_cases

import (
	"image/color"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

// heavyTankLevel — уровень тяжёлого танка, у которого запас здоровья
// показывается миганием спрайта цветным тоном
const heavyTankLevel = 3

// tankHealthTintColors — мультипликатор цвета спрайта по оставшемуся
// здоровью; приглушённые тона не выбиваются из палитры уровня
var tankHealthTintColors = map[uint]color.NRGBA{
	4: {R: 240, G: 150, B: 140, A: 255},
	3: {R: 235, G: 215, B: 150, A: 255},
	2: {R: 165, G: 220, B: 160, A: 255},
}

// bonusCarrierTint — приглушённый красный тон мигающего врага с бонусом
var bonusCarrierTint = color.NRGBA{R: 235, G: 125, B: 115, A: 255}

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

// IsTankBlinking — true для танка, который мигает: враг с бонусом
// или тяжёлый враг с индикацией здоровья
func (uc *RenderUseCases) IsTankBlinking(tank *types.TankEntity) bool {
	_, ok := blinkTintColor(tank)
	return ok
}

// TankTint возвращает тон спрайта танка в текущем кадре; ok=false —
// спрайт рисуется без тона. Мигающий враг чередует обычный
// и тонированный спрайт: враг с бонусом — красный, тяжёлый — цвет брони
func (uc *RenderUseCases) TankTint(
	tank *types.TankEntity,
) (color.NRGBA, bool) {
	tintColor, exists := blinkTintColor(tank)
	if !exists || !tank.GetBlinkFlag() {
		return color.NRGBA{}, false
	}
	return tintColor, true
}

// blinkTintColor — цвет мигания вражеского танка без учёта фазы:
// бонус важнее брони, поэтому тяжёлый враг с бонусом мигает красным
func blinkTintColor(tank *types.TankEntity) (color.NRGBA, bool) {
	if tank == nil || !tank.IsEnemy() {
		return color.NRGBA{}, false
	}
	if tank.GetWithBonus() {
		return bonusCarrierTint, true
	}
	if tank.GetSpecs() == nil ||
		tank.GetSpecs().GetLevel() != heavyTankLevel {
		return color.NRGBA{}, false
	}
	tintColor, exists := tankHealthTintColors[tank.GetHitPoints()]
	return tintColor, exists
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
