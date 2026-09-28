package use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

func newLevelSelect() (
	*use_cases.LevelSelectUseCases,
	*use_cases.ProgressionUseCases,
) {
	progression, _ := newProgression()
	return use_cases.NewLevelSelectUseCases(progression), progression
}

// Курсор встаёт на последний сыгранный уровень, иначе — на последний
// открытый
func TestLevelSelect_NewSelector(t *testing.T) {
	levelSelect, progression := newLevelSelect()

	if selector := levelSelect.NewSelector(5); selector.PackIndex != 1 ||
		selector.Position != 1 {
		t.Errorf("last level 5: %+v", selector)
	}

	_ = progression.RecordResult(1, 1)
	if selector := levelSelect.NewSelector(0); selector.PackIndex != 0 ||
		selector.Position != 1 {
		t.Errorf("no last level: %+v, want level 2", selector)
	}
}

// Влево-вправо курсор идёт сквозь пачки и стоит на краях кампании
func TestLevelSelect_MoveLevel(t *testing.T) {
	levelSelect, _ := newLevelSelect()
	selector := &types.LevelSelectorEntity{}

	levelSelect.MoveLevel(selector, -1)
	if selector.PackIndex != 0 || selector.Position != 0 {
		t.Errorf("before the first level: %+v", selector)
	}

	levelSelect.MoveLevel(selector, 3)
	if selector.PackIndex != 1 || selector.Position != 0 {
		t.Errorf("three steps: %+v, want level 4", selector)
	}

	levelSelect.MoveLevel(selector, 10)
	if selector.PackIndex != 1 || selector.Position != 1 {
		t.Errorf("past the end: %+v, want level 5", selector)
	}

	levelSelect.MoveLevel(selector, -2)
	if selector.PackIndex != 0 || selector.Position != 2 {
		t.Errorf("two steps back: %+v, want level 3", selector)
	}
}

// Смена пачки сохраняет позицию, насколько позволяет длина пачки
func TestLevelSelect_MovePackAndSetPosition(t *testing.T) {
	levelSelect, _ := newLevelSelect()
	selector := &types.LevelSelectorEntity{Position: 2}

	levelSelect.MovePack(selector, 1)
	if selector.PackIndex != 1 || selector.Position != 1 {
		t.Errorf("next pack: %+v", selector)
	}
	levelSelect.MovePack(selector, 5)
	if selector.PackIndex != 1 {
		t.Errorf("past the last pack: %+v", selector)
	}

	levelSelect.SetPosition(selector, 0)
	if level, unlocked := levelSelect.SelectedLevel(selector); level != 4 ||
		unlocked {
		t.Errorf("selected %d unlocked=%v, want locked 4", level, unlocked)
	}
}

func TestLevelSelect_BuildView(t *testing.T) {
	levelSelect, progression := newLevelSelect()
	_ = progression.RecordResult(1, 2)

	level := types.NewLevelEntity(2, "SECOND", 4, 1200, []types.WaveSpec{
		{Tanks: []types.WaveTank{{Level: 0}, {Level: 3}}},
	}, false, nil)
	view := levelSelect.BuildView(
		&types.LevelSelectorEntity{Position: 1},
		map[int]*types.LevelEntity{2: level},
	)

	if view.PackName != "ONE" || view.PacksCount != 2 ||
		view.CampaignMaxStars != 15 {
		t.Errorf("pack view %+v", view)
	}
	if len(view.Entries) != 3 || view.Entries[0].Stars != 2 ||
		!view.Entries[1].Unlocked || view.Entries[2].Unlocked {
		t.Errorf("entries %+v", view.Entries)
	}
	if view.Level != level || !view.LevelUnlocked ||
		view.EnemyCounts != [4]uint{1, 0, 0, 1} ||
		view.Time3StarTicks != 1200 {
		t.Errorf("selected level view %+v", view)
	}
}
