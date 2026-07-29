package use_cases

import (
	"fmt"
	"strings"

	"github.com/shpaker/tnk9x/internal/types"
)

type DebugUseCases struct {
	version string
}

func NewDebugUseCases(version string) *DebugUseCases {
	return &DebugUseCases{
		version: version,
	}
}

// BuildDebugInfo форматирует переданные метрики в текст HUD;
// сбор данных — забота вызывающего слоя
func (uc *DebugUseCases) BuildDebugInfo(data types.DebugInfoData) string {
	version := uc.version
	if version == "" {
		version = "dev"
	}

	debugLines := []string{
		fmt.Sprintf("Version: %s", version),
		fmt.Sprintf("FPS: %.2f", data.FPS),
		fmt.Sprintf("TPS: %.2f", data.TPS),
		fmt.Sprintf(
			"Player1 Lives: %d/%d",
			data.Player1Lives,
			data.Player1InitialLives,
		),
		fmt.Sprintf(
			"Player2 Lives: %d/%d",
			data.Player2Lives,
			data.Player2InitialLives,
		),
		fmt.Sprintf("Enemies: %d/%d", data.RemainingEnemies, data.TotalEnemies),
	}

	return strings.Join(debugLines, "\n")
}
