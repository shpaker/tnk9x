package scripting

import (
	"fmt"
	"io/fs"
	"testing"

	"github.com/shpaker/tnk9x"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/services"
	"github.com/shpaker/tnk9x/internal/types"
)

// Сценарии поведения настоящего enemies.lua на поле 64x64.
// Случайность детерминирована: math.random() возвращает 0.5,
// math.random(a, b) — верхнюю границу b.
const enemiesMapSize = 64

// stubRandom делает math.random детерминированным: без аргументов
// возвращает value, с диапазоном — верхнюю границу
func stubRandom(
	t *testing.T,
	engine interfaces.IAIScriptEngine,
	value float64,
) {
	t.Helper()
	script := fmt.Sprintf(`
math.random = function(a, b)
    if a == nil then return %g end
    if b == nil then return a end
    return b
end
`, value)
	if err := engine.LoadScript(script); err != nil {
		t.Fatalf("stub random: %v", err)
	}
}

func newEnemiesEngine(t *testing.T) interfaces.IAIScriptEngine {
	t.Helper()
	source, err := fs.ReadFile(tnk9x.FS, "assets/scripts/enemies.lua")
	if err != nil {
		t.Fatalf("read enemies.lua: %v", err)
	}

	engine := NewLuaEngine(services.NewNavigationService())
	t.Cleanup(engine.Close)
	engine.SetGlobalNumber("MAP_WIDTH_PX", enemiesMapSize)
	engine.SetGlobalNumber("MAP_HEIGHT_PX", enemiesMapSize)
	engine.SetGlobalNumber("TANK_SIZE_PX", 16)
	if err := engine.LoadScript(string(source)); err != nil {
		t.Fatalf("load enemies.lua: %v", err)
	}
	stubRandom(t, engine, 0.5)
	return engine
}

func tankAt(x, y float64) types.AITankInfo {
	return types.AITankInfo{
		Position: types.Position{X: x, Y: y},
		Size:     types.Size{Width: 16, Height: 16},
	}
}

func enemiesContext(
	level uint,
	stage uint,
	self types.AITankInfo,
	blocks ...*types.BlockEntity,
) types.EnemyAIContext {
	self.ID = 1
	self.Level = level
	return types.EnemyAIContext{
		Self:        self,
		StageNumber: stage,
		HQ: types.AIHQInfo{
			Position: types.Position{X: 48, Y: 48},
			Size:     types.Size{Width: 16, Height: 16},
			Intact:   true,
		},
		Grid: types.NewNavGridEntity(
			types.Size{
				Width:  enemiesMapSize,
				Height: enemiesMapSize,
			},
			4,
			blocks,
		),
	}
}

func decide(
	t *testing.T,
	engine interfaces.IAIScriptEngine,
	context types.EnemyAIContext,
	tick int,
) types.EnemyAIDecision {
	t.Helper()
	context.Tick = tick
	decision, err := engine.UpdateEnemyAI(context)
	if err != nil {
		t.Fatalf("updateEnemyAI failed: %v", err)
	}
	return decision
}

func steel(x, y float64) *types.BlockEntity {
	return types.NewBlockEntity(string(types.Steel), x, y, 8, nil)
}

func brick(x, y float64) *types.BlockEntity {
	return types.NewBlockEntity(string(types.Brick), x, y, 8, nil)
}

// Игрок на линии огня: выстрел только после задержки реакции
func TestEnemiesShootsPlayerAfterReaction(t *testing.T) {
	engine := newEnemiesEngine(t)
	self := tankAt(0, 0)
	self.Direction = types.DirectionDown
	context := enemiesContext(3, 1, self)
	context.Players = []types.AITankInfo{tankAt(0, 48)}

	first := decide(t, engine, context, 100)
	if first.Shoot || first.Move || first.Direction != types.DirectionDown {
		t.Fatalf(
			"armor must hold aim without shooting at first sight: %+v",
			first,
		)
	}

	later := decide(t, engine, context, 160)
	if !later.Shoot || later.Direction != types.DirectionDown {
		t.Fatalf("armor must shoot after reaction delay: %+v", later)
	}

	again := decide(t, engine, context, 170)
	if again.Shoot {
		t.Fatalf("armor must rest between aimed shots: %+v", again)
	}
}

// Бетон между танком и игроком: не стреляет в его сторону
func TestEnemiesDoesNotShootThroughSteel(t *testing.T) {
	engine := newEnemiesEngine(t)
	self := tankAt(0, 0)
	self.Direction = types.DirectionDown
	context := enemiesContext(3, 1, self, steel(4, 32))
	context.Players = []types.AITankInfo{tankAt(0, 48)}

	for tick := 100; tick <= 200; tick += 50 {
		decision := decide(t, engine, context, tick)
		if decision.Shoot && decision.Direction == types.DirectionDown {
			t.Fatalf("must not shoot into steel: %+v", decision)
		}
	}
}

