package tank_use_cases_test

import (
	"math"
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/services"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
	"github.com/shpaker/tnk9x/internal/use_cases"
	"github.com/shpaker/tnk9x/internal/use_cases/tank_use_cases"
)

type commonTestEnv struct {
	tanksRepo *game.TanksRepository
	specsUC   *use_cases.SpecsUseCases
	session   *session_entities.StageSessionEntity
	common    *tank_use_cases.TankCommonUseCases
}

func newCommonTestEnv() *commonTestEnv {
	tanksRepo := game.NewTanksRepository()
	specsUC := use_cases.NewSpecsUseCases()
	session := session_entities.NewStageSessionEntity()
	common := tank_use_cases.NewTankCommonUseCases(
		services.NewTankBrakingService(),
		&stubRenderUseCases{},
		tanksRepo,
		specsUC,
		use_cases.NewMapUseCases(nil),
		session,
	)

	return &commonTestEnv{
		tanksRepo: tanksRepo,
		specsUC:   specsUC,
		session:   session,
		common:    common,
	}
}

func (env *commonTestEnv) newTank(
	role types.TankRole,
	direction types.Direction,
	state types.TankState,
	level uint,
) *types.TankEntity {
	tankValue := types.NewDefaultTankEntity(role, direction)
	tank := &tankValue
	tank.Position = types.Position{X: 100, Y: 100}
	tank.PrevPosition = tank.Position
	tank.State = state
	tank.SetSpecs(env.specsUC.GetTankSpecs(role == types.TankRoleEnemy, level))
	return tank
}

// Движение по направлениям: смещение = скорость из спецификаций * dt
func TestTankCommonUseCases_Update_Movement(t *testing.T) {
	tests := []struct {
		name      string
		role      types.TankRole
		direction types.Direction
		level     uint
		wantX     float64
		wantY     float64
	}{
		// Игрок уровня 0: скорость 32, dt 0.25 -> смещение 8
		{"игрок вверх", types.TankRolePlayer1, types.DirectionUp, 0, 100, 92},
		{"игрок вниз", types.TankRolePlayer1, types.DirectionDown, 0, 100, 108},
		{"игрок влево", types.TankRolePlayer1, types.DirectionLeft, 0, 92, 100},
		{
			"игрок вправо",
			types.TankRolePlayer1,
			types.DirectionRight,
			0,
			108,
			100,
		},
		// Враг уровня 1: скорость 48, dt 0.25 -> смещение 12
		{"враг вправо", types.TankRoleEnemy, types.DirectionRight, 1, 112, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newCommonTestEnv()
			tank := env.newTank(
				tt.role,
				tt.direction,
				types.TankStateMoving,
				tt.level,
			)

			if err := env.common.Update(tank, 0.25); err != nil {
				t.Fatalf("обновление: %v", err)
			}

			want := types.Position{X: tt.wantX, Y: tt.wantY}
			if tank.Position != want {
				t.Errorf("позиция %v, ожидалась %v", tank.Position, want)
			}
			if tank.PrevPosition != (types.Position{X: 100, Y: 100}) {
				t.Errorf(
					"PrevPosition %v, ожидалась стартовая",
					tank.PrevPosition,
				)
			}
		})
	}
}

// Без спецификаций используется скорость по умолчанию 32
func TestTankCommonUseCases_Update_DefaultSpeedWithoutSpecs(t *testing.T) {
	env := newCommonTestEnv()
	tank := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		types.TankStateMoving,
		0,
	)
	tank.SetSpecs(nil)

	if err := env.common.Update(tank, 0.25); err != nil {
		t.Fatalf("обновление: %v", err)
	}
	if tank.Position.X != 108 {
		t.Errorf("X = %v, ожидалось 108", tank.Position.X)
	}
}

