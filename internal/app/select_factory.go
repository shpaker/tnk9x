package app

import (
	"runtime"

	"github.com/shpaker/tnk9x/internal/states"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// newStageSelectState собирает состояние меню выбора уровня
// из долгоживущих зависимостей приложения
func (app *App) newStageSelectState() (*states.StageSelectState, error) {
	// В браузере ebiten.Termination заморозил бы canvas,
	// поэтому пункт QUIT есть только на десктопе
	return states.NewStageSelectState(
		app.config,
		use_cases.NewStageSelectorUseCases(),
		app.textFace,
		app.mapsRepository,
		app.touchControls,
		runtime.GOOS != "js",
	)
}
