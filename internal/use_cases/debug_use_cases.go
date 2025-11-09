package use_cases

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/types/session_entities"
)

// DebugUseCases формирует отладочную информацию
type DebugUseCases struct {
	config  interfaces.IConfigProvider
	session *session_entities.GameSessionEntity
}

// NewDebugUseCases создаёт UseCase для отладочной информации
func NewDebugUseCases(
	config interfaces.IConfigProvider,
	session *session_entities.GameSessionEntity,
) *DebugUseCases {
	return &DebugUseCases{
		config:  config,
		session: session,
	}
}

// BuildDebugInfo возвращает строку для вывода отладочной информации
func (uc *DebugUseCases) BuildDebugInfo() (string, bool) {
	if uc == nil || uc.config == nil || !uc.config.IsDebugEnabled() {
		return "", false
	}

	info := uc.collectDebugData()

	debugString := fmt.Sprintf(
		"FPS: %.2f\nTPS: %.2f\nLives: %d/%d\nEnemies: %d/%d",
		info.FPS,
		info.TPS,
		info.PlayerLives,
		info.PlayerInitialLives,
		info.RemainingEnemies,
		info.TotalEnemies,
	)

	return debugString, true
}

// collectDebugData собирает данные для отладки
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

	data.PlayerLives = stageSession.GetPlayer1Lives()
	data.PlayerInitialLives = stageSession.GetPlayer1InitialLives()
	data.TotalEnemies = stageSession.GetTotalEnemies()
	data.RemainingEnemies = stageSession.GetRemainingEnemies()

	return data
}