func TestTankCommonUseCases_Update_StoppedTankDoesNotMove(t *testing.T) {
	env := newCommonTestEnv()
	tank := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		types.TankStateStopped,
		0,
	)

	if err := env.common.Update(tank, 0.25); err != nil {
		t.Fatalf("обновление: %v", err)
	}
	if tank.Position != (types.Position{X: 100, Y: 100}) {
		t.Errorf("стоящий танк сдвинулся: %v", tank.Position)
	}
}

func TestTankCommonUseCases_Update_InactiveTankError(t *testing.T) {
	env := newCommonTestEnv()

	for _, state := range []types.TankState{
		types.TankStateSpawning,
		types.TankStateExploding,
		types.TankStateExploded,
	} {
		tank := env.newTank(
			types.TankRolePlayer1,
			types.DirectionUp,
			state,
			0,
		)
		if err := env.common.Update(tank, 0.25); err == nil {
			t.Errorf("состояние %v: ожидалась ошибка", state)
		}
	}
}

// Торможение делегируется реальному сервису: танк доезжает до сетки 4px
func TestTankCommonUseCases_Update_BrakingSnapsToGrid(t *testing.T) {
	env := newCommonTestEnv()
	tank := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		types.TankStateBraking,
		0,
	)
	tank.Position.X = 101

	// Большой dt: сразу достигает 104 и останавливается
	if err := env.common.Update(tank, 1.0); err != nil {
		t.Fatalf("обновление: %v", err)
	}
	if tank.Position.X != 104 {
		t.Errorf("X = %v, ожидалось 104", tank.Position.X)
	}
	if tank.State != types.TankStateStopped {
		t.Errorf("состояние %v, ожидалось Stopped", tank.State)
	}
}

func TestTankCommonUseCases_Update_BrakingPartialStep(t *testing.T) {
	env := newCommonTestEnv()
	tank := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		types.TankStateBraking,
		0,
	)
	tank.Position.X = 101

	if err := env.common.Update(tank, 1.0/60.0); err != nil {
		t.Fatalf("обновление: %v", err)
	}

	want := 101.0 + 32.0/60.0
	if math.Abs(tank.Position.X-want) > 1e-9 {
		t.Errorf("X = %v, ожидалось %v", tank.Position.X, want)
	}
	if tank.State != types.TankStateBraking {
		t.Errorf("состояние %v, ожидалось Braking", tank.State)
	}
}

// Торможение на льду: танк доскальзывает +4px за обычной точкой остановки
func TestTankCommonUseCases_Update_BrakingSlidesOnIce(t *testing.T) {
	tanksRepo := game.NewTanksRepository()
	specsUC := use_cases.NewSpecsUseCases()

	// Лёд (104,104)-(112,112) накрывает центр танка (109,108)
	ice := types.NewBlockEntity("ice", 104, 104, 8, nil)
	mapEntity := types.NewMapEntity(
		types.Size{Width: 208, Height: 208},
		types.MapBlocks{ice},
		nil,
	)
	common := tank_use_cases.NewTankCommonUseCases(
		services.NewTankBrakingService(),
		&stubRenderUseCases{},
		tanksRepo,
		specsUC,
		use_cases.NewMapUseCases(mapEntity),
		session_entities.NewStageSessionEntity(),
	)

	tankValue := types.NewDefaultTankEntity(
		types.TankRolePlayer1,
		types.DirectionRight,
	)
	tank := &tankValue
	tank.Position = types.Position{X: 101, Y: 100}
	tank.State = types.TankStateBraking
	tank.SetSpecs(specsUC.GetTankSpecs(false, 0))

	if err := common.Update(tank, 1.0); err != nil {
		t.Fatalf("обновление: %v", err)
	}

	// Обычная остановка на 104, на льду — 108
	if tank.Position.X != 108 {
		t.Errorf("X = %v, ожидалось 108", tank.Position.X)
	}
	if tank.State != types.TankStateStopped {
		t.Errorf("состояние %v, ожидалось Stopped", tank.State)
	}
	if tank.SlideTarget != nil {
		t.Errorf("SlideTarget не сброшен: %v", *tank.SlideTarget)
	}
}

