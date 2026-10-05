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

// Носители бонусов без явной разметки — 4-й, 11-й и 18-й танки,
// как в оригинале
func TestWaveUseCases_DefaultBonusCarrierNumbers(t *testing.T) {
	session := session_entities.NewStageSessionEntity()
	session.SetUpLevel(types.NewLevelEntity(1, "TEST", 4, 0, []types.WaveSpec{
		{Tanks: make([]types.WaveTank, 20), DelayTicks: 30},
	}, false, nil))
	waves := use_cases.NewWaveUseCases()

	for number := 1; number <= 20; number++ {
		tank, ok := waves.NextTank(session)
		if !ok {
			t.Fatalf("танк %d: сценарий закончился раньше времени", number)
		}
		want := number == 4 || number == 11 || number == 18
		if tank.HasBonus != want {
			t.Errorf("танк %d: HasBonus %v, ожидалось %v",
				number, tank.HasBonus, want)
		}
		waves.CommitSpawn(session, 0)
	}
}

// При игре вдвоём пауза между спаунами короче на 20 тиков, как в оригинале
func TestWaveUseCases_CoopSpawnDelay(t *testing.T) {
	newSession := func(
		playerCount uint,
	) *session_entities.StageSessionEntity {
		session := session_entities.NewStageSessionEntity()
		session.SetUpLevel(types.NewLevelEntity(1, "TEST", 4, 0,
			[]types.WaveSpec{
				{Tanks: make([]types.WaveTank, 3), DelayTicks: 90},
			}, true, nil))
		session.SetPlayerCount(playerCount)
		return session
	}
	countdown := func(session *session_entities.StageSessionEntity) int {
		ticks := 0
		for !session.CanSpawnNextEnemy() && ticks < 200 {
			session.UpdateEnemySpawnCountdown()
			ticks++
		}
		return ticks
	}
	waves := use_cases.NewWaveUseCases()

	solo := newSession(1)
	if _, ok := waves.NextTank(solo); !ok {
		t.Fatal("solo: сценарий не выдал танк")
	}
	waves.CommitSpawn(solo, 0)
	if got := countdown(solo); got != 90 {
		t.Errorf("пауза соло %d тиков, ожидалось 90", got)
	}

	coop := newSession(2)
	if _, ok := waves.NextTank(coop); !ok {
		t.Fatal("coop: сценарий не выдал танк")
	}
	waves.CommitSpawn(coop, 0)
	if got := countdown(coop); got != 70 {
		t.Errorf("пауза в кооперативе %d тиков, ожидалось 70", got)
	}
}
