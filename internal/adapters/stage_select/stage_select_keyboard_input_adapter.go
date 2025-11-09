package stage_select

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// StageSelectKeyboardInputAdapter адаптер для обработки пользовательского ввода при выборе уровня
type StageSelectKeyboardInputAdapter struct {
	selector         *types.StageSelectorEntity
	selectorUseCases *use_cases.StageSelectorUseCases
	previousButton   ebiten.Key
	nextButton       ebiten.Key
	enterButton      ebiten.Key
}

// NewStageSelectKeyboardInputAdapter создает новый экземпляр StageSelectKeyboardInputAdapter
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

// Update обрабатывает пользовательский ввод
func (a *StageSelectKeyboardInputAdapter) Update(dt float64) {
	if a.selector == nil || a.selectorUseCases == nil {
		return
	}

	// Обрабатываем нажатия клавиш
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

	// Обрабатываем подтверждение выбора (Enter)
	if inpututil.IsKeyJustPressed(a.enterButton) {
		a.selectorUseCases.Select(a.selector)
	}
}
