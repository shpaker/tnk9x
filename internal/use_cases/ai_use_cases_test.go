package use_cases_test

import (
	"errors"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

type fakeScriptEngine struct {
	decision types.EnemyAIDecision
	err      error

	gotX, gotY              float64
	gotDirection, gotState  int
	gotContext              types.EnemyAIContext
	updateEnemyAICallsCount int
}

func (e *fakeScriptEngine) LoadScript(source string) error { return nil }

func (e *fakeScriptEngine) SetGlobalNumber(name string, value float64) {}

func (e *fakeScriptEngine) UpdateEnemyAI(
	x, y float64,
	direction, state int,
	context types.EnemyAIContext,
) (types.EnemyAIDecision, error) {
	e.updateEnemyAICallsCount++
	e.gotX, e.gotY = x, y
	e.gotDirection, e.gotState = direction, state
	e.gotContext = context
	return e.decision, e.err
}

func (e *fakeScriptEngine) Close() {}

func TestAIUseCasesExecuteAINilTank(t *testing.T) {
	engine := &fakeScriptEngine{}
	uc := use_cases.NewAIUseCases(engine)

	_, err := uc.ExecuteAI(nil, types.EnemyAIContext{})
	if err == nil {
		t.Fatal("expected error for nil tank")
	}
	if engine.updateEnemyAICallsCount != 0 {
		t.Fatal("engine must not be called for nil tank")
	}
}

func TestAIUseCasesExecuteAIPassesTankData(t *testing.T) {
	engine := &fakeScriptEngine{
		decision: types.EnemyAIDecision{Direction: types.DirectionLeft},
	}
	uc := use_cases.NewAIUseCases(engine)

	tank := types.NewDefaultTankEntity(
		types.TankRoleEnemy,
		types.DirectionDown,
	)
	tank.Position = types.Position{X: 48, Y: 96}
	tank.State = types.TankStateStopped

	context := types.EnemyAIContext{
		Phase:     types.EnemyAIPhaseSiege,
		TargetX:   104,
		TargetY:   200,
		HasTarget: true,
	}
	decision, err := uc.ExecuteAI(&tank, context)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Direction != types.DirectionLeft {
		t.Fatalf("decision not passed through: %+v", decision)
	}
	if engine.gotX != 48 || engine.gotY != 96 {
		t.Fatalf("position not passed: %v, %v", engine.gotX, engine.gotY)
	}
	if engine.gotDirection != int(types.DirectionDown) {
		t.Fatalf("direction not passed: %d", engine.gotDirection)
	}
	if engine.gotState != int(types.TankStateStopped) {
		t.Fatalf("state not passed: %d", engine.gotState)
	}
	if engine.gotContext != context {
		t.Fatalf("context not passed: %+v", engine.gotContext)
	}
}

func TestAIUseCasesExecuteAIPropagatesError(t *testing.T) {
	engine := &fakeScriptEngine{err: errors.New("script failed")}
	uc := use_cases.NewAIUseCases(engine)

	tank := types.NewDefaultTankEntity(
		types.TankRoleEnemy,
		types.DirectionUp,
	)
	if _, err := uc.ExecuteAI(&tank, types.EnemyAIContext{}); err == nil {
		t.Fatal("expected engine error to propagate")
	}
}
