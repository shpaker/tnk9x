package services

import (
	"math"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

const brakingTestDT = 1.0 / 60.0

func runBrakingUntilDone(
	t *testing.T,
	service *TankBrakingService,
	tank *types.TankEntity,
	maxTicks int,
) {
	t.Helper()
	for i := 0; i < maxTicks; i++ {
		if tank.State != types.TankStateBraking {
			return
		}
		if err := service.HandleBrakingState(tank, brakingTestDT, false); err != nil {
			t.Fatalf("HandleBrakingState returned error: %v", err)
		}
	}
	t.Fatalf(
		"braking did not finish within %d ticks, state=%v",
		maxTicks,
		tank.State,
	)
}

func TestTankBrakingService_FinishesOnMultipleOf4(t *testing.T) {
	service := NewTankBrakingService()

	tests := []struct {
		name      string
		direction types.Direction
		start     types.Position
	}{
		{
			"up from off-grid",
			types.DirectionUp,
			types.Position{X: 100, Y: 101.3},
		},
		{
			"up just past grid line",
			types.DirectionUp,
			types.Position{X: 100, Y: 100.3},
		},
		{
			"down from off-grid",
			types.DirectionDown,
			types.Position{X: 100, Y: 101.3},
		},
		{
			"left from off-grid",
			types.DirectionLeft,
			types.Position{X: 98.7, Y: 100},
		},
		{
			"right from off-grid",
			types.DirectionRight,
			types.Position{X: 98.7, Y: 100},
		},
		{
			"right just past grid line",
			types.DirectionRight,
			types.Position{X: 100.2, Y: 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tank := types.NewDefaultTankEntity(
				types.TankRolePlayer1,
				tt.direction,
			)
			tank.Position = tt.start
			tank.State = types.TankStateBraking

			runBrakingUntilDone(t, service, &tank, 600)

			var movedCoord float64
			switch tt.direction {
			case types.DirectionUp, types.DirectionDown:
				movedCoord = tank.Position.Y
			default:
				movedCoord = tank.Position.X
			}

			if math.Mod(movedCoord, 4) != 0 {
				t.Errorf(
					"braking finished off-grid: coord=%v (start=%v dir=%v)",
					movedCoord,
					tt.start,
					tt.direction,
				)
			}
			if tank.State != types.TankStateStopped {
				t.Errorf("expected Stopped state, got %v", tank.State)
			}
		})
	}
}

// На льду танк доскальзывает лишние 4px за обычной точкой остановки
func TestTankBrakingService_IceSlideExtendsStop(t *testing.T) {
	service := NewTankBrakingService()

	tests := []struct {
		name      string
		direction types.Direction
		start     types.Position
		want      types.Position
	}{
		{
			"right off-grid",
			types.DirectionRight,
			types.Position{X: 5, Y: 100},
			types.Position{X: 12, Y: 100},
		},
		{
			"right exactly on grid line",
			types.DirectionRight,
			types.Position{X: 8, Y: 100},
			types.Position{X: 12, Y: 100},
		},
		{
			"left off-grid",
			types.DirectionLeft,
			types.Position{X: 5, Y: 100},
			types.Position{X: 0, Y: 100},
		},
		{
			"up exactly on grid line",
			types.DirectionUp,
			types.Position{X: 100, Y: 8},
			types.Position{X: 100, Y: 4},
		},
		{
			"down off-grid",
			types.DirectionDown,
			types.Position{X: 100, Y: 101.3},
			types.Position{X: 100, Y: 108},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tank := types.NewDefaultTankEntity(
				types.TankRolePlayer1,
				tt.direction,
			)
			tank.Position = tt.start
			tank.State = types.TankStateBraking

			for i := 0; i < 600 && tank.State == types.TankStateBraking; i++ {
				err := service.HandleBrakingState(&tank, brakingTestDT, true)
				if err != nil {
					t.Fatalf("HandleBrakingState returned error: %v", err)
				}
			}

			if tank.Position != tt.want {
				t.Errorf(
					"позиция %v, ожидалась %v (start=%v dir=%v)",
					tank.Position,
					tt.want,
					tt.start,
					tt.direction,
				)
			}
			if tank.State != types.TankStateStopped {
				t.Errorf("состояние %v, ожидалось Stopped", tank.State)
			}
			if tank.SlideTarget != nil {
				t.Errorf("SlideTarget не сброшен: %v", *tank.SlideTarget)
			}
		})
	}
}

// Цель скольжения фиксируется один раз: съехав со льда посреди
// скольжения, танк всё равно доезжает до зафиксированной точки
func TestTankBrakingService_SlideTargetLatchedOnce(t *testing.T) {
	service := NewTankBrakingService()

	tank := types.NewDefaultTankEntity(
		types.TankRolePlayer1,
		types.DirectionRight,
	)
	tank.Position = types.Position{X: 5, Y: 100}
	tank.State = types.TankStateBraking

	if err := service.HandleBrakingState(&tank, brakingTestDT, true); err != nil {
		t.Fatalf("HandleBrakingState returned error: %v", err)
	}

	for i := 0; i < 600 && tank.State == types.TankStateBraking; i++ {
		err := service.HandleBrakingState(&tank, brakingTestDT, false)
		if err != nil {
			t.Fatalf("HandleBrakingState returned error: %v", err)
		}
	}

	if tank.Position.X != 12 {
		t.Errorf("X = %v, ожидалось 12", tank.Position.X)
	}
	if tank.SlideTarget != nil {
		t.Errorf("SlideTarget не сброшен: %v", *tank.SlideTarget)
	}
}

func TestTankBrakingService_AdoptsNextDirectionOnFinish(t *testing.T) {
	service := NewTankBrakingService()

	tank := types.NewDefaultTankEntity(
		types.TankRolePlayer1,
		types.DirectionRight,
	)
	tank.Position = types.Position{X: 98.7, Y: 100}
	tank.State = types.TankStateBraking
	next := types.DirectionUp
	tank.NextDirection = &next

	for i := 0; i < 600 && tank.State == types.TankStateBraking; i++ {
		if err := service.HandleBrakingState(&tank, brakingTestDT, false); err != nil {
			t.Fatalf("HandleBrakingState returned error: %v", err)
		}
	}

	if math.Mod(tank.Position.X, 4) != 0 {
		t.Errorf("braking finished off-grid: X=%v", tank.Position.X)
	}
	if tank.State != types.TankStateMoving {
		t.Errorf(
			"expected Moving state after adopting NextDirection, got %v",
			tank.State,
		)
	}
	if tank.Direction != types.DirectionUp {
		t.Errorf("expected direction Up, got %v", tank.Direction)
	}
	if tank.NextDirection != nil {
		t.Errorf("expected NextDirection cleared, got %v", *tank.NextDirection)
	}
}
