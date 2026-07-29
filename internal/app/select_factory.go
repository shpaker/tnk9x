package app

import (
	"github.com/shpaker/tnk9x/internal/states"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// newStageSelectState собирает состояние меню выбора уровня
// из долгоживущих зависимостей приложения
func (app *App) newStageSelectState() (*states.StageSelectState, error) {
	return states.NewStageSelectState(
		app.config,
		use_cases.NewStageSelectorUseCases(),
		app.textFace,
		app.mapsRepository,
	)
}
