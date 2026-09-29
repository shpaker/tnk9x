package use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// Враг из бонуса-танка выходит после всех танков сценария
func TestWaveUseCases_ExtraEnemyAfterWaves(t *testing.T) {
	session := session_entities.NewStageSessionEntity()
	session.SetUpLevel(types.NewLevelEntity(1, "TEST", 4, 0, []types.WaveSpec{
		{Tanks: []types.WaveTank{{Level: 0}}, DelayTicks: 30},
	}, true, nil))
	session.AddExtraEnemy(types.WaveTank{Level: 3})
	waves := use_cases.NewWaveUseCases()

	first, ok := waves.NextTank(session)
	if !ok || first.Level != 0 {
		t.Fatalf("первым ожидался танк сценария, получено %+v %v", first, ok)
	}
	waves.CommitSpawn(session, 0)

	extra, ok := waves.NextTank(session)
	if !ok || extra.Level != 3 {
		t.Fatalf("ожидался дополнительный враг, получено %+v %v", extra, ok)
	}
	waves.CommitSpawn(session, 1)

	if _, ok := waves.NextTank(session); ok {
		t.Error("после дополнительного врага резерв должен опустеть")
	}
	if session.GetTotalEnemies() != 2 || session.GetSpawnedEnemies() != 2 {
		t.Errorf("врагов всего %d, вышло %d, ожидалось 2 и 2",
			session.GetTotalEnemies(), session.GetSpawnedEnemies())
	}
}