// При заморозке враг стоит на месте, танк игрока продолжает движение
func TestTankCommonUseCases_Update_FrozenEnemyDoesNotMove(t *testing.T) {
	env := newCommonTestEnv()
	enemy := env.newTank(
		types.TankRoleEnemy,
		types.DirectionRight,
		types.TankStateMoving,
		0,
	)
	player := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		types.TankStateMoving,
		0,
	)

	env.session.FreezeEnemies(600)

	if err := env.common.Update(enemy, 1.0/60.0); err != nil {
		t.Fatalf("обновление врага: %v", err)
	}
	if err := env.common.Update(player, 1.0/60.0); err != nil {
		t.Fatalf("обновление игрока: %v", err)
	}

	if enemy.Position.X != 100 {
		t.Errorf("замороженный враг сместился: X = %v", enemy.Position.X)
	}
	if player.Position.X == 100 {
		t.Error("танк игрока не сместился при заморозке врагов")
	}
}

// После торможения с NextDirection танк продолжает движение в новую сторону
func TestTankCommonUseCases_Update_BrakingWithNextDirection(t *testing.T) {
	env := newCommonTestEnv()
	tank := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		types.TankStateBraking,
		0,
	)
	tank.Position.X = 101
	next := types.DirectionUp
	tank.NextDirection = &next

	if err := env.common.Update(tank, 1.0); err != nil {
		t.Fatalf("обновление: %v", err)
	}
	if tank.Direction != types.DirectionUp {
		t.Errorf("направление %v, ожидалось Up", tank.Direction)
	}
	if tank.State != types.TankStateMoving {
		t.Errorf("состояние %v, ожидалось Moving", tank.State)
	}
	if tank.NextDirection != nil {
		t.Errorf("NextDirection не сброшен")
	}
}

// UpdateAllTanks обходит все танки репозитория, ошибки глотаются
func TestTankCommonUseCases_UpdateAllTanks(t *testing.T) {
	env := newCommonTestEnv()
	player := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		types.TankStateMoving,
		0,
	)
	enemy := env.newTank(
		types.TankRoleEnemy,
		types.DirectionDown,
		types.TankStateMoving,
		0,
	)
	exploded := env.newTank(
		types.TankRoleEnemy,
		types.DirectionUp,
		types.TankStateExploded,
		0,
	)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, player)
	env.tanksRepo.AddEnemy(enemy)
	env.tanksRepo.AddEnemy(exploded)

	if err := env.common.UpdateAllTanks(0.25); err != nil {
		t.Fatalf("обновление: %v", err)
	}
	if player.Position.X != 108 {
		t.Errorf("игрок не сдвинулся: %v", player.Position)
	}
	if enemy.Position.Y != 108 {
		t.Errorf("враг не сдвинулся: %v", enemy.Position)
	}
	if exploded.Position != (types.Position{X: 100, Y: 100}) {
		t.Errorf("взорванный танк сдвинулся: %v", exploded.Position)
	}
}

func TestTankCommonUseCases_IsAnyPlayerTankMoving(t *testing.T) {
	env := newCommonTestEnv()
	if env.common.IsAnyPlayerTankMoving() {
		t.Error("пустой репозиторий: ожидалось false")
	}

	player := env.newTank(
		types.TankRolePlayer1,
		types.DirectionUp,
		types.TankStateStopped,
		0,
	)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, player)
	if env.common.IsAnyPlayerTankMoving() {
		t.Error("стоящий игрок: ожидалось false")
	}

	// Движущийся враг не учитывается
	enemy := env.newTank(
		types.TankRoleEnemy,
		types.DirectionUp,
		types.TankStateMoving,
		0,
	)
	env.tanksRepo.AddEnemy(enemy)
	if env.common.IsAnyPlayerTankMoving() {
		t.Error("движется только враг: ожидалось false")
	}

	player.State = types.TankStateMoving
	if !env.common.IsAnyPlayerTankMoving() {
		t.Error("движущийся игрок: ожидалось true")
	}
}

