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
// здоровью; каналы не обнулены, чтобы сохранить объём спрайта
var tankHealthTintColors = map[uint]color.NRGBA{
	4: {R: 255, G: 90, B: 90, A: 255},
	3: {R: 255, G: 230, B: 100, A: 255},
	2: {R: 120, G: 255, B: 120, A: 255},
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
// с бонусом; такой танк в этом кадре не отрисовывается. Танк под щитом
// не мигает — поверх него мерцает силовое поле
func (uc *RenderUseCases) IsTankVisible(tank *types.TankEntity) bool {
	if tank == nil {
		return false
	}
	return !(tank.IsEnemy() && tank.GetWithBonus() && !tank.GetBlinkFlag())
}

// IsTankBlinking — true для танка, который мигает: враг с бонусом
// или тяжёлый враг с индикацией здоровья
func (uc *RenderUseCases) IsTankBlinking(tank *types.TankEntity) bool {
	if tank == nil || !tank.IsEnemy() {
		return false
	}
	_, hasTint := healthTintColor(tank)
	return tank.GetWithBonus() || hasTint
}

// BonusPulse реализует IRenderUseCases: пока враг с бонусом виден,
// от него за фазу мигания расходится одно кольцо
func (uc *RenderUseCases) BonusPulse(tank *types.TankEntity) (float64, bool) {
	if tank == nil || !tank.IsEnemy() || !tank.GetWithBonus() ||
		!tank.IsActive() || !uc.IsTankVisible(tank) {
		return 0, false
	}
	return tank.GetBlinkProgress(), true
}

// TankHealthTint возвращает тон спрайта тяжёлого танка в текущем кадре;
// ok=false — спрайт рисуется без тона. Как в NES, танк чередует обычный
// и тонированный спрайт; танк с бонусом и так пропадает в выключенной
// фазе мигания, поэтому видимым он всегда тонирован
func (uc *RenderUseCases) TankHealthTint(
	tank *types.TankEntity,
) (color.NRGBA, bool) {
	tintColor, exists := healthTintColor(tank)
	if !exists || !(tank.GetWithBonus() || tank.GetBlinkFlag()) {
		return color.NRGBA{}, false
	}
	return tintColor, true
}

// healthTintColor — цвет тона по здоровью тяжёлого вражеского танка
func healthTintColor(tank *types.TankEntity) (color.NRGBA, bool) {
	if tank == nil || !tank.IsEnemy() || tank.GetSpecs() == nil ||
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
