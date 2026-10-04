package input_adapters

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
)

// stubAIUseCases возвращает заданное решение и запоминает тик
type stubAIUseCases struct {
	decision types.EnemyAIDecision
	gotTick  int
	calls    int
}

func (s *stubAIUseCases) ExecuteAI(
	_ *types.TankEntity,
	tick int,
) (types.EnemyAIDecision, error) {
	s.calls++
	s.gotTick = tick
	return s.decision, nil
}

func newStoppedEnemy() *types.TankEntity {
	tank := types.NewDefaultTankEntity(types.TankRoleEnemy, types.DirectionDown)
	tank.State = types.TankStateActive
	return &tank
}

func TestAiInputAdapterShootsOnlyByDecision(t *testing.T) {
	actions := &recordingTankActions{}
	ai := &stubAIUseCases{}
	adapter := NewAiInputAdapter(actions, ai)
	adapter.AddTank(newStoppedEnemy())

	adapter.Update(0)
	if actions.shoots != 0 {
		t.Fatal("must not shoot without decision")
	}

	ai.decision.Shoot = true
	adapter.Update(0)
	if actions.shoots != 1 || ai.gotTick != 2 {
		t.Fatalf(
			"expected one shot on tick 2, got %d shots, tick %d",
			actions.shoots,
			ai.gotTick,
		)
	}
}

func TestAiInputAdapterSkipsInactiveAndRemovedTanks(t *testing.T) {
	ai := &stubAIUseCases{}
	adapter := NewAiInputAdapter(&recordingTankActions{}, ai)
	tank := newStoppedEnemy()
	tank.State = types.TankStateSpawning
	adapter.AddTank(tank)
	adapter.AddTank(tank)

	adapter.Update(0)
	if ai.calls != 0 {
		t.Fatal("inactive tank must not be asked")
	}

	tank.State = types.TankStateActive
	adapter.RemoveTank(tank)
	adapter.Update(0)
	if ai.calls != 0 {
		t.Fatal("removed tank must not be asked")
	}
}

// Едущий враг отпускает газ только у узла сетки 8px: там скрипт решает,
// куда дальше
func TestAiInputAdapterReleasesNearTurn(t *testing.T) {
	tests := []struct {
		name      string
		y         float64
		wantStops int
	}{
		{"у узла", 97, 1},
		{"между узлами", 100, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions := &recordingTankActions{}
			adapter := NewAiInputAdapter(actions, &stubAIUseCases{})
			tank := newStoppedEnemy()
			tank.Position.Y = tt.y
			testutil.MovingTank(tank)
			adapter.AddTank(tank)

			adapter.Update(0)
			if actions.stops != tt.wantStops {
				t.Fatalf("stops = %d, want %d", actions.stops, tt.wantStops)
			}
		})
	}
}