// Повышение и понижение уровня с ограничением диапазона 0-3
func TestTankCommonUseCases_LevelUpDownClamping(t *testing.T) {
	env := newCommonTestEnv()
	tank := env.newTank(
		types.TankRolePlayer1,
		types.DirectionUp,
		types.TankStateStopped,
		0,
	)

	for i := 0; i < 5; i++ {
		env.common.LevelUp(tank)
	}
	if got := tank.GetSpecs().GetLevel(); got != 3 {
		t.Errorf("уровень после повышений: %d, ожидалось 3", got)
	}
	if got := tank.GetSpecs().GetSpeed(); got != 40.0 {
		t.Errorf("скорость игрока 3 уровня: %v, ожидалось 40", got)
	}

	for i := 0; i < 5; i++ {
		env.common.LevelDown(tank)
	}
	if got := tank.GetSpecs().GetLevel(); got != 0 {
		t.Errorf("уровень после понижений: %d, ожидалось 0", got)
	}

	// nil и танк без спецификаций не ломают вызовы
	env.common.LevelUp(nil)
	env.common.LevelDown(nil)
	bare := env.newTank(
		types.TankRolePlayer1,
		types.DirectionUp,
		types.TankStateStopped,
		0,
	)
	bare.SetSpecs(nil)
	env.common.LevelUp(bare)
	if bare.GetSpecs() != nil {
		t.Errorf("спецификации появились из ниоткуда")
	}
}

func TestTankEntity_AnimationName(t *testing.T) {
	env := newCommonTestEnv()

	tests := []struct {
		name string
		tank *types.TankEntity
		want string
	}{
		{"nil танк", nil, "player1_level1_tank_up"},
		{
			"игрок 1 уровень 0 вверх",
			env.newTank(
				types.TankRolePlayer1,
				types.DirectionUp,
				types.TankStateStopped,
				0,
			),
			"player1_level1_tank_up",
		},
		{
			"игрок 1 уровень 1 вправо",
			env.newTank(
				types.TankRolePlayer1,
				types.DirectionRight,
				types.TankStateStopped,
				1,
			),
			"player1_level2_tank_right",
		},
		{
			"игрок 2 уровень 2 влево",
			env.newTank(
				types.TankRolePlayer2,
				types.DirectionLeft,
				types.TankStateStopped,
				2,
			),
			"player2_level3_tank_left",
		},
		{
			"враг уровень 3 вниз",
			env.newTank(
				types.TankRoleEnemy,
				types.DirectionDown,
				types.TankStateStopped,
				3,
			),
			"enemy_level4_tank_down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tank.AnimationName(); got != tt.want {
				t.Errorf("имя %q, ожидалось %q", got, tt.want)
			}
		})
	}
}

func TestTankEntity_AnimationName_Fallbacks(t *testing.T) {
	env := newCommonTestEnv()

	// Без спецификаций уровень 0, пустая роль трактуется как player1
	noSpecs := env.newTank(
		types.TankRoleEnemy,
		types.DirectionUp,
		types.TankStateStopped,
		0,
	)
	noSpecs.SetSpecs(nil)
	if got := noSpecs.AnimationName(); got != "enemy_level1_tank_up" {
		t.Errorf("без спецификаций: %q", got)
	}

	emptyRole := env.newTank(
		types.TankRole(""),
		types.DirectionUp,
		types.TankStateStopped,
		0,
	)
	if got := emptyRole.AnimationName(); got != "player1_level1_tank_up" {
		t.Errorf("пустая роль: %q", got)
	}

	// Уровень выше 3 обрезается, неизвестное направление трактуется как up
	weird := env.newTank(
		types.TankRolePlayer1,
		types.Direction(99),
		types.TankStateStopped,
		0,
	)
	weird.SetSpecs(types.NewSpecsEntity(7, 32, false, 120, 1))
	if got := weird.AnimationName(); got != "player1_level4_tank_up" {
		t.Errorf("уровень 7 и направление 99: %q", got)
	}
}