// Союзник на линии огня: не стреляет
func TestEnemiesDoesNotShootAlly(t *testing.T) {
	engine := newEnemiesEngine(t)
	self := tankAt(0, 0)
	self.Direction = types.DirectionDown
	context := enemiesContext(3, 1, self)
	context.Players = []types.AITankInfo{tankAt(0, 48)}
	ally := tankAt(0, 24)
	ally.ID = 2
	context.Enemies = []types.AITankInfo{ally}

	for tick := 100; tick <= 200; tick += 50 {
		if decision := decide(t, engine, context, tick); decision.Shoot {
			t.Fatalf("must not shoot through ally: %+v", decision)
		}
	}
}

// На поздних уровнях танк доворачивает ствол к игроку на одной линии
func TestEnemiesTurnsToAlignedPlayer(t *testing.T) {
	engine := newEnemiesEngine(t)
	self := tankAt(0, 0)
	self.Direction = types.DirectionDown
	context := enemiesContext(1, 20, self)
	context.Players = []types.AITankInfo{tankAt(48, 0)}

	decision := decide(t, engine, context, 100)
	if decision.Direction != types.DirectionRight {
		t.Fatalf("fast tank must turn to the player: %+v", decision)
	}
}

// Power в фазе штаба едет к нему по пути
func TestEnemiesPowerHeadsToHQ(t *testing.T) {
	engine := newEnemiesEngine(t)
	self := tankAt(0, 16)
	self.Direction = types.DirectionUp
	context := enemiesContext(2, 20, self)

	decide(t, engine, context, 0) // Первый вызов заводит память танка
	decision := decide(t, engine, context, 100000)
	if !decision.Move {
		t.Fatalf("power tank must move towards hq: %+v", decision)
	}
	if decision.Direction != types.DirectionDown &&
		decision.Direction != types.DirectionRight {
		t.Fatalf("power tank must head to hq: %+v", decision)
	}
}

// Power на одной линии со штабом пробивает кирпич перед ним без задержки
func TestEnemiesPowerBreachesWallToHQ(t *testing.T) {
	engine := newEnemiesEngine(t)
	self := tankAt(48, 0)
	self.Direction = types.DirectionDown
	context := enemiesContext(2, 1, self, brick(48, 32), brick(56, 32))

	decide(t, engine, context, 0)
	decision := decide(t, engine, context, 100000)
	if !decision.Shoot || decision.Direction != types.DirectionDown {
		t.Fatalf("power tank must shoot the wall towards hq: %+v", decision)
	}
}

// Застрявший танк меняет направление
func TestEnemiesRecoversFromStuck(t *testing.T) {
	engine := newEnemiesEngine(t)
	self := tankAt(16, 16)
	self.Direction = types.DirectionDown
	context := enemiesContext(0, 1, self)

	var decision types.EnemyAIDecision
	for tick := 1; tick <= 13; tick++ {
		decision = decide(t, engine, context, tick)
	}
	if decision.Direction == types.DirectionDown || !decision.Move {
		t.Fatalf("stuck tank must turn: %+v", decision)
	}
}

// Fast на позднем уровне уходит с линии летящей в него пули
func TestEnemiesFastDodgesBullet(t *testing.T) {
	engine := newEnemiesEngine(t)
	self := tankAt(24, 40)
	self.Direction = types.DirectionLeft
	context := enemiesContext(1, 20, self)
	context.Bullets = []types.AIBulletInfo{{
		Position:  types.Position{X: 30, Y: 4},
		Size:      types.Size{Width: 4, Height: 4},
		Direction: types.DirectionDown,
	}}

	decision := decide(t, engine, context, 100)
	if !decision.Move ||
		(decision.Direction != types.DirectionLeft && decision.Direction != types.DirectionRight) {
		t.Fatalf("fast tank must sidestep the bullet: %+v", decision)
	}
}

// Fast, смотрящий на летящую пулю, стреляет навстречу
func TestEnemiesFastShootsIncomingBullet(t *testing.T) {
	engine := newEnemiesEngine(t)
	self := tankAt(24, 40)
	self.Direction = types.DirectionUp
	context := enemiesContext(1, 20, self)
	context.Bullets = []types.AIBulletInfo{{
		Position:  types.Position{X: 30, Y: 4},
		Size:      types.Size{Width: 4, Height: 4},
		Direction: types.DirectionDown,
	}}

	decision := decide(t, engine, context, 100)
	if !decision.Shoot || decision.Direction != types.DirectionUp {
		t.Fatalf("fast tank must shoot back: %+v", decision)
	}
}

