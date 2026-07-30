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
		// Лестница звёзд как в оригинале: 1 — быстрые пули,
		// 2 — две пули, 3 — усиленные пули; скорость танка не меняется
		{"игрок 0", false, 0, 0, 32.0, false, 120.0, 1},
		{"игрок 1", false, 1, 1, 32.0, false, 150.0, 1},
		{"игрок 2", false, 2, 2, 32.0, false, 150.0, 2},
		{"игрок 3", false, 3, 3, 32.0, true, 150.0, 2},
		{"враг 0", true, 0, 0, 32.0, false, 120.0, 1},
		{"враг 1", true, 1, 1, 48.0, false, 120.0, 1},
		{"враг 2", true, 2, 2, 32.0, false, 180.0, 1},
		{"враг 3", true, 3, 3, 32.0, false, 120.0, 1},
		// Уровень выше 3 обрезается до 3
		{"игрок 4 -> 3", false, 4, 3, 32.0, true, 150.0, 2},
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
