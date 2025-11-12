package use_cases

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/types/session_entities"
)

type DebugUseCases struct {
	session *session_entities.GameSessionEntity
	version string
}

func NewDebugUseCases(
	session *session_entities.GameSessionEntity,
	version string,
) *DebugUseCases {
	return &DebugUseCases{
		session: session,
		version: version,
	}
}

func (uc *DebugUseCases) BuildDebugInfo() string {
	if uc == nil {
		return ""
	}
	info := uc.collectDebugData()

	version := uc.version
	if version == "" {
		version = "dev"
	}

	debugLines := []string{
		fmt.Sprintf("Version: %s", version),
		fmt.Sprintf("FPS: %.2f", info.FPS),
		fmt.Sprintf("TPS: %.2f", info.TPS),
		fmt.Sprintf("Lives: %d/%d", info.PlayerLives, info.PlayerInitialLives),
		fmt.Sprintf("Enemies: %d/%d", info.RemainingEnemies, info.TotalEnemies),
	}

	return strings.Join(debugLines, "\n")
}

func (uc *DebugUseCases) collectDebugData() types.DebugInfoData {
	data := types.DebugInfoData{
		FPS: ebiten.ActualFPS(),
		TPS: ebiten.ActualTPS(),
	}

	if uc.session == nil {
		return data
	}

	stageSession := uc.session.StageSession()
	if stageSession == nil {
		return data
	}

	data.PlayerLives = stageSession.GetPlayerLives(types.PlayerTankNumPlayer1)
	data.PlayerInitialLives = stageSession.GetPlayerInitialLives(
		types.PlayerTankNumPlayer1,
	)
	data.TotalEnemies = stageSession.GetTotalEnemies()
	data.RemainingEnemies = stageSession.GetRemainingEnemies()

	return data
}
