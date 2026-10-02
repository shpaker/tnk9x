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
	// HitBack — тап попал в экранную кнопку выхода в главное меню
	// последней отрисовки
	HitBack(position types.Position) bool
	// HitLevel — позиция ячейки уровня последней отрисовки под точкой
	HitLevel(position types.Position) (int, bool)
	// HitPack — стрелка пачки под точкой: -1 или +1
	HitPack(position types.Position) (int, bool)
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
	if s.menuInput.Back() || s.backTapped() {
		return types.StateTransition{Target: types.TransitionToMainMenu}
	}

	s.handleNavigation()

	if s.menuInput.Confirmed() || s.handlePointer() {
		return s.startTransition()
	}

	return types.StateTransition{}
}

func (s *LevelSelectState) Draw(screen *ebiten.Image) {
	view := s.levelSelectUseCases.BuildView(s.selector, s.levels)
	view.TouchActive = s.menuInput.IsTouchActive()
	view.PointerActive = s.menuInput.IsPointerActive()
	view.PlayerCount = s.settings.GetPlayers()
	s.renderer.Draw(screen, view)
}

// backTapped — тап по экранной кнопке выхода в главное меню
func (s *LevelSelectState) backTapped() bool {
	position, tapped := s.menuInput.Tapped()
	return tapped && s.renderer.HitBack(position)
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

// handlePointer: наведение выбирает уровень, клик по стрелкам листает
// пачки, клик по ячейке выбирает уровень; true — уровень выбран
// кликом и его пора запускать
func (s *LevelSelectState) handlePointer() bool {
	if position, pointed := s.menuInput.Pointed(); pointed {
		if level, ok := s.renderer.HitLevel(position); ok {
			s.levelSelectUseCases.SetPosition(s.selector, level)
		}
	}
	position, tapped := s.menuInput.Tapped()
	if !tapped {
		return false
	}
	if step, ok := s.renderer.HitPack(position); ok {
		s.levelSelectUseCases.MovePack(s.selector, step)
		return false
	}
	level, ok := s.renderer.HitLevel(position)
	if ok {
		s.levelSelectUseCases.SetPosition(s.selector, level)
	}
	return ok
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
