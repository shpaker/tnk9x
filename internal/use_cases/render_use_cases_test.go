package use_cases_test

import (
	"image/color"
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

type renderTestEnv struct {
	animations  *game.AnimationsRepository
	tileService *testutil.FakeTileService
	tilesUC     *use_cases.TilesUseCases
	render      *use_cases.RenderUseCases
}

func newRenderTestEnv() *renderTestEnv {
	animations := game.NewAnimationsRepository()
	tileService := &testutil.FakeTileService{}
	tilesUC := use_cases.NewTilesUseCasesWithAnimations(
		nil, // реестр тайлсетов не нужен для анимаций танков
		types.TilesetTypePlayer,
		animations,
		tileService,
		nil,
	)

	return &renderTestEnv{
		animations:  animations,
		tileService: tileService,
		tilesUC:     tilesUC,
		render:      use_cases.NewRenderUseCases(tilesUC),
	}
}

func (env *renderTestEnv) newTankInState(
	role types.TankRole,
	state types.TankState,
) *types.TankEntity {
	tankValue := types.NewDefaultTankEntity(role, types.DirectionUp)
	tank := &tankValue
	tank.State = state
	return tank
}

// Новая анимация создаётся по имени и роли танка, прошлая останавливается,
// новая регистрируется и синхронизируется с состоянием
func TestRenderUseCases_UpdateTankAnimation(t *testing.T) {
	env := newRenderTestEnv()
	tank := testutil.MovingTank(env.newTankInState(types.TankRolePlayer1, types.TankStateActive))
	previous := image_providers.NewAnimationProvider(
		types.AnimationData{{Image: "old", Duration: 1}},
	)
	previous.IsAnimating = true
	tank.Image = previous

	env.render.UpdateTankAnimation(tank)

	if len(env.tileService.Created) != 1 ||
		env.tileService.Created[0] != "player/player1_level1_tank_up" {
		t.Errorf("созданные анимации: %v", env.tileService.Created)
	}
	if previous.IsAnimating {
		t.Error("прошлая анимация не остановлена")
	}

	animation, ok := tank.Image.(*image_providers.AnimationProvider)
	if !ok || animation == previous {
		t.Fatalf("изображение не заменено: %T", tank.Image)
	}
	// Танк движется — новая анимация сразу запущена
	if !animation.IsAnimating {
		t.Error("анимация движущегося танка не запущена")
	}

	all := env.animations.GetAllAnimations()
	if len(all) != 1 || all[0] != animation {
		t.Errorf("анимация не зарегистрирована: %v", all)
	}
}

// Вражеский танк получает анимацию из тайлсета enemy
func TestRenderUseCases_UpdateTankAnimation_Enemy(t *testing.T) {
	env := newRenderTestEnv()
	tank := env.newTankInState(types.TankRoleEnemy, types.TankStateActive)

	env.render.UpdateTankAnimation(tank)

	if len(env.tileService.Created) != 1 ||
		env.tileService.Created[0] != "enemy/enemy_level1_tank_up" {
		t.Errorf("созданные анимации: %v", env.tileService.Created)
	}

	// Остановленный танк — анимация не запускается
	animation, ok := tank.Image.(*image_providers.AnimationProvider)
	if !ok {
		t.Fatalf("изображение не AnimationProvider: %T", tank.Image)
	}
	if animation.IsAnimating {
		t.Error("анимация остановленного танка запущена")
	}
}

func TestRenderUseCases_UpdateTankAnimation_NilTank(t *testing.T) {
	env := newRenderTestEnv()

	env.render.UpdateTankAnimation(nil)

	if len(env.tileService.Created) != 0 {
		t.Errorf("создана анимация для nil-танка: %v", env.tileService.Created)
	}
}

// Ошибка тайл-сервиса оставляет прежнее изображение танка
func TestRenderUseCases_UpdateTankAnimation_TileError(t *testing.T) {
	env := newRenderTestEnv()
	env.tileService.Err = errTileNotFound
	tank := testutil.MovingTank(env.newTankInState(types.TankRolePlayer1, types.TankStateActive))
	previous := image_providers.NewAnimationProvider(
		types.AnimationData{{Image: "old", Duration: 1}},
	)
	previous.IsAnimating = true
	tank.Image = previous

	env.render.UpdateTankAnimation(tank)

	if tank.Image != previous {
		t.Errorf("изображение заменено при ошибке: %v", tank.Image)
	}
	if !previous.IsAnimating {
		t.Error("прошлая анимация остановлена при ошибке")
	}
	if got := len(env.animations.GetAllAnimations()); got != 0 {
		t.Errorf("анимаций в репозитории %d, ожидалось 0", got)
	}
}

func TestRenderUseCases_SyncTankAnimationWithState(t *testing.T) {
	tests := []struct {
		name          string
		state         types.TankState
		motion        func(*types.TankEntity) *types.TankEntity
		wasAnimating  bool
		wantAnimating bool
	}{
		{"остановка прекращает анимацию", types.TankStateActive, nil, true, false},
		{"движение запускает анимацию", types.TankStateActive, testutil.MovingTank, false, true},
		{"докатывание запускает анимацию", types.TankStateActive, testutil.DockingTank, false, true},
		{"взрыв останавливает анимацию", types.TankStateExploding, nil, true, false},
		{"спавн не запускает анимацию", types.TankStateSpawning, nil, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newRenderTestEnv()
			tank := env.newTankInState(types.TankRolePlayer1, tt.state)
			tank.Position.Y = 101 // вне узла сетки: есть куда докатывать
			if tt.motion != nil {
				tt.motion(tank)
			}
			animation := image_providers.NewAnimationProvider(
				types.AnimationData{{Image: "frame", Duration: 1}},
			)
			animation.IsAnimating = tt.wasAnimating
			tank.Image = animation

			env.render.SyncTankAnimationWithState(tank)

			if animation.IsAnimating != tt.wantAnimating {
				t.Errorf(
					"IsAnimating=%v, ожидалось %v",
					animation.IsAnimating,
					tt.wantAnimating,
				)
			}
		})
	}
}

