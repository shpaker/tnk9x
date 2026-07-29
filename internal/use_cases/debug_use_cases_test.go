package use_cases_test

import (
	"strings"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

func newDebugTestEnv(version string) *use_cases.DebugUseCases {
	return use_cases.NewDebugUseCases(version)
}

// Полный HUD: шесть строк с версией, метриками движка и прогрессом
func TestDebugUseCases_BuildDebugInfo(t *testing.T) {
	debugUC := newDebugTestEnv("1.2.3")
	data := types.DebugInfoData{
		FPS:                 59.9,
		TPS:                 60,
		Player1Lives:        2,
		Player1InitialLives: 3,
		Player2Lives:        0,
		Player2InitialLives: 3,
		TotalEnemies:        20,
		RemainingEnemies:    5,
	}

	want := strings.Join([]string{
		"Version: 1.2.3",
		"FPS: 59.90",
		"TPS: 60.00",
		"Player1 Lives: 2/3",
		"Player2 Lives: 0/3",
		"Enemies: 5/20",
	}, "\n")

	if got := debugUC.BuildDebugInfo(data); got != want {
		t.Errorf("HUD:\n%s\nожидалось:\n%s", got, want)
	}
}

// Пустая версия сборки подменяется на "dev"
func TestDebugUseCases_BuildDebugInfo_VersionFallback(t *testing.T) {
	debugUC := newDebugTestEnv("")

	got := debugUC.BuildDebugInfo(types.DebugInfoData{})
	lines := strings.Split(got, "\n")
	if len(lines) != 6 {
		t.Fatalf("строк %d, ожидалось 6", len(lines))
	}
	if lines[0] != "Version: dev" {
		t.Errorf("первая строка %q, ожидалась \"Version: dev\"", lines[0])
	}
}
