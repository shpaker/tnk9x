package game

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

func TestTanksRepositoryAssignsEnemyIDs(t *testing.T) {
	repository := NewTanksRepository()
	first := types.NewDefaultTankEntity(
		types.TankRoleEnemy,
		types.DirectionDown,
	)
	second := types.NewDefaultTankEntity(
		types.TankRoleEnemy,
		types.DirectionDown,
	)

	repository.AddEnemy(&first)
	repository.AddEnemy(&second)

	if first.GetID() != 1 || second.GetID() != 2 {
		t.Fatalf(
			"expected ids 1 and 2, got %d and %d",
			first.GetID(),
			second.GetID(),
		)
	}
}
