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
	upButton         ebiten.Key
	downButton       ebiten.Key
	enterButton      ebiten.Key
}

// NewStageSelectKeyboardInputAdapter создает новый экземпляр StageSelectKeyboardInputAdapter
func NewStageSelectKeyboardInputAdapter(
	selector *types.StageSelectorEntity,
	selectorUseCases *use_cases.StageSelectorUseCases,
	upButton ebiten.Key,
	downButton ebiten.Key,
	enterButton ebiten.Key,
) *StageSelectKeyboardInputAdapter {
	return &StageSelectKeyboardInputAdapter{
		selector:         selector,
		selectorUseCases: selectorUseCases,
		upButton:         upButton,
		downButton:       downButton,
		enterButton:      enterButton,
	}
}

// Update обрабатывает пользовательский ввод
func (a *StageSelectKeyboardInputAdapter) Update(dt float64) {
	if a.selector == nil || a.selectorUseCases == nil {
		return
	}

	// Обрабатываем нажатия клавиш
	if inpututil.IsKeyJustPressed(a.upButton) {
		a.selectorUseCases.Previous(a.selector)
	}

	if inpututil.IsKeyJustPressed(a.downButton) {
		a.selectorUseCases.Next(a.selector)
	}

	// Обрабатываем подтверждение выбора (Enter)
	if inpututil.IsKeyJustPressed(a.enterButton) {
		a.selectorUseCases.Select(a.selector)
	}
}
