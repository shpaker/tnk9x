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
	tank := env.newTankInState(types.TankRolePlayer1, types.TankStateMoving)
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
	tank := env.newTankInState(types.TankRoleEnemy, types.TankStateStopped)

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
	tank := env.newTankInState(types.TankRolePlayer1, types.TankStateMoving)
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
		wasAnimating  bool
		wantAnimating bool
	}{
		{"остановка прекращает анимацию", types.TankStateStopped, true, false},
		{"движение запускает анимацию", types.TankStateMoving, false, true},
		{"торможение запускает анимацию", types.TankStateBraking, false, true},
		{"взрыв останавливает анимацию", types.TankStateExploding, true, false},
		{
			"спавн не запускает анимацию",
			types.TankStateSpawning,
			false,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newRenderTestEnv()
			tank := env.newTankInState(types.TankRolePlayer1, tt.state)
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
	tank := env.newTankInState(types.TankRolePlayer1, types.TankStateMoving)
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
	first := env.newTankInState(types.TankRolePlayer1, types.TankStateStopped)
	second := env.newTankInState(types.TankRoleEnemy, types.TankStateStopped)

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

// Вражеский танк с бонусом невидим в выключенной фазе мигания
func TestRenderUseCases_IsTankVisible(t *testing.T) {
	env := newRenderTestEnv()

	newTank := func(
		role types.TankRole,
		withBonus bool,
		blinkOn bool,
	) *types.TankEntity {
		tank := env.newTankInState(role, types.TankStateMoving)
		tank.SetWithBonus(withBonus)
		if blinkOn {
			for i := 0; i < 10; i++ {
				tank.UpdateBlink()
			}
		}
		return tank
	}

	tests := []struct {
		name string
		tank *types.TankEntity
		want bool
	}{
		{"nil-танк", nil, false},
		{
			"враг с бонусом в тёмной фазе",
			newTank(types.TankRoleEnemy, true, false),
			false,
		},
		{
			"враг с бонусом в видимой фазе",
			newTank(types.TankRoleEnemy, true, true),
			true,
		},
		{
			"враг без бонуса",
			newTank(types.TankRoleEnemy, false, false),
			true,
		},
		{
			"игрок с флагом бонуса",
			newTank(types.TankRolePlayer1, true, false),
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := env.render.IsTankVisible(tt.tank); got != tt.want {
				t.Errorf("IsTankVisible = %v, ожидалось %v", got, tt.want)
			}
		})
	}
}

// Слой здоровья рисуется только для тяжёлого вражеского танка
// с запасом здоровья 2-4
func TestRenderUseCases_TankHealthOverlay(t *testing.T) {
	env := newRenderTestEnv()

	newTank := func(
		role types.TankRole,
		specs *types.SpecsEntity,
		hitPoints uint,
	) *types.TankEntity {
		tank := env.newTankInState(role, types.TankStateMoving)
		tank.SetSpecs(specs)
		tank.SetHitPoints(hitPoints)
		return tank
	}
	heavySpecs := func() *types.SpecsEntity {
		return types.NewSpecsEntity(3, 1, false, 1, 1)
	}

	tests := []struct {
		name      string
		tank      *types.TankEntity
		wantColor color.NRGBA
		wantOK    bool
	}{
		{
			"4 HP — красный",
			newTank(types.TankRoleEnemy, heavySpecs(), 4),
			color.NRGBA{R: 255, G: 0, B: 0, A: 128},
			true,
		},
		{
			"3 HP — жёлтый",
			newTank(types.TankRoleEnemy, heavySpecs(), 3),
			color.NRGBA{R: 255, G: 255, B: 0, A: 128},
			true,
		},
		{
			"2 HP — зелёный",
			newTank(types.TankRoleEnemy, heavySpecs(), 2),
			color.NRGBA{R: 0, G: 255, B: 0, A: 128},
			true,
		},
		{
			"1 HP — без слоя",
			newTank(types.TankRoleEnemy, heavySpecs(), 1),
			color.NRGBA{},
			false,
		},
		{
			"0 HP — без слоя",
			newTank(types.TankRoleEnemy, heavySpecs(), 0),
			color.NRGBA{},
			false,
		},
		{
			"не тяжёлый враг",
			newTank(
				types.TankRoleEnemy,
				types.NewSpecsEntity(2, 1, false, 1, 1),
				4,
			),
			color.NRGBA{},
			false,
		},
		{
			"игрок",
			newTank(types.TankRolePlayer1, heavySpecs(), 4),
			color.NRGBA{},
			false,
		},
		{
			"враг без спецификаций",
			newTank(types.TankRoleEnemy, nil, 4),
			color.NRGBA{},
			false,
		},
		{"nil-танк", nil, color.NRGBA{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotColor, gotOK := env.render.TankHealthOverlay(tt.tank)
			if gotOK != tt.wantOK || gotColor != tt.wantColor {
				t.Errorf(
					"TankHealthOverlay = (%v, %v), ожидалось (%v, %v)",
					gotColor,
					gotOK,
					tt.wantColor,
					tt.wantOK,
				)
			}
		})
	}
}
