package use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

type mapTestEnv struct {
	blocks    types.MapBlocks
	spawnAt   []types.Position
	mapEntity *types.MapEntity
	mapUC     *use_cases.MapUseCases
}

func newMapTestEnv() *mapTestEnv {
	blocks := types.MapBlocks{
		brick(0, 0),
		brick(8, 0),
		brick(16, 0),
	}
	spawnAt := []types.Position{
		{X: 16, Y: 32},
		{X: 48, Y: 64},
	}
	mapEntity := types.NewMapEntity(
		types.Size{Width: 208, Height: 208},
		blocks,
		spawnAt,
	)

	return &mapTestEnv{
		blocks:    blocks,
		spawnAt:   spawnAt,
		mapEntity: mapEntity,
		mapUC:     use_cases.NewMapUseCases(mapEntity),
	}
}

func TestMapUseCases_GetBlocks(t *testing.T) {
	env := newMapTestEnv()

	got := env.mapUC.GetBlocks()
	if len(got) != len(env.blocks) {
		t.Fatalf("блоков %d, ожидалось %d", len(got), len(env.blocks))
	}
	for i, block := range got {
		if block != env.blocks[i] {
			t.Errorf("блок %d не совпадает", i)
		}
	}
}

func TestMapUseCases_RemoveBlock(t *testing.T) {
	env := newMapTestEnv()
	removed := env.blocks[1]

	if err := env.mapUC.RemoveBlock(removed); err != nil {
		t.Fatalf("удаление блока: %v", err)
	}

	got := env.mapUC.GetBlocks()
	if len(got) != 2 {
		t.Fatalf("блоков %d, ожидалось 2", len(got))
	}
	for _, block := range got {
		if block == removed {
			t.Error("блок не удалён из карты")
		}
	}

	// Повторное удаление и nil-блок не ошибочны
	if err := env.mapUC.RemoveBlock(removed); err != nil {
		t.Errorf("повторное удаление: %v", err)
	}
	if err := env.mapUC.RemoveBlock(nil); err != nil {
		t.Errorf("удаление nil-блока: %v", err)
	}
	if got := len(env.mapUC.GetBlocks()); got != 2 {
		t.Errorf("блоков %d, ожидалось 2", got)
	}
}

func TestMapUseCases_GetSizePx(t *testing.T) {
	env := newMapTestEnv()

	want := types.Size{Width: 208, Height: 208}
	if got := env.mapUC.GetSizePx(); got != want {
		t.Errorf("размер %v, ожидался %v", got, want)
	}
}

func TestMapUseCases_GetRandomBonusSpawnPosition(t *testing.T) {
	env := newMapTestEnv()
	allowed := make(map[types.Position]bool, len(env.spawnAt))
	for _, position := range env.spawnAt {
		allowed[position] = true
	}

	for i := 0; i < 50; i++ {
		if got := env.mapUC.GetRandomBonusSpawnPosition(); !allowed[got] {
			t.Fatalf("недопустимая позиция бонуса: %v", got)
		}
	}

	// Карта без свободных мест — позиция по умолчанию
	emptyUC := use_cases.NewMapUseCases(
		types.NewMapEntity(types.Size{}, nil, nil),
	)
	if got := emptyUC.GetRandomBonusSpawnPosition(); got != (types.Position{}) {
		t.Errorf("позиция %v, ожидалась нулевая", got)
	}
}

func TestMapUseCases_IsIceAt(t *testing.T) {
	ice := types.NewBlockEntity("ice", 16, 16, 8, &stubImageProvider{})
	mapUC := use_cases.NewMapUseCases(types.NewMapEntity(
		types.Size{Width: 208, Height: 208},
		types.MapBlocks{brick(0, 0), ice},
		nil,
	))

	if !mapUC.IsIceAt(types.Position{X: 20, Y: 20}) {
		t.Error("точка внутри льда: ожидалось true")
	}
	if !mapUC.IsIceAt(types.Position{X: 16, Y: 16}) {
		t.Error("левая/верхняя граница включительно: ожидалось true")
	}
	if mapUC.IsIceAt(types.Position{X: 24, Y: 20}) {
		t.Error("правая граница исключительно: ожидалось false")
	}
	if mapUC.IsIceAt(types.Position{X: 4, Y: 4}) {
		t.Error("кирпич не лёд: ожидалось false")
	}
	if use_cases.NewMapUseCases(nil).IsIceAt(types.Position{X: 20, Y: 20}) {
		t.Error("без карты: ожидалось false")
	}
}

// Без карты use case возвращает нейтральные значения
func TestMapUseCases_NilMapEntity(t *testing.T) {
	mapUC := use_cases.NewMapUseCases(nil)

	if got := mapUC.GetBlocks(); len(got) != 0 {
		t.Errorf("блоки без карты: %v", got)
	}
	if err := mapUC.RemoveBlock(brick(0, 0)); err != nil {
		t.Errorf("удаление без карты: %v", err)
	}
	if got := mapUC.GetSizePx(); got != (types.Size{}) {
		t.Errorf("размер без карты: %v", got)
	}
	if got := mapUC.GetRandomBonusSpawnPosition(); got != (types.Position{}) {
		t.Errorf("позиция без карты: %v", got)
	}
}
