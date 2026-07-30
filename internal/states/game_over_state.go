package states

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk9x/internal/adapters/game_over"
	"github.com/shpaker/tnk9x/internal/types"
)

// gameOverDelayTicks — сколько экран висит до автоперехода на титул
const gameOverDelayTicks = 240

// GameOverState — экран «GAME OVER», ведёт обратно на титул
type GameOverState struct {
	rendererAdapter *game_over.GameOverRendererAdapter
	ticks           int
}

func NewGameOverState(
	rendererAdapter *game_over.GameOverRendererAdapter,
) *GameOverState {
	return &GameOverState{rendererAdapter: rendererAdapter}
}

func (s *GameOverState) Update() types.StateTransition {
	s.ticks++

	anyKeyPressed := len(inpututil.AppendJustPressedKeys(nil)) > 0
	if anyKeyPressed || s.ticks >= gameOverDelayTicks {
		return types.StateTransition{Target: types.TransitionToTitle}
	}
	return types.StateTransition{}
}

func (s *GameOverState) Draw(screen *ebiten.Image) {
	s.rendererAdapter.DrawAll(screen)
}
