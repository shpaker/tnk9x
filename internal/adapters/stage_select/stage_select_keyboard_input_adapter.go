package stage_select

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

type StageSelectKeyboardInputAdapter struct {
	selector         *types.StageSelectorEntity
	selectorUseCases *use_cases.StageSelectorUseCases
	previousButton   ebiten.Key
	nextButton       ebiten.Key
	enterButton      ebiten.Key
}

func NewStageSelectKeyboardInputAdapter(
	selector *types.StageSelectorEntity,
	selectorUseCases *use_cases.StageSelectorUseCases,
	previousButton ebiten.Key,
	nextButton ebiten.Key,
	enterButton ebiten.Key,
) *StageSelectKeyboardInputAdapter {
	return &StageSelectKeyboardInputAdapter{
		selector:         selector,
		selectorUseCases: selectorUseCases,
		previousButton:   previousButton,
		nextButton:       nextButton,
		enterButton:      enterButton,
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

	if inpututil.IsKeyJustPressed(a.enterButton) {
		a.selectorUseCases.Select(a.selector)
	}
}
