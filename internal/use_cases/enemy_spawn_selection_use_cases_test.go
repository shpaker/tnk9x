package use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// fixedRandom — последовательность «случайных» чисел по кругу
type fixedRandom struct {
	values []float64
	next   int
}

func (r *fixedRandom) Float64() float64 {
	value := r.values[r.next%len(r.values)]
	r.next++
	return value
}

// blockedSpawners помечает занятыми спаунеры с заданными позициями
type blockedSpawners struct {
	blocked map[types.Position]bool
}

func (s *blockedSpawners) IsSpawnerBlocked(
	position types.Position,
	size types.Size,
	tanks []*types.TankEntity,
) bool {
	return s.blocked[position]
}

var testSpawners = []types.Position{{X: 0, Y: 0}, {X: 6, Y: 0}, {X: 12, Y: 0}}

func newSpawnSelection(
	tanks []*types.TankEntity,
	blocked map[types.Position]bool,
	random ...float64,
) *use_cases.EnemySpawnSelectionUseCases {
	return use_cases.NewEnemySpawnSelectionUseCases(
		&recordingTankCommon{tanks: tanks},
		&blockedSpawners{blocked: blocked},
		&fixedRandom{values: random},
		types.SpawnLayout{
			EnemySpawners: testSpawners,
			BaseSize:      types.Size{Width: 16, Height: 16},
		},
		4,
	)
}

func newPlayerAt(x, y float64) *types.TankEntity {
	tankValue := types.NewDefaultTankEntity(
		types.TankRolePlayer1, types.DirectionUp,
	)
	tank := &tankValue
	tank.Position = types.Position{X: x, Y: y}
	tank.Size = types.Size{Width: 16, Height: 16}
	tank.State = types.TankStateActive
	return tank
}

// Без истории и игроков веса равны: случайное число делит отрезок
// на три равные части
func TestEnemySpawnSelection_EqualWeights(t *testing.T) {
	session := session_entities.NewStageSessionEntity()
	tests := map[float64]int{0.1: 0, 0.5: 1, 0.9: 2}
	for random, want := range tests {
		selection := newSpawnSelection(nil, nil, random)
		if got, ok := selection.SelectSpawner(session); !ok || got != want {
			t.Errorf(
				"random %.1f: spawner %d %v, want %d",
				random,
				got,
				ok,
				want,
			)
		}
	}
	if newSpawnSelection(nil, nil, 0).GetSpawnersCount() != 3 {
		t.Error("spawners count must be 3")
	}
}

// Занятые спаунеры пропускаются; все заняты — выбора нет
func TestEnemySpawnSelection_SkipsBlocked(t *testing.T) {
	session := session_entities.NewStageSessionEntity()
	blocked := map[types.Position]bool{
		testSpawners[0]: true,
		testSpawners[1]: true,
	}

	selection := newSpawnSelection(nil, blocked, 0.0, 0.99)
	for i := 0; i < 2; i++ {
		if got, ok := selection.SelectSpawner(session); !ok || got != 2 {
			t.Errorf("spawner %d %v, want the only free one", got, ok)
		}
	}

	blocked[testSpawners[2]] = true
	if _, ok := newSpawnSelection(nil, blocked, 0.5).SelectSpawner(session); ok {
		t.Error("every spawner is blocked, selection must fail")
	}
}

// Последний использованный спаунер выбирается реже остальных
func TestEnemySpawnSelection_RecentSpawnerPenalty(t *testing.T) {
	session := session_entities.NewStageSessionEntity()
	session.RegisterEnemySpawned(0, 0)

	// Веса 0.25, 1, 1: доля первого — 0.25 / 2.25 ≈ 0.11
	if got, _ := newSpawnSelection(nil, nil, 0.10).SelectSpawner(session); got != 0 {
		t.Errorf("random 0.10: spawner %d, want 0", got)
	}
	if got, _ := newSpawnSelection(nil, nil, 0.12).SelectSpawner(session); got != 1 {
		t.Errorf("random 0.12: spawner %d, want 1", got)
	}
}

// Игрок рядом со спаунером почти исключает его из выбора
func TestEnemySpawnSelection_NearPlayerPenalty(t *testing.T) {
	session := session_entities.NewStageSessionEntity()
	player := newPlayerAt(12*16, 16)

	// Веса 1, 1, 0.15: правый спаунер получает только хвост отрезка
	selection := newSpawnSelection([]*types.TankEntity{player}, nil, 0.9)
	if got, _ := selection.SelectSpawner(session); got != 1 {
		t.Errorf("random 0.9: spawner %d, want 1", got)
	}

	// Игрок далеко — штрафа нет
	far := newPlayerAt(12*16, 12*16)
	selection = newSpawnSelection([]*types.TankEntity{far}, nil, 0.9)
	if got, _ := selection.SelectSpawner(session); got != 2 {
		t.Errorf("far player: spawner %d, want 2", got)
	}
}
