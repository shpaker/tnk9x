package stage_select

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IInputAdapter = (*StageSelectKeyboardInputAdapter)(nil)

type StageSelectKeyboardInputAdapter struct {
	selector         *types.StageSelectorEntity
	selectorUseCases interfaces.IStageSelectorUseCases
	previousButton   ebiten.Key
	nextButton       ebiten.Key
}

func NewStageSelectKeyboardInputAdapter(
	selector *types.StageSelectorEntity,
	selectorUseCases interfaces.IStageSelectorUseCases,
	previousButton ebiten.Key,
	nextButton ebiten.Key,
) *StageSelectKeyboardInputAdapter {
	return &StageSelectKeyboardInputAdapter{
		selector:         selector,
		selectorUseCases: selectorUseCases,
		previousButton:   previousButton,
		nextButton:       nextButton,
	}
}

func (a *StageSelectKeyboardInputAdapter) Update(dt float64) {
	if a.selector == nil || a.selectorUseCases == nil {
		return
	}

	if inpututil.IsKeyJustPressed(a.previousButton) ||
		inpututil.IsKeyJustPressed(ebiten.KeyLeft) ||
		inpututil.IsKeyJustPressed(ebiten.KeyA) {
		a.selectorUseCases.Previous(a.selector)
	}

	if inpututil.IsKeyJustPressed(a.nextButton) ||
		inpututil.IsKeyJustPressed(ebiten.KeyRight) ||
		inpututil.IsKeyJustPressed(ebiten.KeyD) {
		a.selectorUseCases.Next(a.selector)
	}
}
