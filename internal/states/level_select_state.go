package states

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// LevelSelectRenderer — контракт рендера экрана выбора уровня,
// определён у потребителя
type LevelSelectRenderer interface {
	Draw(screen *ebiten.Image, view types.LevelSelectViewData)
}

// LevelSelectStateDependencies — готовый граф зависимостей экрана
// выбора уровня; собирается composition root'ом, все поля обязательны
type LevelSelectStateDependencies struct {
	// Use Cases — кампания текущего режима (1P или 2P)
	LevelSelectUseCases interfaces.ILevelSelectUseCases

	// Adapters
	Renderer  LevelSelectRenderer
	MenuInput interfaces.IMenuInputAdapter

	// Entities
	Settings *types.SettingsEntity
	// Levels — разобранные уровни кампании для превью
	Levels map[int]*types.LevelEntity

	// LastLevel — уровень под курсором при открытии экрана
	LastLevel int
}

// LevelSelectState — экран выбора уровня кампании текущего режима:
// пачки по строкам, превью выбранной карты, звёзды и замки; назад —
// в главное меню
type LevelSelectState struct {
	// Use Cases
	levelSelectUseCases interfaces.ILevelSelectUseCases
	// Adapters
	renderer  LevelSelectRenderer
	menuInput interfaces.IMenuInputAdapter
	// Entities
	settings *types.SettingsEntity
	selector *types.LevelSelectorEntity
	levels   map[int]*types.LevelEntity
}

func NewLevelSelectState(deps LevelSelectStateDependencies) *LevelSelectState {
	return &LevelSelectState{
		levelSelectUseCases: deps.LevelSelectUseCases,
		renderer:            deps.Renderer,
		menuInput:           deps.MenuInput,
		settings:            deps.Settings,
		selector: deps.LevelSelectUseCases.NewSelector(
			deps.LastLevel,
		),
		levels: deps.Levels,
	}
}

func (s *LevelSelectState) Update() types.StateTransition {
	if s.menuInput.Back() {
		return types.StateTransition{Target: types.TransitionToMainMenu}
	}

	s.handleNavigation()

	if s.menuInput.Confirmed() {
		return s.startTransition()
	}

	return types.StateTransition{}
}

func (s *LevelSelectState) Draw(screen *ebiten.Image) {
	view := s.levelSelectUseCases.BuildView(s.selector, s.levels)
	view.TouchActive = s.menuInput.IsTouchActive()
	view.PlayerCount = s.settings.GetPlayers()
	s.renderer.Draw(screen, view)
}

// handleNavigation — влево-вправо листают уровни, вверх-вниз — пачки
func (s *LevelSelectState) handleNavigation() {
	if step := s.menuInput.SideStep(); step != 0 {
		s.levelSelectUseCases.MoveLevel(s.selector, step)
	}
	moveUp, moveDown := s.menuInput.Steps()
	if moveUp {
		s.levelSelectUseCases.MovePack(s.selector, -1)
	}
	if moveDown {
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
