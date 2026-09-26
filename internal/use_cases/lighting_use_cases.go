package use_cases

import (
	"image/color"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

// Приоритеты источников: при переполнении types.MaxLights отбрасываются
// менее важные — сначала танки, в последнюю очередь взрывы
const (
	lightPriorityExplosion = iota
	lightPriorityBullet
	lightPriorityAccent
	lightPriorityAmbient
	lightPriorityCount
)

// Параметры взрыва: радиус растёт по кадрам анимации, яркость
// мерцает на нечётных кадрах; взрыв штаба крупнее танкового
const (
	explosionBaseRadius     = 40.0
	explosionRadiusGrowth   = 32.0
	explosionIntensity      = 1.8
	explosionFlickerFactor  = 0.8
	hqExplosionRadiusFactor = 1.5
)

// lightSpec — радиус, цвет и яркость источника одного вида
type lightSpec struct {
	radius    float64
	color     color.NRGBA
	intensity float64
}

var (
	explosionLightColor = color.NRGBA{R: 255, G: 170, B: 60, A: 255}

	bulletLight = lightSpec{
		28,
		color.NRGBA{R: 255, G: 230, B: 150, A: 255},
		1.3,
	}
	shieldLight = lightSpec{
		32,
		color.NRGBA{R: 120, G: 220, B: 255, A: 255},
		1.0,
	}
	bonusLight = lightSpec{
		28,
		color.NRGBA{R: 255, G: 240, B: 200, A: 255},
		1.0,
	}
	spawnLight = lightSpec{
		30,
		color.NRGBA{R: 220, G: 230, B: 255, A: 255},
		0.8,
	}
	playerLight = lightSpec{
		48,
		color.NRGBA{R: 255, G: 235, B: 200, A: 255},
		0.8,
	}
	enemyLight = lightSpec{24, color.NRGBA{R: 255, G: 90, B: 70, A: 255}, 0.6}
	hqLight    = lightSpec{36, color.NRGBA{R: 255, G: 210, B: 90, A: 255}, 0.6}
)

// surfaceMaterials — свойства поверхностей: кирпич и сталь бросают
// тень, лес затеняет частично, вода и лёд дают блики
var surfaceMaterials = map[types.BlockType]types.SurfaceMaterial{
	types.Brick:  {Opacity: 1},
	types.Steel:  {Opacity: 1, Reflectivity: 0.3},
	types.Forest: {Opacity: 0.35},
	types.Water:  {Reflectivity: 1, Ripple: 1},
	types.Ice:    {Reflectivity: 0.6},
}

var _ interfaces.ILightingUseCases = (*LightingUseCases)(nil)

// LightingUseCases реализует ILightingUseCases: собирает источники
// света из сущностей уровня
type LightingUseCases struct {
	// Use Cases
	tankCommonUseCases interfaces.ITankCommonUseCases
	bulletUseCases     interfaces.IBulletUseCases
	hqUseCases         interfaces.IHQUseCases
	bonusUseCases      interfaces.IBonusUseCases
}

func NewLightingUseCases(
	tankCommonUseCases interfaces.ITankCommonUseCases,
	bulletUseCases interfaces.IBulletUseCases,
	hqUseCases interfaces.IHQUseCases,
	bonusUseCases interfaces.IBonusUseCases,
) *LightingUseCases {
	return &LightingUseCases{
		tankCommonUseCases: tankCommonUseCases,
		bulletUseCases:     bulletUseCases,
		hqUseCases:         hqUseCases,
		bonusUseCases:      bonusUseCases,
	}
}

// GetLights реализует ILightingUseCases
func (uc *LightingUseCases) GetLights() []types.LightEntity {
	var buckets [lightPriorityCount][]types.LightEntity

	for _, tank := range uc.tankCommonUseCases.GetAllTanks() {
		if tank == nil {
			continue
		}
		light, priority, ok := tankLight(tank)
		if ok {
			buckets[priority] = append(buckets[priority], light)
		}
	}

	for _, bullet := range uc.bulletUseCases.GetBullets() {
		if bullet == nil {
			continue
		}
		buckets[lightPriorityBullet] = append(
			buckets[lightPriorityBullet],
			newLight(bullet.Position, bullet.GetSize(), bulletLight),
		)
	}

	for _, bonus := range uc.bonusUseCases.VisibleBonuses() {
		buckets[lightPriorityAccent] = append(
			buckets[lightPriorityAccent],
			newLight(bonus.GetPosition(), bonus.GetSize(), bonusLight),
		)
	}

	if light, priority, ok := hqLightOf(uc.hqUseCases.GetHQ()); ok {
		buckets[priority] = append(buckets[priority], light)
	}

	lights := make([]types.LightEntity, 0, types.MaxLights)
	for _, bucket := range buckets {
		for _, light := range bucket {
			if len(lights) == types.MaxLights {
				return lights
			}
			lights = append(lights, light)
		}
	}
	return lights
}

// GetMaterial реализует ILightingUseCases; у неизвестного блока
// нулевой материал — свет проходит без блика
func (uc *LightingUseCases) GetMaterial(
	blockType types.BlockType,
) types.SurfaceMaterial {
	return surfaceMaterials[blockType]
}

// tankLight — свет танка по его состоянию; щит заменяет собственный
// свет танка, уничтоженный танк не светит
func tankLight(tank *types.TankEntity) (types.LightEntity, int, bool) {
	switch {
	case tank.State == types.TankStateExploded:
		return types.LightEntity{}, 0, false
	case tank.State == types.TankStateExploding:
		return explosionLight(tank.Image, tank.Position, tank.Size, 1),
			lightPriorityExplosion, true
	case tank.State == types.TankStateSpawning:
		return newLight(tank.Position, tank.Size, spawnLight),
			lightPriorityAccent, true
	case tank.HasShield():
		return newLight(tank.Position, tank.Size, shieldLight),
			lightPriorityAccent, true
	case tank.IsEnemy():
		return newLight(tank.Position, tank.Size, enemyLight),
			lightPriorityAmbient, true
	default:
		return newLight(tank.Position, tank.Size, playerLight),
			lightPriorityAmbient, true
	}
}

// hqLightOf — свет штаба: целый слабо светится, взрывающийся даёт
// крупную вспышку, разрушенный не светит
func hqLightOf(hq *types.HQEntity) (types.LightEntity, int, bool) {
	if hq == nil {
		return types.LightEntity{}, 0, false
	}
	switch hq.State {
	case types.HQStateIntact:
		return newLight(hq.Position, hq.GetSize(), hqLight),
			lightPriorityAmbient, true
	case types.HQStateExploding:
		return explosionLight(
				hq.Image,
				hq.Position,
				hq.GetSize(),
				hqExplosionRadiusFactor,
			),
			lightPriorityExplosion, true
	default:
		return types.LightEntity{}, 0, false
	}
}

// explosionLight — вспышка взрыва: радиус растёт к концу анимации,
// нечётные кадры тусклее — пламя мерцает
func explosionLight(
	image types.IImageProvider,
	position types.Position,
	size types.Size,
	radiusFactor float64,
) types.LightEntity {
	progress := 0.0
	intensity := explosionIntensity
	if anim, ok := image.(*image_providers.AnimationProvider); ok &&
		len(anim.AnimationFrames) > 0 {
		progress = float64(anim.CurrentFrame+1) /
			float64(len(anim.AnimationFrames))
		if anim.CurrentFrame%2 == 1 {
			intensity *= explosionFlickerFactor
		}
	}

	return newLight(position, size, lightSpec{
		radius: (explosionBaseRadius + explosionRadiusGrowth*progress) *
			radiusFactor,
		color:     explosionLightColor,
		intensity: intensity,
	})
}

// newLight размещает источник в центре сущности
func newLight(
	position types.Position,
	size types.Size,
	spec lightSpec,
) types.LightEntity {
	return types.LightEntity{
		Position: types.Position{
			X: position.X + float64(size.Width)/2,
			Y: position.Y + float64(size.Height)/2,
		},
		Radius:    spec.radius,
		Color:     spec.color,
		Intensity: spec.intensity,
	}
}
