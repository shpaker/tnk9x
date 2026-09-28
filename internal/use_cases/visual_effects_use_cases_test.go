package use_cases_test

import (
	"image"
	"image/color"
	"math"
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

type visualEffectsTestEnv struct {
	tankCommon *recordingTankCommon
	bullets    *stubBulletList
	tilesets   *testutil.FakeTilesetRegistry
	repository *game.VisualEffectsRepository
	effects    *use_cases.VisualEffectsUseCases
}

func newVisualEffectsTestEnv() *visualEffectsTestEnv {
	tankCommon := &recordingTankCommon{}
	bullets := &stubBulletList{}
	tilesets := &testutil.FakeTilesetRegistry{}
	repository := game.NewVisualEffectsRepository()
	return &visualEffectsTestEnv{
		tankCommon: tankCommon,
		bullets:    bullets,
		tilesets:   tilesets,
		repository: repository,
		effects: use_cases.NewVisualEffectsUseCases(
			repository,
			tilesets,
			tankCommon,
			bullets,
		),
	}
}

// Событие разбирается на Update: до него частиц и вспышек нет
func TestVisualEffectsUseCases_SteelHitSparksAndFlash(t *testing.T) {
	env := newVisualEffectsTestEnv()
	env.effects.RequestEffect(types.VisualEventEntity{
		Kind:      types.VisualEventSteelHit,
		Position:  types.Position{X: 50, Y: 50},
		Direction: types.DirectionUp,
	})
	if len(env.effects.GetParticles()) != 0 {
		t.Fatal("частицы появились до Update")
	}

	env.effects.Update()

	if len(env.effects.GetParticles()) == 0 {
		t.Error("удар по стали без искр")
	}
	flashes := env.effects.GetFlashLights()
	if len(flashes) != 1 ||
		flashes[0].Position != (types.Position{X: 50, Y: 50}) {
		t.Errorf("вспышки %v, ожидалась одна в точке удара", flashes)
	}
}

// Частицы летят, гаснут и удаляются; вспышка догорает
func TestVisualEffectsUseCases_ParticlesExpire(t *testing.T) {
	env := newVisualEffectsTestEnv()
	env.effects.RequestEffect(types.VisualEventEntity{
		Kind:     types.VisualEventBrickHit,
		Position: types.Position{X: 50, Y: 50},
	})
	env.effects.Update()
	start := env.effects.GetParticles()[0].Position

	env.effects.Update()
	if env.effects.GetParticles()[0].Position == start {
		t.Error("частица не сдвинулась")
	}

	for range 100 {
		env.effects.Update()
	}
	if got := len(env.effects.GetParticles()); got != 0 {
		t.Errorf("осталось частиц %d, ожидалось 0", got)
	}
	if got := len(env.effects.GetFlashLights()); got != 0 {
		t.Errorf("осталось вспышек %d, ожидалось 0", got)
	}
}

// Взрыв трясёт экран, тряска затухает до нуля
func TestVisualEffectsUseCases_ShakeDecays(t *testing.T) {
	env := newVisualEffectsTestEnv()
	if env.effects.GetShakeOffset() != (types.Position{}) {
		t.Fatal("тряска без событий")
	}

	env.effects.RequestEffect(types.VisualEventEntity{
		Kind:     types.VisualEventHQExplosion,
		Position: types.Position{X: 104, Y: 200},
	})
	env.effects.Update()

	shaken := false
	for range 5 {
		if env.effects.GetShakeOffset() != (types.Position{}) {
			shaken = true
		}
		env.effects.Update()
	}
	if !shaken {
		t.Error("взрыв штаба не трясёт экран")
	}

	for range 60 {
		env.effects.Update()
	}
	if offset := env.effects.GetShakeOffset(); offset != (types.Position{}) {
		t.Errorf("тряска не затухла: %v", offset)
	}
}

// Выстрел откатывает танк на несколько тиков и даёт вспышку у ствола
func TestVisualEffectsUseCases_ShotRecoil(t *testing.T) {
	env := newVisualEffectsTestEnv()
	tank := newPlayerTank(types.TankRolePlayer1)
	env.tankCommon.tanks = []*types.TankEntity{tank}

	env.effects.RequestEffect(types.VisualEventEntity{
		Kind:      types.VisualEventShot,
		Direction: tank.Direction,
		Tank:      tank,
	})
	env.effects.Update()

	if !tank.IsRecoiling() {
		t.Fatal("выстрел без отдачи")
	}
	// Танк 16x16 в (0,0) смотрит вверх: срез ствола над центром
	flash := env.effects.GetFlashLights()[0]
	if flash.Position != (types.Position{X: 8, Y: 0}) {
		t.Errorf("вспышка в %v, ожидалась у ствола", flash.Position)
	}

	for range 3 {
		env.effects.Update()
	}
	if tank.IsRecoiling() {
		t.Error("отдача не закончилась")
	}
}

// Пыль идёт только из-под движущихся танков
func TestVisualEffectsUseCases_DustOnlyWhenMoving(t *testing.T) {
	env := newVisualEffectsTestEnv()
	tank := newPlayerTank(types.TankRolePlayer1)
	env.tankCommon.tanks = []*types.TankEntity{tank}

	for range 60 {
		env.effects.Update()
	}
	if got := len(env.effects.GetParticles()); got != 0 {
		t.Fatalf("стоящий танк пылит: %d частиц", got)
	}

	// Пыль случайна и недолговечна: считаем, появлялась ли она
	// хоть раз за время движения, а не только в последнем кадре
	tank.State = types.TankStateMoving
	dusted := false
	for range 120 {
		env.effects.Update()
		dusted = dusted || len(env.effects.GetParticles()) > 0
	}
	if !dusted {
		t.Error("движущийся танк не пылит")
	}
}

// Обломки вылетают из разрушенной области стены цветами её спрайта;
// прозрачные места спрайта обломков не дают
func TestVisualEffectsUseCases_BlockDebrisFromWallCells(t *testing.T) {
	env := newVisualEffectsTestEnv()
	brickColor := color.NRGBA{R: 200, G: 90, B: 30, A: 255}
	sprite := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	// Нижняя половина тайла — кирпич, верхняя прозрачна
	for y := 4; y < 8; y++ {
		for x := 0; x < 8; x++ {
			sprite.SetNRGBA(x, y, brickColor)
		}
	}
	env.tilesets.Image = sprite
	block := types.NewBlockEntity("brick", 40, 40, 8, &stubImageProvider{})

	// Пуля летит вверх и срезает нижний слой 8x4
	env.effects.RequestEffect(types.VisualEventEntity{
		Kind:      types.VisualEventBlockDebris,
		Position:  types.Position{X: 40, Y: 44},
		Size:      types.Size{Width: 8, Height: 4},
		Direction: types.DirectionUp,
		Block:     block,
	})
	env.effects.Update()

	particles := env.effects.GetParticles()
	if len(particles) != 8 {
		t.Fatalf("обломков %d, ожидалось 8 кусков 2x2", len(particles))
	}
	for _, particle := range particles {
		if particle.Color != brickColor {
			t.Errorf("цвет обломка %v, ожидался цвет спрайта", particle.Color)
		}
		if particle.Position.Y < 44 || particle.Position.Y > 48 {
			t.Errorf(
				"обломок в %v, ожидался в срезанном слое",
				particle.Position,
			)
		}
	}

	// Верхний слой прозрачен — обломков нет
	env.repository.SetParticles(nil)
	env.effects.RequestEffect(types.VisualEventEntity{
		Kind:     types.VisualEventBlockDebris,
		Position: types.Position{X: 40, Y: 40},
		Size:     types.Size{Width: 8, Height: 4},
		Block:    block,
	})
	env.effects.Update()
	if got := len(env.effects.GetParticles()); got != 0 {
		t.Errorf("обломков из прозрачной области %d, ожидалось 0", got)
	}
}

// Попадание во врага трясёт экран, в игрока — сильнее
func TestVisualEffectsUseCases_ShakeStrongerForPlayer(t *testing.T) {
	maxOffset := func(kind types.VisualEventKind) float64 {
		env := newVisualEffectsTestEnv()
		env.effects.RequestEffect(types.VisualEventEntity{Kind: kind})
		env.effects.Update()
		strongest := 0.0
		for range 30 {
			offset := env.effects.GetShakeOffset()
			strongest = max(strongest, math.Abs(offset.X), math.Abs(offset.Y))
			env.effects.Update()
		}
		return strongest
	}

	enemy := maxOffset(types.VisualEventEnemyHit)
	player := maxOffset(types.VisualEventPlayerHit)
	if enemy == 0 {
		t.Error("попадание во врага не трясёт экран")
	}
	if player <= enemy {
		t.Errorf("тряска от попадания в игрока %v, во врага %v", player, enemy)
	}
}

// Столкновение пуль: вспышка, искры во все стороны и тряска
func TestVisualEffectsUseCases_BulletClash(t *testing.T) {
	env := newVisualEffectsTestEnv()
	env.effects.RequestEffect(types.VisualEventEntity{
		Kind:     types.VisualEventBulletClash,
		Position: types.Position{X: 60, Y: 60},
	})
	env.effects.Update()

	if len(env.effects.GetParticles()) == 0 {
		t.Error("bullet clash without sparks")
	}
	if len(env.effects.GetFlashLights()) != 1 {
		t.Error("bullet clash without a flash")
	}
	shaken := false
	for range 5 {
		if env.effects.GetShakeOffset() != (types.Position{}) {
			shaken = true
		}
		env.effects.Update()
	}
	if !shaken {
		t.Error("bullet clash does not shake the screen")
	}
}

func newFlyingBullet(
	direction types.Direction,
	specs *types.SpecsEntity,
) *types.BulletEntity {
	return types.NewBulletEntity(
		types.Position{X: 100, Y: 100},
		types.Size{Width: 4, Height: 4},
		types.GROUND,
		nil,
		direction,
		specs,
		nil,
	)
}

// Трассер: искры появляются за кормой пули и летят назад
func TestVisualEffectsUseCases_TracerBehindBullet(t *testing.T) {
	env := newVisualEffectsTestEnv()
	env.bullets.bullets = []*types.BulletEntity{
		newFlyingBullet(types.DirectionRight, nil),
	}
	sparks := 0
	for range 40 {
		env.effects.Update()
		for _, particle := range env.effects.GetParticles() {
			sparks++
			if particle.Position.X > 100 {
				t.Fatalf(
					"tracer spark at %v is ahead of the bullet tail",
					particle.Position,
				)
			}
		}
	}
	if sparks == 0 {
		t.Error("flying bullet leaves no tracer sparks")
	}
}

// Усиленная пуля искрит бело-голубым, обычная — тёплым
func TestVisualEffectsUseCases_TracerByBulletUpgrade(t *testing.T) {
	colorsOf := func(specs *types.SpecsEntity) []color.NRGBA {
		env := newVisualEffectsTestEnv()
		env.bullets.bullets = []*types.BulletEntity{
			newFlyingBullet(types.DirectionUp, specs),
		}
		colors := []color.NRGBA{}
		for range 40 {
			env.effects.Update()
			for _, particle := range env.effects.GetParticles() {
				colors = append(colors, particle.Color)
			}
		}
		return colors
	}

	for _, spark := range colorsOf(types.NewSpecsEntity(0, 32, false, 120, 1)) {
		if spark.B >= spark.R {
			t.Errorf("regular tracer spark %v is not warm", spark)
		}
	}
	reinforced := colorsOf(types.NewSpecsEntity(2, 32, true, 150, 1))
	if len(reinforced) == 0 {
		t.Fatal("reinforced bullet leaves no tracer sparks")
	}
	for _, spark := range reinforced {
		if spark.B < spark.R {
			t.Errorf("reinforced tracer spark %v is not blue-white", spark)
		}
	}
}

// Выстрел: конус света вперёд по стволу и дымок
func TestVisualEffectsUseCases_ShotConeAndSmoke(t *testing.T) {
	env := newVisualEffectsTestEnv()
	env.effects.RequestEffect(types.VisualEventEntity{
		Kind:      types.VisualEventShot,
		Position:  types.Position{X: 50, Y: 50},
		Direction: types.DirectionRight,
	})
	env.effects.Update()

	cone := false
	for _, light := range env.effects.GetFlashLights() {
		if light.IsCone() && light.Direction == (types.Position{X: 1, Y: 0}) {
			cone = true
		}
	}
	if !cone {
		t.Error("shot without a muzzle cone along the barrel")
	}

	for range 20 {
		env.effects.Update()
	}
	if len(env.effects.GetParticles()) == 0 {
		t.Error("no muzzle smoke left after the sparks faded")
	}
}
