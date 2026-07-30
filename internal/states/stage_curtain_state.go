package states

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk9x/internal/adapters/curtain"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// curtainAutoDelayTicks — сколько шторка висит перед автопереходом
// на этап (~2 секунды при 60 TPS)
const curtainAutoDelayTicks = 120

// StageCurtainState — экран «STAGE N» перед этапом. При старте нового
// забега (allowSelect) стрелки листают этап, Enter начинает игру;
// между этапами шторка показывается на время и уходит сама.
type StageCurtainState struct {
	selector         *types.StageSelectorEntity
	selectorUseCases interfaces.IStageSelectorUseCases
	rendererAdapter  *curtain.CurtainRendererAdapter

	allowSelect bool
	ticks       int
}

func NewStageCurtainState(
	selector *types.StageSelectorEntity,
	selectorUseCases interfaces.IStageSelectorUseCases,
	rendererAdapter *curtain.CurtainRendererAdapter,
	allowSelect bool,
) *StageCurtainState {
	return &StageCurtainState{
		selector:         selector,
		selectorUseCases: selectorUseCases,
		rendererAdapter:  rendererAdapter,
		allowSelect:      allowSelect,
	}
}

func (s *StageCurtainState) Update() types.StateTransition {
	if s.allowSelect {
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) ||
			inpututil.IsKeyJustPressed(ebiten.KeyA) {
			s.selectorUseCases.Previous(s.selector)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) ||
			inpututil.IsKeyJustPressed(ebiten.KeyD) {
			s.selectorUseCases.Next(s.selector)
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			return types.StateTransition{
				Target: types.TransitionToStage,
				Level:  s.selectorUseCases.Select(s.selector),
			}
		}
		return types.StateTransition{}
	}

	s.ticks++
	if s.ticks >= curtainAutoDelayTicks {
		return types.StateTransition{
			Target: types.TransitionToStage,
			Level:  s.selector.CurrentStage,
		}
	}
	return types.StateTransition{}
}

func (s *StageCurtainState) Draw(screen *ebiten.Image) {
	s.rendererAdapter.DrawAll(
		screen,
		s.selectorUseCases.String(s.selector),
	)
}