// Не-анимационные изображения синхронизация не трогает
func TestRenderUseCases_SyncTankAnimationWithState_NonAnimation(
	t *testing.T,
) {
	env := newRenderTestEnv()
	tank := testutil.MovingTank(env.newTankInState(types.TankRolePlayer1, types.TankStateActive))
	tank.Image = nil
	env.render.SyncTankAnimationWithState(tank)

	static := &stubImageProvider{}
	tank.Image = static
	env.render.SyncTankAnimationWithState(tank)
	if tank.Image != static {
		t.Errorf("изображение изменено: %v", tank.Image)
	}
}

func TestRenderUseCases_IsTankAnimationFinished(t *testing.T) {
	env := newRenderTestEnv()
	tank := env.newTankInState(types.TankRolePlayer1, types.TankStateSpawning)

	// Нет изображения — анимация не завершена
	tank.Image = nil
	if env.render.IsTankSpawnAnimationFinished(tank) {
		t.Error("nil-изображение: ожидалось false")
	}
	if env.render.IsTankExplosionAnimationFinished(tank) {
		t.Error("nil-изображение: ожидалось false")
	}

	// Статичное изображение — не анимация
	tank.Image = &stubImageProvider{}
	if env.render.IsTankSpawnAnimationFinished(tank) {
		t.Error("статичное изображение: ожидалось false")
	}
	if env.render.IsTankExplosionAnimationFinished(tank) {
		t.Error("статичное изображение: ожидалось false")
	}

	// Провайдер анимации: результат следует за IsFinished
	animation := image_providers.NewAnimationProvider(
		types.AnimationData{{Image: "frame", Duration: 1}},
	)
	tank.Image = animation

	animation.IsAnimating = true
	if env.render.IsTankSpawnAnimationFinished(tank) {
		t.Error("идущая анимация: ожидалось false")
	}
	animation.IsAnimating = false
	if !env.render.IsTankSpawnAnimationFinished(tank) {
		t.Error("завершённая анимация: ожидалось true")
	}
	if !env.render.IsTankExplosionAnimationFinished(tank) {
		t.Error("завершённая анимация: ожидалось true")
	}
}

// Мигание обновляется у каждого объекта, nil-элементы пропускаются
func TestRenderUseCases_UpdateBlink(t *testing.T) {
	env := newRenderTestEnv()
	first := env.newTankInState(types.TankRolePlayer1, types.TankStateActive)
	second := env.newTankInState(types.TankRoleEnemy, types.TankStateActive)

	env.render.UpdateBlink(nil) // nil-список не приводит к панике

	// Флаг мигания танка переключается каждые 10 тиков
	for i := 0; i < 10; i++ {
		env.render.UpdateBlink([]types.IBlink{first, nil, second})
	}

	if !first.GetBlinkFlag() || !second.GetBlinkFlag() {
		t.Errorf(
			"флаги мигания: %v %v, ожидалось true true",
			first.GetBlinkFlag(),
			second.GetBlinkFlag(),
		)
	}
}

