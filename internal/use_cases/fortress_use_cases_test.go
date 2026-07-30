package use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

func newFortressTestEnv() (
	*use_cases.FortressUseCases,
	*use_cases.MapUseCases,
	*session_entities.StageSessionEntity,
) {
	ring := []types.Position{
		{X: 88, Y: 184},
		{X: 96, Y: 184},
	}
	blocks := types.MapBlocks{
		types.NewBlockEntity("brick", 88, 184, 8, &stubImageProvider{}),
		types.NewBlockEntity("brick", 96, 184, 8, &stubImageProvider{}),
	}
	mapEntity := types.NewMapEntity(
		types.Size{Width: 208, Height: 208},
		blocks,
		nil,
	)
	mapUC := use_cases.NewMapUseCases(mapEntity)
	session := session_entities.NewStageSessionEntity(nil)

	return use_cases.NewFortressUseCases(
		mapUC,
		session,
		ring,
		8,
	), mapUC, session
}

func ringBlockTypes(mapUC *use_cases.MapUseCases) map[types.BlockType]int {
	counts := make(map[types.BlockType]int)
	for _, block := range mapUC.GetBlocks() {
		if block != nil && block.Data != nil {
			counts[block.Data.Name]++
		}
	}
	return counts
}

// Лопата: кольцо становится стальным, по истечении — снова кирпичным
func TestFortressUseCases_ApplyAndExpire(t *testing.T) {
	fortress, mapUC, session := newFortressTestEnv()

	fortress.Apply()

	counts := ringBlockTypes(mapUC)
	if counts[types.Steel] != 2 || counts[types.Brick] != 0 {
		t.Fatalf("после Apply блоки: %v, ожидалась сталь", counts)
	}
	if session.GetShovelTicks() == 0 {
		t.Fatal("отсчёт лопаты не запущен")
	}

	// Доводим отсчёт до конца
	for session.GetShovelTicks() > 0 {
		fortress.Update()
	}

	counts = ringBlockTypes(mapUC)
	if counts[types.Brick] != 2 || counts[types.Steel] != 0 {
		t.Errorf("после отката блоки: %v, ожидался кирпич", counts)
	}
}

// Кольцо строится заново даже поверх разрушенных тайлов
func TestFortressUseCases_RebuildsDestroyedRing(t *testing.T) {
	fortress, mapUC, session := newFortressTestEnv()

	// «Разрушаем» один тайл кольца до применения лопаты
	blocks := mapUC.GetBlocks()
	_ = mapUC.RemoveBlock(blocks[0])

	fortress.Apply()

	counts := ringBlockTypes(mapUC)
	if counts[types.Steel] != 2 {
		t.Errorf("кольцо не восстановлено сталью: %v", counts)
	}
	_ = session
}

// Update без активной лопаты не трогает карту
func TestFortressUseCases_UpdateIdle(t *testing.T) {
	fortress, mapUC, _ := newFortressTestEnv()

	fortress.Update()

	counts := ringBlockTypes(mapUC)
	if counts[types.Brick] != 2 {
		t.Errorf("карта изменена без лопаты: %v", counts)
	}
}
