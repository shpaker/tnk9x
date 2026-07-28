package use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/use_cases"
)

func TestSpecsUseCases_GetTankSpecs(t *testing.T) {
	uc := use_cases.NewSpecsUseCases()

	tests := []struct {
		name             string
		isEnemy          bool
		level            uint
		wantLevel        uint
		wantSpeed        float64
		wantReinforced   bool
		wantBulletsSpeed float64
		wantBulletsLimit uint
	}{
		{"игрок 0", false, 0, 0, 32.0, false, 120.0, 1},
		{"игрок 1", false, 1, 1, 32.0, false, 150.0, 1},
		{"игрок 2", false, 2, 2, 32.0, true, 150.0, 1},
		{"игрок 3", false, 3, 3, 40.0, true, 150.0, 2},
		{"враг 0", true, 0, 0, 32.0, false, 120.0, 1},
		{"враг 1", true, 1, 1, 48.0, false, 120.0, 1},
		{"враг 2", true, 2, 2, 32.0, false, 180.0, 1},
		{"враг 3", true, 3, 3, 32.0, false, 120.0, 1},
		// Уровень выше 3 обрезается до 3
		{"игрок 4 -> 3", false, 4, 3, 40.0, true, 150.0, 2},
		{"враг 100 -> 3", true, 100, 3, 32.0, false, 120.0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specs := uc.GetTankSpecs(tt.isEnemy, tt.level)
			if specs == nil {
				t.Fatal("specs == nil")
			}
			if specs.GetLevel() != tt.wantLevel {
				t.Errorf(
					"level = %d, ожидалось %d",
					specs.GetLevel(),
					tt.wantLevel,
				)
			}
			if specs.GetSpeed() != tt.wantSpeed {
				t.Errorf(
					"speed = %v, ожидалось %v",
					specs.GetSpeed(),
					tt.wantSpeed,
				)
			}
			if specs.GetBulletsReinforced() != tt.wantReinforced {
				t.Errorf(
					"reinforced = %v, ожидалось %v",
					specs.GetBulletsReinforced(),
					tt.wantReinforced,
				)
			}
			if specs.GetBulletsSpeed() != tt.wantBulletsSpeed {
				t.Errorf(
					"bulletsSpeed = %v, ожидалось %v",
					specs.GetBulletsSpeed(),
					tt.wantBulletsSpeed,
				)
			}
			if specs.GetBulletsLimit() != tt.wantBulletsLimit {
				t.Errorf(
					"bulletsLimit = %d, ожидалось %d",
					specs.GetBulletsLimit(),
					tt.wantBulletsLimit,
				)
			}
		})
	}
}

// Детерминированные ветки: первые танки всегда 0 уровня, танки 4-5 — уровня 1
func TestSpecsUseCases_GetEnemyLevelByRemainingCount_Deterministic(
	t *testing.T,
) {
	uc := use_cases.NewSpecsUseCases()

	tests := []struct {
		remaining uint
		want      uint
	}{
		{100, 0},
		{20, 0},
		{19, 0},
		{18, 0},
		{17, 1},
		{16, 1},
	}

	for _, tt := range tests {
		if got := uc.GetEnemyLevelByRemainingCount(tt.remaining); got != tt.want {
			t.Errorf(
				"remaining %d: level = %d, ожидалось %d",
				tt.remaining,
				got,
				tt.want,
			)
		}
	}
}

// Случайные ветки: проверяем принадлежность допустимому множеству уровней
func TestSpecsUseCases_GetEnemyLevelByRemainingCount_Random(t *testing.T) {
	uc := use_cases.NewSpecsUseCases()

	tests := []struct {
		name      string
		remaining []uint
		allowed   map[uint]bool
	}{
		{
			"оставшиеся 11-15: уровни 1 или 2",
			[]uint{11, 13, 15},
			map[uint]bool{1: true, 2: true},
		},
		{
			"оставшиеся 6-10: уровни 1, 2 или 3",
			[]uint{6, 8, 10},
			map[uint]bool{1: true, 2: true, 3: true},
		},
		{
			"оставшиеся 0-5: уровни 1, 2 или 3",
			[]uint{0, 1, 3, 5},
			map[uint]bool{1: true, 2: true, 3: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, remaining := range tt.remaining {
				for i := 0; i < 200; i++ {
					got := uc.GetEnemyLevelByRemainingCount(remaining)
					if !tt.allowed[got] {
						t.Fatalf(
							"remaining %d: недопустимый уровень %d",
							remaining,
							got,
						)
					}
				}
			}
		})
	}
}