// Тон даёт только мигающий враг и только во включённой фазе: враг
// с бонусом — красный, тяжёлый с запасом здоровья 2-4 — цвет брони
func TestRenderUseCases_TankTint(t *testing.T) {
	env := newRenderTestEnv()

	newTank := func(
		role types.TankRole,
		specs *types.SpecsEntity,
		hitPoints uint,
	) *types.TankEntity {
		tank := testutil.MovingTank(env.newTankInState(role, types.TankStateActive))
		tank.SetSpecs(specs)
		tank.SetHitPoints(hitPoints)
		return tank
	}
	heavySpecs := func() *types.SpecsEntity {
		return types.NewSpecsEntity(3, 1, false, 1, 1)
	}
	// blinkOn переводит танк во включённую фазу мигания
	blinkOn := func(tank *types.TankEntity) *types.TankEntity {
		for !tank.GetBlinkFlag() {
			tank.UpdateBlink()
		}
		return tank
	}
	withBonus := func(tank *types.TankEntity) *types.TankEntity {
		tank.SetWithBonus(true)
		return tank
	}

	tests := []struct {
		name      string
		tank      *types.TankEntity
		wantColor color.NRGBA
		wantOK    bool
	}{
		{
			"4 HP — красный",
			blinkOn(newTank(types.TankRoleEnemy, heavySpecs(), 4)),
			color.NRGBA{R: 240, G: 150, B: 140, A: 255},
			true,
		},
		{
			"3 HP — жёлтый",
			blinkOn(newTank(types.TankRoleEnemy, heavySpecs(), 3)),
			color.NRGBA{R: 235, G: 215, B: 150, A: 255},
			true,
		},
		{
			"2 HP — зелёный",
			blinkOn(newTank(types.TankRoleEnemy, heavySpecs(), 2)),
			color.NRGBA{R: 165, G: 220, B: 160, A: 255},
			true,
		},
		{
			"выключенная фаза мигания — без тона",
			newTank(types.TankRoleEnemy, heavySpecs(), 4),
			color.NRGBA{},
			false,
		},
		{
			"враг с бонусом — красный",
			blinkOn(withBonus(newTank(
				types.TankRoleEnemy,
				types.NewSpecsEntity(0, 1, false, 1, 1),
				1,
			))),
			color.NRGBA{R: 235, G: 125, B: 115, A: 255},
			true,
		},
		{
			"тяжёлый с бонусом — красный бонуса, не цвет брони",
			blinkOn(withBonus(newTank(types.TankRoleEnemy, heavySpecs(), 3))),
			color.NRGBA{R: 235, G: 125, B: 115, A: 255},
			true,
		},
		{
			"враг с бонусом в выключенной фазе — без тона",
			withBonus(newTank(types.TankRoleEnemy, heavySpecs(), 4)),
			color.NRGBA{},
			false,
		},
		{
			"1 HP — без тона",
			blinkOn(newTank(types.TankRoleEnemy, heavySpecs(), 1)),
			color.NRGBA{},
			false,
		},
		{
			"0 HP — без тона",
			blinkOn(newTank(types.TankRoleEnemy, heavySpecs(), 0)),
			color.NRGBA{},
			false,
		},
		{
			"не тяжёлый враг",
			blinkOn(newTank(
				types.TankRoleEnemy,
				types.NewSpecsEntity(2, 1, false, 1, 1),
				4,
			)),
			color.NRGBA{},
			false,
		},
		{
			"игрок",
			blinkOn(newTank(types.TankRolePlayer1, heavySpecs(), 4)),
			color.NRGBA{},
			false,
		},
		{
			"враг без спецификаций",
			blinkOn(newTank(types.TankRoleEnemy, nil, 4)),
			color.NRGBA{},
			false,
		},
		{"nil-танк", nil, color.NRGBA{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotColor, gotOK := env.render.TankTint(tt.tank)
			if gotOK != tt.wantOK || gotColor != tt.wantColor {
				t.Errorf(
					"TankTint = (%v, %v), ожидалось (%v, %v)",
					gotColor,
					gotOK,
					tt.wantColor,
					tt.wantOK,
				)
			}
		})
	}
}

// Мигают враг с бонусом и тяжёлый враг с индикацией здоровья
func TestRenderUseCases_IsTankBlinking(t *testing.T) {
	env := newRenderTestEnv()

	newTank := func(
		role types.TankRole,
		level uint,
		hitPoints uint,
		withBonus bool,
	) *types.TankEntity {
		tank := testutil.MovingTank(env.newTankInState(role, types.TankStateActive))
		tank.SetSpecs(types.NewSpecsEntity(level, 1, false, 1, 1))
		tank.SetHitPoints(hitPoints)
		tank.SetWithBonus(withBonus)
		return tank
	}

	tests := []struct {
		name string
		tank *types.TankEntity
		want bool
	}{
		{"враг с бонусом", newTank(types.TankRoleEnemy, 0, 1, true), true},
		{"тяжёлый враг", newTank(types.TankRoleEnemy, 3, 4, false), true},
		{
			"тяжёлый враг с 1 HP",
			newTank(types.TankRoleEnemy, 3, 1, false),
			false,
		},
		{"обычный враг", newTank(types.TankRoleEnemy, 1, 1, false), false},
		{"игрок", newTank(types.TankRolePlayer1, 3, 4, true), false},
		{"nil-танк", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := env.render.IsTankBlinking(tt.tank); got != tt.want {
				t.Errorf("IsTankBlinking = %v, ожидалось %v", got, tt.want)
			}
		})
	}
}
