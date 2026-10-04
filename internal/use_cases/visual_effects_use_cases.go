package use_cases

import (
	"image"
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// Отдача и пыль
const (
	recoilTicks = 3 // сколько тиков танк откачен после выстрела
	// dustChance — пыль из-под гусениц в среднем раз в dustChance тиков
	dustChance = 6
	// barrelReach — расстояние от центра танка до среза ствола
	barrelReach = 8.0
	// trackOffset — половина колеи: пыль идёт из-под двух гусениц
	trackOffset = 5.0
)

// Трассер пули: искры с кормы летящей пули. Быстрая пуля искрит
// чаще, хвост плотнее; шанс — искра в среднем раз в N тиков
const (
	tracerChance     = 2
	fastTracerChance = 1
	// tracerTail — расстояние от центра пули до точки вылета искр
	tracerTail = 2.0
	// baseBulletSpeed — скорость пули без прокачки: быстрее — быстрая
	baseBulletSpeed = 120.0
)

// Тряска экрана: травма от событий, затухание за тик, смещение
// в пикселях при полной травме. Смещение — shakeMaxOffset·травма²
// с округлением до пикселя: травма меньше ~0.4 экран не сдвигает.
// Попадание во врага трясёт ощутимо, в игрока — сильнее
const (
	shakeBulletClash     = 0.5
	shakeEnemyHit        = 0.55
	shakeEnemyExplosion  = 0.65
	shakePlayerHit       = 0.85
	shakePlayerExplosion = 1.0
	shakeHQExplosion     = 1.0
	shakeDecay           = 0.03
	shakeMaxOffset       = 4.0
)

// burstSpec — разлёт частиц одного вида: число, скорость
// в пикселях за тик, половина угла разлёта, жизнь, размер и цвета
type burstSpec struct {
	count              int
	speedMin, speedMax float64
	spread             float64 // math.Pi — во все стороны
	lifeMin, lifeMax   uint
	sizeMin, sizeMax   float64
	drag               float64
	colors             []color.NRGBA
}

// Обломки стены: разрушенная область распадается на куски
// debrisChunk x debrisChunk пикселей цвета её спрайта, которые
// осыпаются из стены навстречу пуле
const (
	debrisChunk    = 2
	debrisSpeedMin = 0.25
	debrisSpeedMax = 1.1
	debrisSpread   = 0.9
	debrisLifeMin  = 18
	debrisLifeMax  = 34
	debrisDrag     = 0.82
)

// flashSpec — вспышка света одного вида; coneCos > 0 — конус
// вдоль направления события
type flashSpec struct {
	radius    float64
	color     color.NRGBA
	intensity float64
	life      uint
	coneCos   float64
}

var (
	muzzleSparks = burstSpec{
		count: 4, speedMin: 1.2, speedMax: 2.2, spread: 0.35,
		lifeMin: 6, lifeMax: 10, sizeMin: 1, sizeMax: 1, drag: 0.8,
		colors: []color.NRGBA{
			{R: 255, G: 240, B: 170, A: 255},
			{R: 255, G: 200, B: 90, A: 255},
		},
	}
	// Дымок из ствола: медленно уходит вперёд и расползается
	muzzleSmoke = burstSpec{
		count: 3, speedMin: 0.1, speedMax: 0.35, spread: 0.5,
		lifeMin: 30, lifeMax: 45, sizeMin: 2, sizeMax: 2, drag: 0.93,
		colors: []color.NRGBA{
			{R: 110, G: 105, B: 100, A: 120},
			{R: 140, G: 135, B: 125, A: 100},
		},
	}
	// Искры трассера обычной пули: тёплые, гаснут за несколько тиков
	tracerSparks = burstSpec{
		count: 1, speedMin: 0.3, speedMax: 0.8, spread: 0.3,
		lifeMin: 6, lifeMax: 12, sizeMin: 1, sizeMax: 1, drag: 0.8,
		colors: []color.NRGBA{
			{R: 255, G: 220, B: 130, A: 255},
			{R: 255, G: 170, B: 70, A: 255},
		},
	}
	// Быстрая пуля: хвост плотнее и длиннее
	fastTracerSparks = burstSpec{
		count: 2, speedMin: 0.3, speedMax: 1, spread: 0.3,
		lifeMin: 8, lifeMax: 16, sizeMin: 1, sizeMax: 1, drag: 0.82,
		colors: []color.NRGBA{
			{R: 255, G: 235, B: 160, A: 255},
			{R: 255, G: 180, B: 80, A: 255},
		},
	}
	// Усиленная пуля, пробивающая сталь: бело-голубые искры
	// сыплются и в стороны
	reinforcedTracerSparks = burstSpec{
		count: 2, speedMin: 0.4, speedMax: 1.3, spread: 0.9,
		lifeMin: 8, lifeMax: 16, sizeMin: 1, sizeMax: 1, drag: 0.82,
		colors: []color.NRGBA{
			{R: 255, G: 255, B: 255, A: 255},
			{R: 190, G: 225, B: 255, A: 255},
			{R: 130, G: 190, B: 255, A: 255},
		},
	}
	brickDust = burstSpec{
		count: 3, speedMin: 0.2, speedMax: 0.5, spread: 1.2,
		lifeMin: 20, lifeMax: 30, sizeMin: 2, sizeMax: 2, drag: 0.9,
		colors: []color.NRGBA{{R: 120, G: 110, B: 100, A: 150}},
	}
	steelSparks = burstSpec{
		count: 8, speedMin: 1.5, speedMax: 3, spread: 1.1,
		lifeMin: 8, lifeMax: 14, sizeMin: 1, sizeMax: 1, drag: 0.82,
		colors: []color.NRGBA{
			{R: 255, G: 250, B: 220, A: 255},
			{R: 255, G: 220, B: 120, A: 255},
			{R: 200, G: 230, B: 255, A: 255},
		},
	}
	shieldSparks = burstSpec{
		count: 8, speedMin: 1, speedMax: 2.4, spread: math.Pi,
		lifeMin: 8, lifeMax: 14, sizeMin: 1, sizeMax: 1, drag: 0.85,
		colors: []color.NRGBA{
			{R: 150, G: 230, B: 255, A: 255},
			{R: 255, G: 255, B: 255, A: 255},
		},
	}
	// Столкновение пуль: искры во все стороны
	clashSparks = burstSpec{
		count: 12, speedMin: 1.2, speedMax: 3, spread: math.Pi,
		lifeMin: 8, lifeMax: 16, sizeMin: 1, sizeMax: 1, drag: 0.84,
		colors: []color.NRGBA{
			{R: 255, G: 250, B: 220, A: 255},
			{R: 255, G: 210, B: 110, A: 255},
			{R: 255, G: 150, B: 60, A: 255},
		},
	}
	embers = burstSpec{
		count: 14, speedMin: 0.8, speedMax: 2.6, spread: math.Pi,
		lifeMin: 18, lifeMax: 34, sizeMin: 1, sizeMax: 2, drag: 0.88,
		colors: []color.NRGBA{
			{R: 255, G: 200, B: 80, A: 255},
			{R: 255, G: 140, B: 40, A: 255},
			{R: 255, G: 90, B: 30, A: 255},
		},
	}
	smoke = burstSpec{
		count: 6, speedMin: 0.15, speedMax: 0.5, spread: math.Pi,
		lifeMin: 40, lifeMax: 64, sizeMin: 2, sizeMax: 3, drag: 0.95,
		colors: []color.NRGBA{
			{R: 70, G: 70, B: 70, A: 170},
			{R: 95, G: 90, B: 85, A: 170},
		},
	}
	trackDust = burstSpec{
		count: 1, speedMin: 0.1, speedMax: 0.3, spread: 0.6,
		lifeMin: 18, lifeMax: 26, sizeMin: 1, sizeMax: 2, drag: 0.9,
		colors: []color.NRGBA{{R: 110, G: 105, B: 95, A: 140}},
	}

	// Круговой отсвет выстрела слабее конуса: светит вокруг танка
	muzzleFlash = flashSpec{
		32, color.NRGBA{R: 255, G: 220, B: 150, A: 255}, 1.1, 4, 0,
	}
	// Конус дульной вспышки вперёд по стволу, узкий и короткий
	muzzleCone = flashSpec{
		56, color.NRGBA{R: 255, G: 230, B: 170, A: 255}, 2, 3, 0.8,
	}
	// Выстрел врага — короткий тусклый проблеск, а не прожектор
	enemyMuzzleFlash = flashSpec{
		22, color.NRGBA{R: 255, G: 200, B: 140, A: 255}, 0.8, 3, 0,
	}
	enemyMuzzleCone = flashSpec{
		36, color.NRGBA{R: 255, G: 210, B: 150, A: 255}, 1.1, 3, 0.8,
	}
	brickFlash = flashSpec{
		18, color.NRGBA{R: 255, G: 170, B: 90, A: 255}, 0.8, 3, 0,
	}
	steelFlash = flashSpec{
		26, color.NRGBA{R: 220, G: 235, B: 255, A: 255}, 1.4, 4, 0,
	}
	clashFlash = flashSpec{
		32, color.NRGBA{R: 255, G: 225, B: 160, A: 255}, 1.6, 5, 0,
	}
	shieldFlash = flashSpec{
		30, color.NRGBA{R: 120, G: 220, B: 255, A: 255}, 1.2, 5, 0,
	}
)

var _ interfaces.IVisualEffectsUseCases = (*VisualEffectsUseCases)(nil)

// VisualEffectsUseCases реализует IVisualEffectsUseCases: превращает
// игровые события в частицы, вспышки, тряску и отдачу
type VisualEffectsUseCases struct {
	// Repositories
	visualEffectsRepository interfaces.IVisualEffectsRepository

	tilesetRegistry interfaces.ITilesetRepositoryRegistry

	// Use Cases
	tankCommonUseCases interfaces.ITankCommonUseCases
	bulletUseCases     interfaces.IBulletUseCases
}

func NewVisualEffectsUseCases(
	visualEffectsRepository interfaces.IVisualEffectsRepository,
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	bulletUseCases interfaces.IBulletUseCases,
) *VisualEffectsUseCases {
	return &VisualEffectsUseCases{
		visualEffectsRepository: visualEffectsRepository,
		tilesetRegistry:         tilesetRegistry,
		tankCommonUseCases:      tankCommonUseCases,
		bulletUseCases:          bulletUseCases,
	}
}

// RequestEffect реализует IVisualEffectsUseCases
func (uc *VisualEffectsUseCases) RequestEffect(
	event types.VisualEventEntity,
) {
	uc.visualEffectsRepository.AddEvent(event)
}

// Update реализует IVisualEffectsUseCases. Отдача отсчитывается
// до разбора событий: выстрел этого тика виден все recoilTicks кадров
func (uc *VisualEffectsUseCases) Update() {
	uc.updateTanks()
	uc.updateParticles()
	uc.updateBullets()
	uc.updateFlashes()
	uc.visualEffectsRepository.GetScreenShake().Decay(shakeDecay)

	for _, event := range uc.visualEffectsRepository.DrainEvents() {
		uc.applyEvent(event)
	}
}

// GetParticles реализует IVisualEffectsUseCases
func (uc *VisualEffectsUseCases) GetParticles() []types.ParticleEntity {
	return uc.visualEffectsRepository.GetParticles()
}

// GetFlashLights реализует IVisualEffectsUseCases: вспышка тускнеет
// и сжимается к концу жизни
func (uc *VisualEffectsUseCases) GetFlashLights() []types.LightEntity {
	flashes := uc.visualEffectsRepository.GetFlashes()
	lights := make([]types.LightEntity, 0, len(flashes))
	for i := range flashes {
		fade := flashes[i].Fade()
		lights = append(lights, types.LightEntity{
			Position:  flashes[i].Position,
			Radius:    flashes[i].Radius * (0.6 + 0.4*fade),
			Color:     flashes[i].Color,
			Intensity: flashes[i].Intensity * fade,
			Direction: flashes[i].Direction,
			ConeCos:   flashes[i].ConeCos,
		})
	}
	return lights
}

// GetShakeOffset реализует IVisualEffectsUseCases: смещение растёт
// как квадрат травмы, направление задаёт сумма синусоид по фазе —
// дрожь без рывков и без генератора случайных чисел в отрисовке
func (uc *VisualEffectsUseCases) GetShakeOffset() types.Position {
	shake := uc.visualEffectsRepository.GetScreenShake()
	trauma := shake.GetTrauma()
	if trauma <= 0 {
		return types.Position{}
	}
	amplitude := shakeMaxOffset * trauma * trauma
	phase := float64(shake.GetTicks())
	return types.Position{
		X: math.Round(amplitude * (0.6*math.Sin(phase*2.1) +
			0.4*math.Sin(phase*5.3+1))),
		Y: math.Round(amplitude * (0.6*math.Sin(phase*1.7+2) +
			0.4*math.Sin(phase*4.7))),
	}
}

// Продвижение эффектов

// updateTanks отсчитывает отдачу и пылит из-под гусениц движущихся танков
func (uc *VisualEffectsUseCases) updateTanks() {
	for _, tank := range uc.tankCommonUseCases.GetAllTanks() {
		if tank == nil {
			continue
		}
		tank.TickRecoil()
		if !tank.IsDriving() || rand.IntN(dustChance) != 0 {
			continue
		}
		direction := tank.Direction.Vector()
		side := trackOffset
		if rand.IntN(2) == 0 {
			side = -side
		}
		center := tankCenter(tank)
		// Задняя кромка танка, левая или правая гусеница
		position := types.Position{
			X: center.X - direction.X*barrelReach - direction.Y*side,
			Y: center.Y - direction.Y*barrelReach + direction.X*side,
		}
		uc.emitBurst(position, tank.Direction.Angle()+math.Pi, trackDust)
	}
}

// updateBullets сыплет искры трассера с кормы летящих пуль;
// вид и плотность искр зависят от прокачки пули
func (uc *VisualEffectsUseCases) updateBullets() {
	for _, bullet := range uc.bulletUseCases.GetBullets() {
		if bullet == nil {
			continue
		}
		spec, chance := tracerSpecOf(bullet)
		if rand.IntN(chance) != 0 {
			continue
		}
		direction := bullet.Direction.Vector()
		size := bullet.GetSize()
		position := types.Position{
			X: bullet.Position.X + float64(size.Width)/2 -
				direction.X*tracerTail,
			Y: bullet.Position.Y + float64(size.Height)/2 -
				direction.Y*tracerTail,
		}
		uc.emitBurst(position, bullet.Direction.Angle()+math.Pi, spec)
	}
}

// tracerSpecOf выбирает искры трассера и их шанс по пуле:
// усиленная искрит голубым, быстрая — чаще и гуще
func tracerSpecOf(bullet *types.BulletEntity) (burstSpec, int) {
	switch {
	case bullet.IsReinforced():
		return reinforcedTracerSparks, fastTracerChance
	case bullet.GetSpeed() > baseBulletSpeed:
		return fastTracerSparks, fastTracerChance
	default:
		return tracerSparks, tracerChance
	}
}

// updateParticles двигает частицы с трением и удаляет погасшие,
// уплотняя срез на месте
func (uc *VisualEffectsUseCases) updateParticles() {
	particles := uc.visualEffectsRepository.GetParticles()
	alive := particles[:0]
	for _, particle := range particles {
		if particle.Life <= 1 {
			continue
		}
		particle.Life--
		particle.Position.X += particle.Velocity.X
		particle.Position.Y += particle.Velocity.Y
		particle.Velocity.X *= particle.Drag
		particle.Velocity.Y *= particle.Drag
		alive = append(alive, particle)
	}
	uc.visualEffectsRepository.SetParticles(alive)
}

// updateFlashes гасит вспышки и удаляет догоревшие
func (uc *VisualEffectsUseCases) updateFlashes() {
	flashes := uc.visualEffectsRepository.GetFlashes()
	alive := flashes[:0]
	for _, flash := range flashes {
		if flash.Life <= 1 {
			continue
		}
		flash.Life--
		alive = append(alive, flash)
	}
	uc.visualEffectsRepository.SetFlashes(alive)
}

// Разбор событий

// applyEvent запускает эффекты события: частицы разлетаются вдоль
// выстрела или назад от стены, в которую попала пуля
func (uc *VisualEffectsUseCases) applyEvent(event types.VisualEventEntity) {
	backward := event.Direction.Angle() + math.Pi
	shake := uc.visualEffectsRepository.GetScreenShake()

	switch event.Kind {
	case types.VisualEventShot:
		position := event.Position
		flash, cone := muzzleFlash, muzzleCone
		if event.Tank != nil {
			if event.Tank.IsEnemy() {
				flash, cone = enemyMuzzleFlash, enemyMuzzleCone
			}
			event.Tank.StartRecoil(recoilTicks)
			direction := event.Tank.Direction.Vector()
			center := tankCenter(event.Tank)
			position = types.Position{
				X: center.X + direction.X*barrelReach,
				Y: center.Y + direction.Y*barrelReach,
			}
		}
		uc.addFlash(position, flash)
		uc.addDirectedFlash(position, event.Direction, cone)
		uc.emitBurst(position, event.Direction.Angle(), muzzleSparks)
		uc.emitBurst(position, event.Direction.Angle(), muzzleSmoke)
	case types.VisualEventBrickHit:
		uc.addFlash(event.Position, brickFlash)
		uc.emitBurst(event.Position, backward, brickDust)
	case types.VisualEventBlockDebris:
		uc.emitDebris(event, backward)
	case types.VisualEventSteelHit:
		uc.addFlash(event.Position, steelFlash)
		uc.emitBurst(event.Position, backward, steelSparks)
	case types.VisualEventBulletClash:
		shake.AddTrauma(shakeBulletClash)
		uc.addFlash(event.Position, clashFlash)
		uc.emitBurst(event.Position, 0, clashSparks)
	case types.VisualEventShieldHit:
		uc.addFlash(event.Position, shieldFlash)
		uc.emitBurst(event.Position, backward, shieldSparks)
	case types.VisualEventPlayerHit:
		shake.AddTrauma(shakePlayerHit)
		uc.emitBurst(event.Position, backward, steelSparks)
	case types.VisualEventEnemyHit:
		shake.AddTrauma(shakeEnemyHit)
		uc.addFlash(event.Position, steelFlash)
		uc.emitBurst(event.Position, backward, steelSparks)
	case types.VisualEventTankExplosion:
		position := event.Position
		trauma := shakeEnemyExplosion
		if event.Tank != nil {
			position = tankCenter(event.Tank)
			if !event.Tank.IsEnemy() {
				trauma = shakePlayerExplosion
			}
		}
		shake.AddTrauma(trauma)
		uc.emitBurst(position, 0, embers)
		uc.emitBurst(position, 0, smoke)
	case types.VisualEventHQExplosion:
		shake.AddTrauma(shakeHQExplosion)
		for range 2 {
			uc.emitBurst(event.Position, 0, embers)
			uc.emitBurst(event.Position, 0, smoke)
		}
	}
}

// emitBurst выпускает частицы из точки в пределах spec.spread
// вокруг угла angle
func (uc *VisualEffectsUseCases) emitBurst(
	position types.Position,
	angle float64,
	spec burstSpec,
) {
	for range spec.count {
		heading := angle + (rand.Float64()*2-1)*spec.spread
		speed := randomBetween(spec.speedMin, spec.speedMax)
		life := spec.lifeMin + uint(rand.IntN(int(spec.lifeMax-spec.lifeMin)+1))
		uc.visualEffectsRepository.AddParticle(types.ParticleEntity{
			Position: position,
			Velocity: types.Position{
				X: math.Cos(heading) * speed,
				Y: math.Sin(heading) * speed,
			},
			Color:   spec.colors[rand.IntN(len(spec.colors))],
			Size:    math.Round(randomBetween(spec.sizeMin, spec.sizeMax)),
			Life:    life,
			MaxLife: life,
			Drag:    spec.drag,
		})
	}
}

// emitDebris разбивает разрушенную область стены на куски: каждый
// берёт цвет своего места в спрайте блока и осыпается под углом
// angle; прозрачные места спрайта обломков не дают
func (uc *VisualEffectsUseCases) emitDebris(
	event types.VisualEventEntity,
	angle float64,
) {
	sprite, origin, ok := uc.blockSprite(event.Block)
	if !ok {
		return
	}
	bounds := sprite.Bounds()
	for y := 0; y < event.Size.Height; y += debrisChunk {
		for x := 0; x < event.Size.Width; x += debrisChunk {
			position := types.Position{
				X: event.Position.X + float64(x),
				Y: event.Position.Y + float64(y),
			}
			// Координаты куска внутри спрайта исходного тайла
			spriteX := bounds.Min.X + int(position.X-origin.X)
			spriteY := bounds.Min.Y + int(position.Y-origin.Y)
			pixel, _ := color.NRGBAModel.Convert(
				sprite.At(spriteX, spriteY),
			).(color.NRGBA)
			if pixel.A == 0 {
				continue
			}
			heading := angle + (rand.Float64()*2-1)*debrisSpread
			speed := randomBetween(debrisSpeedMin, debrisSpeedMax)
			life := uint(
				debrisLifeMin + rand.IntN(debrisLifeMax-debrisLifeMin+1),
			)
			uc.visualEffectsRepository.AddParticle(types.ParticleEntity{
				// Центр куска: частица рисуется квадратом вокруг позиции
				Position: types.Position{
					X: position.X + debrisChunk/2,
					Y: position.Y + debrisChunk/2,
				},
				Velocity: types.Position{
					X: math.Cos(heading) * speed,
					Y: math.Sin(heading) * speed,
				},
				Color:   pixel,
				Size:    debrisChunk,
				Life:    life,
				MaxLife: life,
				Drag:    debrisDrag,
			})
		}
	}
}

// blockSprite возвращает спрайт блока и origin исходного тайла
// в координатах поля: у сколотого остатка спрайт общий с целым тайлом
func (uc *VisualEffectsUseCases) blockSprite(
	block *types.BlockEntity,
) (image.Image, types.Position, bool) {
	if block == nil || block.Data == nil {
		return nil, types.Position{}, false
	}
	imageID, err := block.GetImageID()
	if err != nil {
		return nil, types.Position{}, false
	}
	sprite, err := uc.tilesetRegistry.GetImageData(
		types.TilesetTypeBlocks,
		imageID,
	)
	if err != nil {
		return nil, types.Position{}, false
	}
	return sprite, block.Data.Position, true
}

func (uc *VisualEffectsUseCases) addFlash(
	position types.Position,
	spec flashSpec,
) {
	uc.addDirectedFlash(position, types.DirectionUp, spec)
}

// addDirectedFlash добавляет вспышку; у конусной вспышки ось
// смотрит по direction, у круговой направление не действует
func (uc *VisualEffectsUseCases) addDirectedFlash(
	position types.Position,
	direction types.Direction,
	spec flashSpec,
) {
	flash := types.FlashEntity{
		Position:  position,
		Radius:    spec.radius,
		Color:     spec.color,
		Intensity: spec.intensity,
		Life:      spec.life,
		MaxLife:   spec.life,
	}
	if spec.coneCos > 0 {
		flash.Direction = direction.Vector()
		flash.ConeCos = spec.coneCos
	}
	uc.visualEffectsRepository.AddFlash(flash)
}

func randomBetween(low, high float64) float64 {
	return low + rand.Float64()*(high-low)
}

// tankCenter — центр танка в координатах поля
func tankCenter(tank *types.TankEntity) types.Position {
	return types.Position{
		X: tank.Position.X + float64(tank.Size.Width)/2,
		Y: tank.Position.Y + float64(tank.Size.Height)/2,
	}
}
