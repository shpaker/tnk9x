package scripting

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

func TestLuaEngineUpdateEnemyAIDecision(t *testing.T) {
	engine := NewLuaEngine()
	defer engine.Close()

	script := `
function updateEnemyAI(x, y, direction, state)
    return true, LEFT_DIRECTION
end
`
	engine.SetGlobalNumber("LEFT_DIRECTION", float64(types.DirectionLeft))
	if err := engine.LoadScript(script); err != nil {
		t.Fatalf("LoadScript failed: %v", err)
	}

	decision, err := engine.UpdateEnemyAI(10, 20, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Direction != types.DirectionLeft {
		t.Fatalf("expected DirectionLeft, got %v", decision.Direction)
	}
}

// Скрипт вернул shouldMove == false: движок отдаёт нулевое решение и nil
// вместо ошибки — существующее поведение, на него завязан AI-адаптер.
func TestLuaEngineUpdateEnemyAIShouldNotMoveQuirk(t *testing.T) {
	engine := NewLuaEngine()
	defer engine.Close()

	script := `
function updateEnemyAI(x, y, direction, state)
    return false, 3
end
`
	if err := engine.LoadScript(script); err != nil {
		t.Fatalf("LoadScript failed: %v", err)
	}

	decision, err := engine.UpdateEnemyAI(0, 0, 0, 0)
	if err != nil {
		t.Fatalf("expected nil error for shouldMove=false, got: %v", err)
	}
	if decision != (types.EnemyAIDecision{}) {
		t.Fatalf("expected zero decision, got %+v", decision)
	}
}

func TestLuaEngineUpdateEnemyAINonBoolFirstResult(t *testing.T) {
	engine := NewLuaEngine()
	defer engine.Close()

	script := `
function updateEnemyAI(x, y, direction, state)
    return 1, 2
end
`
	if err := engine.LoadScript(script); err != nil {
		t.Fatalf("LoadScript failed: %v", err)
	}

	// Не-булево первое значение трактуется как false
	decision, err := engine.UpdateEnemyAI(0, 0, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision != (types.EnemyAIDecision{}) {
		t.Fatalf("expected zero decision, got %+v", decision)
	}
}

func TestLuaEngineUpdateEnemyAIMissingFunction(t *testing.T) {
	engine := NewLuaEngine()
	defer engine.Close()

	if _, err := engine.UpdateEnemyAI(0, 0, 0, 0); err == nil {
		t.Fatal("expected error when updateEnemyAI is not defined")
	}
}

func TestLuaEngineReceivesArgumentsAndGlobals(t *testing.T) {
	engine := NewLuaEngine()
	defer engine.Close()

	engine.SetGlobalNumber("MAP_WIDTH_PX", 208)

	script := `
function updateEnemyAI(x, y, direction, state)
    if x == 16 and y == 32 and direction == 2 and state == 1
        and MAP_WIDTH_PX == 208 then
        return true, direction
    end
    return false, 0
end
`
	if err := engine.LoadScript(script); err != nil {
		t.Fatalf("LoadScript failed: %v", err)
	}

	decision, err := engine.UpdateEnemyAI(16, 32, 2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Direction != types.Direction(2) {
		t.Fatalf("arguments/globals not visible to script: %+v", decision)
	}
}
