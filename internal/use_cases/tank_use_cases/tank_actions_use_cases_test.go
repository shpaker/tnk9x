package tank_use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases/tank_use_cases"
)

func newStoppedTank() *types.TankEntity {
	tank := types.NewDefaultTankEntity(types.TankRoleEnemy, types.DirectionDown)
	tank.State = types.TankStateActive
	return &tank
}

func TestApplyDecisionRotatesAndMoves(t *testing.T) {
	actions := tank_use_cases.NewTankActionsUseCases(
		nil,
		&testutil.FakeTankCommonUseCases{},
		&stubRenderUseCases{},
		nil,
		&testutil.FakeVisualEffectsUseCases{},
	)
	tank := newStoppedTank()

	actions.ApplyDecision(tank, types.EnemyAIDecision{
		Direction: types.DirectionLeft,
		Move:      true,
	})

	if drive, ok := tank.GetDrive(); tank.Direction != types.DirectionLeft ||
		!ok || drive != types.DirectionLeft {
		t.Fatalf(
			"tank must turn left and drive: dir=%v drive=%v",
			tank.Direction,
			drive,
		)
	}
}

func TestApplyDecisionWithoutMoveKeepsTankStopped(t *testing.T) {
	actions := tank_use_cases.NewTankActionsUseCases(
		nil,
		&testutil.FakeTankCommonUseCases{},
		&stubRenderUseCases{},
		nil,
		&testutil.FakeVisualEffectsUseCases{},
	)
	tank := newStoppedTank()

	actions.ApplyDecision(
		tank,
		types.EnemyAIDecision{Direction: types.DirectionUp},
	)

	if _, driving := tank.GetDrive(); tank.Direction != types.DirectionUp ||
		driving {
		t.Fatalf("tank must turn up and stay: dir=%v", tank.Direction)
	}
}

func newMotionActions() *tank_use_cases.TankActionsUseCases {
	return tank_use_cases.NewTankActionsUseCases(
		nil,
		&testutil.FakeTankCommonUseCases{},
		&stubRenderUseCases{},
		nil,
		&testutil.FakeVisualEffectsUseCases{},
	)
}

func newDrivingTank(x float64) *types.TankEntity {
	tank := types.NewDefaultTankEntity(
		types.TankRolePlayer1,
		types.DirectionRight,
	)
	tank.Position = types.Position{X: x, Y: 100}
	return testutil.MovingTank(&tank)
}

func TestRotateWhileMovingDefersTurn(t *testing.T) {
	actions := newMotionActions()
	tank := newDrivingTank(101)

	_ = actions.Rotate(tank, types.DirectionUp)

	drive, _ := tank.GetDrive()
	if tank.Direction != types.DirectionRight || drive != types.DirectionUp {
		t.Fatalf("dir=%v drive=%v, want Right then Up", tank.Direction, drive)
	}
}

func TestMoveCancelsDocking(t *testing.T) {
	actions := newMotionActions()
	tank := newDrivingTank(101)
	actions.Stop(tank, false)

	if err := actions.Move(tank); err != nil {
		t.Fatal(err)
	}
	if drive, ok := tank.GetDrive(); !ok || drive != types.DirectionRight {
		t.Fatalf("drive=%v, want Right", drive)
	}
}

func TestStopByCollisionHaltsInPlace(t *testing.T) {
	actions := newMotionActions()
	tank := newDrivingTank(101)

	actions.Stop(tank, true)

	if !tank.IsStopped() || tank.Position.X != 101 {
		t.Fatalf(
			"stopped=%v X=%v, want halted at 101",
			tank.IsStopped(),
			tank.Position.X,
		)
	}
	if _, ok := tank.GetDrive(); ok {
		t.Fatal("halted tank keeps driving")
	}
}
