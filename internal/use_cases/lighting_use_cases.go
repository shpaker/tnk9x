package use_cases

import (
	"image/color"
	"math"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

// Приоритеты источников: при переполнении types.MaxLights отбрасываются
// менее важные — сначала свечение врагов, в последнюю очередь фары
// и аура игроков
const (
	lightPriorityPlayer = iota
	lightPriorityExplosion
	lightPriorityFlash
	lightPriorityBullet
	lightPriorityEnemy
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

	// Пуля — трассер: светится сама и освещает коридор, по которому
	// летит, открывая то, что дальше фары
	bulletLight = lightSpec{
		44,
		color.NRGBA{R: 255, G: 230, B: 150, A: 255},
		1.6,
	}
	// Усиленная пуля, пробивающая сталь, светится бело-голубым
	reinforcedBulletLight = lightSpec{
		48,
		color.NRGBA{R: 200, G: 230, B: 255, A: 255},
		1.8,
	}
	// Силовое поле вспыхивает ярко, с пересветом для bloom
	shieldLight = lightSpec{
		36,
		color.NRGBA{R: 120, G: 220, B: 255, A: 255},
		1.5,
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
	// Свет игрока: дальний конус фары по стволу и мягкая аура рядом
	headlight = lightSpec{
		145,
		color.NRGBA{R: 255, G: 238, B: 205, A: 255},
		1.9,
	}
	playerAura = lightSpec{
		30,
		color.NRGBA{R: 255, G: 235, B: 200, A: 255},
		0.7,
	}
	// Фара врага короче и тусклее, тёплого красноватого света:
	// враг отличим от игрока издалека
	enemyHeadlight = lightSpec{
		120,
		color.NRGBA{R: 255, G: 150, B: 100, A: 255},
		1.4,
	}
	// Корпус врага слабо тлеет красным
	enemyLight = lightSpec{14, color.NRGBA{R: 255, G: 90, B: 70, A: 255}, 0.25}
	hqLight    = lightSpec{36, color.NRGBA{R: 255, G: 210, B: 90, A: 255}, 0.6}
)

// Параметры фары: половина угла конуса у игрока и врага, вынос
// источника к срезу ствола и доля угла, доворачиваемая за тик
// (~6 тиков на поворот)
const (
	headlightHalfAngle      = 36 * math.Pi / 180
	enemyHeadlightHalfAngle = 30 * math.Pi / 180
	headlightReach          = 6.0
	headlightTurnRate       = 0.35
)

// surfaceMaterials — свойства поверхностей: кирпич и сталь бросают
// тень, лес затеняет частично, вода и лёд дают блики. Сталь — металл,
// её освещённый фасад блестит; кирпич матовый
var surfaceMaterials = map[types.BlockType]types.SurfaceMaterial{
	types.Brick:  {Opacity: 1},
	types.Steel:  {Opacity: 1, Reflectivity: 1},
	types.Forest: {Opacity: 0.35},
	types.Water:  {Reflectivity: 1, Ripple: 1},
	types.Ice:    {Reflectivity: 0.6},
}

var _ interfaces.ILightingUseCases = (*LightingUseCases)(nil)

// LightingUseCases реализует ILightingUseCases: собирает источники
// света из сущностей уровня
type LightingUseCases struct {
	// Use Cases
	tankCommonUseCases    interfaces.ITankCommonUseCases
	bulletUseCases        interfaces.IBulletUseCases
	hqUseCases            interfaces.IHQUseCases
	bonusUseCases         interfaces.IBonusUseCases
	visualEffectsUseCases interfaces.IVisualEffectsUseCases
}

func NewLightingUseCases(
	tankCommonUseCases interfaces.ITankCommonUseCases,
	bulletUseCases interfaces.IBulletUseCases,
	hqUseCases interfaces.IHQUseCases,
	bonusUseCases interfaces.IBonusUseCases,
	visualEffectsUseCases interfaces.IVisualEffectsUseCases,
) *LightingUseCases {
	return &LightingUseCases{
		tankCommonUseCases:    tankCommonUseCases,
		bulletUseCases:        bulletUseCases,
		hqUseCases:            hqUseCases,
		bonusUseCases:         bonusUseCases,
		visualEffectsUseCases: visualEffectsUseCases,
	}
}

// GetLights реализует ILightingUseCases
func (uc *LightingUseCases) GetLights() []types.LightEntity {
	var buckets [lightPriorityCount][]types.LightEntity

	for _, tank := range uc.tankCommonUseCases.GetAllTanks() {
		if tank == nil {
			continue
		}
		if light, priority, ok := headlightOf(tank); ok {
			buckets[priority] = append(buckets[priority], light)
		}
		light, priority, ok := tankLight(tank)
		if ok {
			buckets[priority] = append(buckets[priority], light)
		}
	}

	buckets[lightPriorityFlash] = append(
		buckets[lightPriorityFlash],
		uc.visualEffectsUseCases.GetFlashLights()...,
	)

	for _, bullet := range uc.bulletUseCases.GetBullets() {
		if bullet == nil {
			continue
		}
		buckets[lightPriorityBullet] = append(
			buckets[lightPriorityBullet],
			newLight(bullet.Position, bullet.GetSize(), bulletLightOf(bullet)),
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

// UpdateHeadlights реализует ILightingUseCases
func (uc *LightingUseCases) UpdateHeadlights() {
	for _, tank := range uc.tankCommonUseCases.GetAllTanks() {
		if hasHeadlight(tank) {
			tank.TurnHeadlight(headlightTurnRate)
		}
	}
}

// GetMaterial реализует ILightingUseCases; у неизвестного блока
// нулевой материал — свет проходит без блика
func (uc *LightingUseCases) GetMaterial(
	blockType types.BlockType,
) types.SurfaceMaterial {
	return surfaceMaterials[blockType]
}

// hasHeadlight сообщает, светит ли танк фарой: любой активный танк
func hasHeadlight(tank *types.TankEntity) bool {
	return tank != nil && tank.IsActive()
}

// headlightOf — конус фары у среза ствола по текущему углу фары:
// у игрока длинный и яркий, у врага короче и красноватый
func headlightOf(tank *types.TankEntity) (types.LightEntity, int, bool) {
	if !hasHeadlight(tank) {
		return types.LightEntity{}, 0, false
	}
	spec, halfAngle, priority := headlight, headlightHalfAngle, lightPriorityPlayer
	if tank.IsEnemy() {
		spec, halfAngle, priority =
			enemyHeadlight, enemyHeadlightHalfAngle, lightPriorityEnemy
	}
	angle := tank.GetHeadlightAngle()
	direction := types.Position{X: math.Cos(angle), Y: math.Sin(angle)}
	light := newLight(tank.Position, tank.Size, spec)
	light.Position.X += direction.X * headlightReach
	light.Position.Y += direction.Y * headlightReach
	light.Direction = direction
	light.ConeCos = math.Cos(halfAngle)
	return light, priority, true
}

// tankLight — собственный свет танка по его состоянию: у игрока аура
// обзора, щит заменяет собственный свет, уничтоженный танк не светит
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
		return shieldLightOf(tank), lightPriorityAccent, true
	case tank.IsEnemy():
		return newLight(tank.Position, tank.Size, enemyLight),
			lightPriorityAmbient, true
	default:
		return newLight(tank.Position, tank.Size, playerAura),
			lightPriorityPlayer, true
	}
}

// Тусклая фаза мерцания щита: свечение почти гаснет и сжимается —
// поле заметно пульсирует
const (
	shieldFlickerDim    = 0.35
	shieldFlickerShrink = 0.8
)

// shieldLightOf — свечение неуязвимого танка мерцает в такт кадрам
// силового поля
func shieldLightOf(tank *types.TankEntity) types.LightEntity {
	light := newLight(tank.Position, tank.Size, shieldLight)
	if tank.GetShieldPhase() == 1 {
		light.Intensity *= shieldFlickerDim
		light.Radius *= shieldFlickerShrink
	}
	return light
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

// bulletLightOf — свет пули по её прокачке
func bulletLightOf(bullet *types.BulletEntity) lightSpec {
	if bullet.IsReinforced() {
		return reinforcedBulletLight
	}
	return bulletLight
}
