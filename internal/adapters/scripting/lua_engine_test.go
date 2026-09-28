package scripting

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/services"
	"github.com/shpaker/tnk9x/internal/types"
)

func newTestEngine(t *testing.T, script string) interfaces.IAIScriptEngine {
	t.Helper()
	engine := NewLuaEngine(services.NewNavigationService())
	t.Cleanup(engine.Close)
	if err := engine.LoadScript(script); err != nil {
		t.Fatalf("LoadScript failed: %v", err)
	}
	return engine
}

// testContext — пустое поле 64x64 с танком в левом верхнем углу
func testContext(blocks ...*types.BlockEntity) types.EnemyAIContext {
	return types.EnemyAIContext{
		Self: types.AITankInfo{
			ID:        3,
			Position:  types.Position{X: 0, Y: 0},
			Size:      types.Size{Width: 16, Height: 16},
			Direction: types.DirectionDown,
			Level:     2,
		},
		HQ: types.AIHQInfo{
			Position: types.Position{X: 48, Y: 48},
			Size:     types.Size{Width: 16, Height: 16},
			Intact:   true,
		},
		StageNumber: 5,
		Tick:        100,
		Grid: types.NewNavGridEntity(
			types.Size{Width: 64, Height: 64}, 4, blocks,
		),
	}
}

func TestLuaEngineParsesDecisionTable(t *testing.T) {
	engine := newTestEngine(t, `
function updateEnemyAI(ctx)
    return {direction = 2, move = true, shoot = true}
end
`)

	decision, err := engine.UpdateEnemyAI(testContext())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := types.EnemyAIDecision{Direction: types.DirectionLeft, Move: true, Shoot: true}
	if decision != want {
		t.Fatalf("expected %+v, got %+v", want, decision)
	}
}

// Без полей решение безопасное: текущее направление, стоять, не стрелять
func TestLuaEngineDecisionDefaults(t *testing.T) {
	engine := newTestEngine(t, `
function updateEnemyAI(ctx)
    return nil
end
`)

	decision, err := engine.UpdateEnemyAI(testContext())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := types.EnemyAIDecision{Direction: types.DirectionDown}
	if decision != want {
		t.Fatalf("expected %+v, got %+v", want, decision)
	}
}

func TestLuaEngineMissingFunction(t *testing.T) {
	engine := newTestEngine(t, ``)

	if _, err := engine.UpdateEnemyAI(testContext()); err == nil {
		t.Fatal("expected error when updateEnemyAI is not defined")
	}
}

func TestLuaEngineScriptError(t *testing.T) {
	engine := newTestEngine(t, `
function updateEnemyAI(ctx)
    error("boom")
end
`)

	if _, err := engine.UpdateEnemyAI(testContext()); err == nil {
		t.Fatal("expected script error to propagate")
	}
}

func TestLuaEnginePassesContextAndGlobals(t *testing.T) {
	engine := newTestEngine(t, `
function updateEnemyAI(ctx)
    local ok = ctx.tick == 100 and ctx.stage == 5
        and ctx.self.id == 3 and ctx.self.level == 2 and ctx.self.size == 16
        and ctx.self.reinforced == false
        and #ctx.players == 1 and ctx.players[1].x == 32
        and #ctx.enemies == 1 and ctx.enemies[1].id == 4
        and #ctx.bullets == 1 and ctx.bullets[1].enemy == true
        and ctx.hq.intact and ctx.hq.x == 48
        and MAP_WIDTH_PX == 64
    return {direction = 3, shoot = ok}
end
`)
	engine.SetGlobalNumber("MAP_WIDTH_PX", 64)

	context := testContext()
	context.Players = []types.AITankInfo{{Position: types.Position{X: 32, Y: 0}, Size: types.Size{Width: 16, Height: 16}}}
	context.Enemies = []types.AITankInfo{{ID: 4, Size: types.Size{Width: 16, Height: 16}}}
	context.Bullets = []types.AIBulletInfo{{FromEnemy: true}}

	decision, err := engine.UpdateEnemyAI(context)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Shoot {
		t.Fatal("context or globals not visible to script")
	}
}

func TestLuaEngineAPIFunctions(t *testing.T) {
	engine := newTestEngine(t, `
function updateEnemyAI(ctx)
    local dir, length = ai.findPath(0, 0, 48, 0)
    local kind, distance = ai.castRay(8, 8, 3)
    local ok = dir == 3 and length == 6
        and kind == "brick" and distance == 16
        and ai.tileAt(26, 10) == "brick" and ai.tileAt(40, 40) == nil
    return {direction = dir, shoot = ok}
end
`)

	// Кирпич в стороне от пути по верхней кромке не перекрывает проезд,
	// но стоит на линии огня из центра танка
	context := testContext(types.NewBlockEntity(string(types.Brick), 24, 8, 8, nil))
	context.Self.Position = types.Position{X: 0, Y: 0}

	decision, err := engine.UpdateEnemyAI(context)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Shoot {
		t.Fatalf("ai.* returned unexpected values: %+v", decision)
	}
}

func TestLuaEngineAPIOutsideCall(t *testing.T) {
	engine := newTestEngine(t, `
outside = ai.findPath(0, 0, 8, 8) == nil and ai.castRay(0, 0, 0) == nil
    and ai.tileAt(0, 0) == nil
function updateEnemyAI(ctx)
    return {shoot = outside}
end
`)

	decision, err := engine.UpdateEnemyAI(testContext())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Shoot {
		t.Fatal("ai.* must return nil outside UpdateEnemyAI")
	}
}
