package use_cases

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk25/internal/types"
	"github.com/shpaker/tnk25/internal/types/session_entities"
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
		fmt.Sprintf(
			"Player1 Lives: %d/%d",
			info.Player1Lives,
			info.Player1InitialLives,
		),
		fmt.Sprintf(
			"Player2 Lives: %d/%d",
			info.Player2Lives,
			info.Player2InitialLives,
		),
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

	data.Player1Lives = stageSession.GetPlayerLives(types.PlayerTankNumPlayer1)
	data.Player1InitialLives = stageSession.GetPlayerInitialLives(
		types.PlayerTankNumPlayer1,
	)
	data.Player2Lives = stageSession.GetPlayerLives(types.PlayerTankNumPlayer2)
	data.Player2InitialLives = stageSession.GetPlayerInitialLives(
		types.PlayerTankNumPlayer2,
	)
	data.TotalEnemies = stageSession.GetTotalEnemies()
	data.RemainingEnemies = stageSession.GetRemainingEnemies()

	return data
}
