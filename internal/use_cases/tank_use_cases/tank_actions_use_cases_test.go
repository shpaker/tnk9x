package tank_use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases/tank_use_cases"
)

func newStoppedTank() *types.TankEntity {
	tank := types.NewDefaultTankEntity(types.TankRoleEnemy, types.DirectionDown)
	tank.State = types.TankStateStopped
	return &tank
}

func TestApplyDecisionRotatesAndMoves(t *testing.T) {
	actions := tank_use_cases.NewTankActionsUseCases(
		nil,
		nil,
		&testutil.FakeTankCommonUseCases{},
		&stubRenderUseCases{},
		nil,
		nil,
		&testutil.FakeVisualEffectsUseCases{},
	)
	tank := newStoppedTank()

	actions.ApplyDecision(tank, types.EnemyAIDecision{
		Direction: types.DirectionLeft,
		Move:      true,
	})

	if tank.Direction != types.DirectionLeft ||
		tank.State != types.TankStateMoving {
		t.Fatalf(
			"tank must turn left and move: dir=%v state=%v",
			tank.Direction,
			tank.State,
		)
	}
}

// Решение без движения только поворачивает танк: он стоит и целится
func TestApplyDecisionWithoutMoveKeepsTankStopped(t *testing.T) {
	actions := tank_use_cases.NewTankActionsUseCases(
		nil,
		nil,
		&testutil.FakeTankCommonUseCases{},
		&stubRenderUseCases{},
		nil,
		nil,
		&testutil.FakeVisualEffectsUseCases{},
	)
	tank := newStoppedTank()

	actions.ApplyDecision(
		tank,
		types.EnemyAIDecision{Direction: types.DirectionUp},
	)

	if tank.Direction != types.DirectionUp ||
		tank.State != types.TankStateStopped {
		t.Fatalf(
			"tank must turn up and stay: dir=%v state=%v",
			tank.Direction,
			tank.State,
		)
	}
}