// На ранних уровнях танк не замечает игрока дальше дальности обзора
func TestEnemiesIgnoresDistantPlayerOnEarlyStages(t *testing.T) {
	engine := newEnemiesEngine(t)
	engine.SetGlobalNumber("MAP_WIDTH_PX", 208)
	engine.SetGlobalNumber("MAP_HEIGHT_PX", 208)
	self := tankAt(0, 0)
	self.Direction = types.DirectionDown
	context := enemiesContext(3, 1, self)
	context.Grid = types.NewNavGridEntity(
		types.Size{Width: 208, Height: 208},
		4,
		nil,
	)
	context.Players = []types.AITankInfo{tankAt(0, 192)}

	for tick := 100; tick <= 300; tick += 100 {
		if decision := decide(t, engine, context, tick); decision.Shoot {
			t.Fatalf(
				"distant player must not be noticed on stage 1: %+v",
				decision,
			)
		}
	}

	context.StageNumber = 20
	context.Self.ID = 2
	decide(t, engine, context, 400)
	if decision := decide(t, engine, context, 500); !decision.Shoot {
		t.Fatalf("distant player must be noticed on late stages: %+v", decision)
	}
}

// Бродящий танк на ходу обстреливает стену по курсу
func TestEnemiesShootsWallAhead(t *testing.T) {
	engine := newEnemiesEngine(t)
	self := tankAt(0, 0)
	self.Direction = types.DirectionDown
	context := enemiesContext(2, 20, self, brick(4, 48))

	decision := decide(t, engine, context, 100)
	if !decision.Move || !decision.Shoot ||
		decision.Direction != types.DirectionDown {
		t.Fatalf(
			"power tank must fire at the wall ahead on the move: %+v",
			decision,
		)
	}
}

// Бродящий танк разворачивается и прорубает проход в боковой стене
func TestEnemiesBreachesSideWall(t *testing.T) {
	engine := newEnemiesEngine(t)
	stubRandom(t, engine, 0.01)
	self := tankAt(16, 16)
	self.Direction = types.DirectionDown
	context := enemiesContext(0, 1, self, brick(32, 20))

	decision := decide(t, engine, context, 100)
	if decision.Move || !decision.Shoot ||
		decision.Direction != types.DirectionRight {
		t.Fatalf("tank must turn and breach the side wall: %+v", decision)
	}

	// Продолжает пробивать, пока стена на линии огня
	next := decide(t, engine, context, 101)
	if !next.Shoot || next.Direction != types.DirectionRight {
		t.Fatalf("tank must keep breaching: %+v", next)
	}
}

func bonusAt(x, y float64) types.AIBonusInfo {
	return types.AIBonusInfo{
		Position: types.Position{X: x, Y: y},
		Size:     types.Size{Width: 16, Height: 16},
		Type:     types.BonusTypeStar,
	}
}

// Ближайший к бонусу враг едет за ним; если другой ближе — не едет
func TestEnemiesNearestRacesForBonus(t *testing.T) {
	engine := newEnemiesEngine(t)
	context := enemiesContext(0, 1, tankAt(24, 0))
	context.Self.Direction = types.DirectionDown
	context.Bonuses = []types.AIBonusInfo{bonusAt(0, 0)}

	decision := decide(t, engine, context, 1000)
	if decision.Direction != types.DirectionLeft || !decision.Move {
		t.Fatalf("ожидался путь влево к бонусу, получено %+v", decision)
	}

	other := tankAt(0, 16)
	other.ID = 2
	context.Enemies = []types.AITankInfo{other}
	decision = decide(t, engine, context, 1001)
	if decision.Direction == types.DirectionLeft {
		t.Errorf("к бонусу должен ехать ближайший союзник: %+v", decision)
	}
}

// Враг с лодкой прокладывает путь к бонусу через воду
func TestEnemiesBoatPathsOverWater(t *testing.T) {
	var blocks []*types.BlockEntity
	for y := 0.0; y < enemiesMapSize; y += 8 {
		blocks = append(blocks,
			types.NewBlockEntity(string(types.Water), 16, y, 8, nil),
			types.NewBlockEntity(string(types.Water), 24, y, 8, nil),
		)
	}
	engine := newEnemiesEngine(t)
	self := tankAt(0, 0)
	self.Boat = true
	context := enemiesContext(0, 1, self, blocks...)
	context.Self.Direction = types.DirectionRight
	context.Bonuses = []types.AIBonusInfo{bonusAt(48, 0)}

	decision := decide(t, engine, context, 1000)
	if decision.Direction != types.DirectionRight || !decision.Move {
		t.Errorf("ожидался путь вправо через воду, получено %+v", decision)
	}
}
