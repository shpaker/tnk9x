package states

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// LevelSelectRenderer — контракт рендера экрана выбора уровня,
// определён у потребителя
type LevelSelectRenderer interface {
	Draw(screen *ebiten.Image, view types.LevelSelectViewData)
}

// LevelSelectState — экран выбора уровня кампании: пачки по строкам,
// превью выбранной карты, звёзды и замки
type LevelSelectState struct {
	// Use Cases
	levelSelectUseCases interfaces.ILevelSelectUseCases
	// Adapters
	renderer      LevelSelectRenderer
	touchControls interfaces.ITouchControlsAdapter
	// Entities
	selector *types.LevelSelectorEntity
	// levels — разобранные уровни кампании для превью
	levels map[int]*types.LevelEntity
	// quitAvailable — выход по ESC есть только там, где приложение
	// может завершиться (десктоп)
	quitAvailable bool
}

func NewLevelSelectState(
	levelSelectUseCases interfaces.ILevelSelectUseCases,
	renderer LevelSelectRenderer,
	touchControls interfaces.ITouchControlsAdapter,
	levels map[int]*types.LevelEntity,
	lastLevel int,
	quitAvailable bool,
) *LevelSelectState {
	return &LevelSelectState{
		levelSelectUseCases: levelSelectUseCases,
		renderer:            renderer,
		touchControls:       touchControls,
		selector:            levelSelectUseCases.NewSelector(lastLevel),
		levels:              levels,
		quitAvailable:       quitAvailable,
	}
}

func (s *LevelSelectState) Update() types.StateTransition {
	if s.quitAvailable && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return types.StateTransition{Target: types.TransitionToQuit}
	}

	s.handleNavigation()

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		s.touchControls.FireJustPressed() {
		return s.startTransition()
	}

	return types.StateTransition{}
}

func (s *LevelSelectState) Draw(screen *ebiten.Image) {
	view := s.levelSelectUseCases.BuildView(s.selector, s.levels)
	view.TouchActive = s.touchControls.IsTouchActive()
	view.QuitAvailable = s.quitAvailable
	s.renderer.Draw(screen, view)
}

// handleNavigation — стрелки клавиатуры и крестовина: влево-вправо
// листают уровни, вверх-вниз — пачки
func (s *LevelSelectState) handleNavigation() {
	dpad, pressed := s.touchControls.DPadJustPressed()
	step := func(keys [2]ebiten.Key, direction types.Direction) bool {
		return inpututil.IsKeyJustPressed(keys[0]) ||
			inpututil.IsKeyJustPressed(keys[1]) ||
			(pressed && dpad == direction)
	}
	if step([2]ebiten.Key{ebiten.KeyLeft, ebiten.KeyA}, types.DirectionLeft) {
		s.levelSelectUseCases.MoveLevel(s.selector, -1)
	}
	if step([2]ebiten.Key{ebiten.KeyRight, ebiten.KeyD}, types.DirectionRight) {
		s.levelSelectUseCases.MoveLevel(s.selector, 1)
	}
	if step([2]ebiten.Key{ebiten.KeyUp, ebiten.KeyW}, types.DirectionUp) {
		s.levelSelectUseCases.MovePack(s.selector, -1)
	}
	if step([2]ebiten.Key{ebiten.KeyDown, ebiten.KeyS}, types.DirectionDown) {
		s.levelSelectUseCases.MovePack(s.selector, 1)
	}
}

// startTransition запускает выбранный уровень, если он открыт
func (s *LevelSelectState) startTransition() types.StateTransition {
	level, unlocked := s.levelSelectUseCases.SelectedLevel(s.selector)
	if !unlocked {
		return types.StateTransition{}
	}

	return types.StateTransition{
		Target: types.TransitionToStage,
		Level:  uint(level),
	}
}
