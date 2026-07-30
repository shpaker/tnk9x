package app

import (
	"github.com/shpaker/tnk9x/internal/adapters/curtain"
	"github.com/shpaker/tnk9x/internal/adapters/game_over"
	"github.com/shpaker/tnk9x/internal/adapters/score"
	"github.com/shpaker/tnk9x/internal/adapters/title"
	"github.com/shpaker/tnk9x/internal/states"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// newTitleState собирает титульный экран из долгоживущих
// зависимостей приложения
func (app *App) newTitleState() *states.TitleState {
	rendererAdapter := title.NewTitleRendererAdapter(
		title.TitleRendererDependencies{
			FontFace:         app.textFace,
			HUDFontFace:      app.hudTextFace,
			TitleFontSize:    int(app.config.GetTitleFontSize()),
			RegularFontSize:  int(app.config.GetRegularFontSize()),
			SubtitleFontSize: int(app.config.GetSubtitleFontSize()),
			GameTitle:        app.config.GetGameTitle(),
		},
	)
	return states.NewTitleState(rendererAdapter, app.session.RunSession())
}

// newCurtainState собирает шторку «STAGE N»; allowSelect включает
// выбор этапа при старте нового забега
func (app *App) newCurtainState(
	allowSelect bool,
) (*states.StageCurtainState, error) {
	levelsCount, err := app.mapsRepository.GetLevelsCount()
	if err != nil {
		return nil, err
	}

	selector := types.NewStageSelector(uint(levelsCount))
	selector.CurrentStage = app.session.RunSession().GetStage()

	return states.NewStageCurtainState(
		selector,
		use_cases.NewStageSelectorUseCases(),
		curtain.NewCurtainRendererAdapter(app.hudTextFace),
		allowSelect,
	), nil
}

// newScoreState собирает экран итогов этапа; следующий этап
// вычисляется с цикличным переходом после последнего
func (app *App) newScoreState(
	stageWon bool,
) (*states.ScoreState, error) {
	levelsCount, err := app.mapsRepository.GetLevelsCount()
	if err != nil {
		return nil, err
	}

	runSession := app.session.RunSession()
	nextStage := runSession.GetStage() + 1
	if levelsCount > 0 && nextStage > uint(levelsCount) {
		nextStage = 1
	}

	return states.NewScoreState(
		score.NewScoreRendererAdapter(app.hudTextFace),
		runSession,
		app.soundAdapter,
		stageWon,
		nextStage,
	), nil
}

func (app *App) newGameOverState() *states.GameOverState {
	return states.NewGameOverState(
		game_over.NewGameOverRendererAdapter(
			app.textFace,
			int(app.config.GetTitleFontSize()),
		),
	)
}
