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
	HitTest(pos types.Position) types.LevelSelectHit
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
		inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return s.startTransition()
	}

	return s.handleTap()
}

func (s *LevelSelectState) Draw(screen *ebiten.Image) {
	view := s.levelSelectUseCases.BuildView(s.selector, s.levels)
	view.TouchActive = s.touchControls.IsTouchActive()
	view.QuitAvailable = s.quitAvailable
	s.renderer.Draw(screen, view)
}

func (s *LevelSelectState) handleNavigation() {
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) ||
		inpututil.IsKeyJustPressed(ebiten.KeyA) {
		s.levelSelectUseCases.MoveLevel(s.selector, -1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) ||
		inpututil.IsKeyJustPressed(ebiten.KeyD) {
		s.levelSelectUseCases.MoveLevel(s.selector, 1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW) {
		s.levelSelectUseCases.MovePack(s.selector, -1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) ||
		inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.levelSelectUseCases.MovePack(s.selector, 1)
	}
}

// handleTap — тап по ячейке выбирает уровень, повторный тап по
// выбранной запускает его; стрелки листают пачки, нижняя полоса —
// запуск
func (s *LevelSelectState) handleTap() types.StateTransition {
	pos, ok := s.touchControls.TapJustPressed()
	if !ok {
		return types.StateTransition{}
	}

	hit := s.renderer.HitTest(pos)
	switch hit.Kind {
	case types.LevelSelectHitPrevPack:
		s.levelSelectUseCases.MovePack(s.selector, -1)
	case types.LevelSelectHitNextPack:
		s.levelSelectUseCases.MovePack(s.selector, 1)
	case types.LevelSelectHitLevel:
		if hit.Position == s.selector.Position {
			return s.startTransition()
		}
		s.levelSelectUseCases.SetPosition(s.selector, hit.Position)
	case types.LevelSelectHitStart:
		return s.startTransition()
	}

	return types.StateTransition{}
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
