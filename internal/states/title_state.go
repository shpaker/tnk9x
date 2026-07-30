package states

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk9x/internal/adapters/title"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

// TitleState — титульный экран: выбор 1/2 игроков, показ HI-SCORE.
// Enter начинает новый забег и ведёт на шторку выбора этапа.
type TitleState struct {
	rendererAdapter *title.TitleRendererAdapter
	runSession      *session_entities.RunSessionEntity

	twoPlayersSelected bool
}

func NewTitleState(
	rendererAdapter *title.TitleRendererAdapter,
	runSession *session_entities.RunSessionEntity,
) *TitleState {
	return &TitleState{
		rendererAdapter: rendererAdapter,
		runSession:      runSession,
	}
}

func (s *TitleState) Update() types.StateTransition {
	moveUp := inpututil.IsKeyJustPressed(ebiten.KeyUp) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW)
	moveDown := inpututil.IsKeyJustPressed(ebiten.KeyDown) ||
		inpututil.IsKeyJustPressed(ebiten.KeyS)

	if moveUp {
		s.twoPlayersSelected = false
	}
	if moveDown {
		s.twoPlayersSelected = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		playerCount := uint(1)
		if s.twoPlayersSelected {
			playerCount = 2
		}
		return types.StateTransition{
			Target:      types.TransitionToCurtain,
			Level:       1,
			PlayerCount: playerCount,
			NewRun:      true,
		}
	}

	return types.StateTransition{}
}

func (s *TitleState) Draw(screen *ebiten.Image) {
	hiScore := uint(0)
	if s.runSession != nil {
		hiScore = s.runSession.GetHiScore()
	}
	s.rendererAdapter.DrawAll(screen, types.TitleViewData{
		TwoPlayersSelected: s.twoPlayersSelected,
		HiScore:            hiScore,
	})
}
