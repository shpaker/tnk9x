package states

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk9x/internal/adapters/score"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

const (
	// scoreMinTicks — минимальное время показа итогов до реакции на ввод
	scoreMinTicks = 90

	// scoreAutoDelayTicks — автопереход дальше без нажатий (~5 секунд)
	scoreAutoDelayTicks = 300
)

// ScoreState — экран подсчёта очков после этапа. После победы ведёт
// на шторку следующего этапа, после поражения — на GAME OVER.
type ScoreState struct {
	rendererAdapter    *score.ScoreRendererAdapter
	runSession         *session_entities.RunSessionEntity
	soundPlayerAdapter interfaces.ISoundPlayerAdapter

	stageWon  bool
	nextStage uint
	ticks     int
}

func NewScoreState(
	rendererAdapter *score.ScoreRendererAdapter,
	runSession *session_entities.RunSessionEntity,
	soundPlayerAdapter interfaces.ISoundPlayerAdapter,
	stageWon bool,
	nextStage uint,
) *ScoreState {
	return &ScoreState{
		rendererAdapter:    rendererAdapter,
		runSession:         runSession,
		soundPlayerAdapter: soundPlayerAdapter,
		stageWon:           stageWon,
		nextStage:          nextStage,
	}
}

func (s *ScoreState) Update() types.StateTransition {
	s.ticks++

	if s.ticks == 1 {
		_ = s.soundPlayerAdapter.Play(types.SoundIDScore)
	}
	s.soundPlayerAdapter.Update()

	anyKeyPressed := s.ticks >= scoreMinTicks &&
		len(inpututil.AppendJustPressedKeys(nil)) > 0
	if !anyKeyPressed && s.ticks < scoreAutoDelayTicks {
		return types.StateTransition{}
	}

	if s.stageWon {
		return types.StateTransition{
			Target: types.TransitionToCurtain,
			Level:  s.nextStage,
		}
	}
	return types.StateTransition{Target: types.TransitionToGameOver}
}

func (s *ScoreState) Draw(screen *ebiten.Image) {
	view := types.ScoreViewData{
		StageNumber: s.runSession.GetStage(),
		PlayerCount: s.runSession.GetPlayerCount(),
		HiScore:     s.runSession.GetHiScore(),
		Player1: types.ScorePlayerViewData{
			Score: s.runSession.GetScore(types.PlayerTankNumPlayer1),
			Kills: s.runSession.GetStageKills(types.PlayerTankNumPlayer1),
		},
		Player2: types.ScorePlayerViewData{
			Score: s.runSession.GetScore(types.PlayerTankNumPlayer2),
			Kills: s.runSession.GetStageKills(types.PlayerTankNumPlayer2),
		},
	}
	s.rendererAdapter.DrawAll(screen, view)
}
