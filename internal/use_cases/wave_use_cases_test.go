package use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/use_cases"
)

// Интервал спавна по формуле NES: 190 - 4*этап - 20*(игроки-1),
// с минимумом и нормализацией нулевых аргументов
func TestWaveUseCases_SpawnDelayTicks(t *testing.T) {
	uc := use_cases.NewWaveUseCases()

	tests := []struct {
		name        string
		stage       uint
		playerCount uint
		want        uint
	}{
		{"этап 1, один игрок", 1, 1, 186},
		{"этап 10, один игрок", 10, 1, 150},
		{"этап 1, два игрока", 1, 2, 166},
		{"этап 35, два игрока", 35, 2, 30},
		{"поздний цикл упирается в минимум", 45, 2, 30},
		{"нулевые аргументы нормализуются", 0, 0, 186},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uc.SpawnDelayTicks(tt.stage, tt.playerCount)
			if got != tt.want {
				t.Errorf(
					"SpawnDelayTicks(%d, %d) = %d, ожидалось %d",
					tt.stage,
					tt.playerCount,
					got,
					tt.want,
				)
			}
		})
	}
}
