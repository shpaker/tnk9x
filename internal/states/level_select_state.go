package states

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// LevelSelectRenderer — контракт рендера экрана выбора уровня,
// определён у потребителя
type LevelSelectRenderer interface {
	Draw(screen *ebiten.Image, view types.LevelSelectViewData)
	DrawMenu(screen *ebiten.Image, view types.LevelSelectMenuViewData)
}

// LevelSelectStateDependencies — готовый граф зависимостей экрана
// выбора уровня; собирается composition root'ом, все поля обязательны
type LevelSelectStateDependencies struct {
	// Use Cases
	LevelSelectUseCases interfaces.ILevelSelectUseCases
	SettingsUseCases    interfaces.ISettingsUseCases

	// Adapters
	Renderer  LevelSelectRenderer
	MenuInput interfaces.IMenuInputAdapter

	// Presentation
	SettingsOverlay *SettingsOverlay

	// Entities
	Settings *types.SettingsEntity
	// Levels — разобранные уровни кампании для превью
	Levels map[int]*types.LevelEntity

	// LastLevel — уровень под курсором при открытии экрана
	LastLevel int
	// QuitAvailable — выход из игры есть только там, где приложение
	// может завершиться (десктоп)
	QuitAvailable bool
}

// LevelSelectState — экран выбора уровня кампании: пачки по строкам,
// превью выбранной карты, звёзды и замки; Esc, Start или пауза
// открывают меню с выбором числа игроков, настройками и выходом
type LevelSelectState struct {
	// Use Cases
	levelSelectUseCases interfaces.ILevelSelectUseCases
	settingsUseCases    interfaces.ISettingsUseCases
	// Adapters
	renderer  LevelSelectRenderer
	menuInput interfaces.IMenuInputAdapter
	// Presentation
	settingsOverlay *SettingsOverlay
	// Entities
	settings *types.SettingsEntity
	selector *types.LevelSelectorEntity
	levels   map[int]*types.LevelEntity

	// Меню экрана
	menuItems []types.LevelSelectMenuItem
	menuIndex int
	menuOpen  bool
}

func NewLevelSelectState(deps LevelSelectStateDependencies) *LevelSelectState {
	menuItems := []types.LevelSelectMenuItem{
		types.LevelSelectMenuItemBack,
		types.LevelSelectMenuItemPlayers,
		types.LevelSelectMenuItemSettings,
	}
	if deps.QuitAvailable {
		menuItems = append(menuItems, types.LevelSelectMenuItemQuit)
	}

	return &LevelSelectState{
		levelSelectUseCases: deps.LevelSelectUseCases,
		settingsUseCases:    deps.SettingsUseCases,
		renderer:            deps.Renderer,
		menuInput:           deps.MenuInput,
		settingsOverlay:     deps.SettingsOverlay,
		settings:            deps.Settings,
		selector: deps.LevelSelectUseCases.NewSelector(
			deps.LastLevel,
		),
		levels:    deps.Levels,
		menuItems: menuItems,
	}
}

func (s *LevelSelectState) Update() types.StateTransition {
	// Настройки открываются из меню и возвращают в него
	if s.settingsOverlay.IsOpen() {
		s.settingsOverlay.Update()
		return types.StateTransition{}
	}
	if s.menuOpen {
		return s.handleMenu()
	}

	if s.menuInput.Back() {
		s.menuOpen = true
		s.menuIndex = 0
		return types.StateTransition{}
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

	switch {
	case s.settingsOverlay.IsOpen():
		s.settingsOverlay.Draw(screen)
	case s.menuOpen:
		s.renderer.DrawMenu(screen, types.LevelSelectMenuViewData{
			Items:       s.menuItems,
			ActiveIndex: s.menuIndex,
			Players:     s.settings.GetPlayers(),
		})
	}
}

// handleMenu — меню экрана: вверх-вниз — пункты, на PLAYERS
// влево-вправо или выбор переключают число игроков; назад
// закрывает меню
func (s *LevelSelectState) handleMenu() types.StateTransition {
	if s.menuInput.Back() {
		s.menuOpen = false
		return types.StateTransition{}
	}

	moveUp, moveDown := s.menuInput.Steps()
	s.menuIndex = stepIndex(s.menuIndex, len(s.menuItems), moveUp, moveDown)
	item := s.menuItems[s.menuIndex]
	confirmed := s.menuInput.Confirmed()

	switch item {
	case types.LevelSelectMenuItemPlayers:
		step := s.menuInput.SideStep()
		if confirmed {
			step = 1
		}
		if err := s.settingsUseCases.ChangePlayers(
			s.settings, step,
		); err != nil {
			log.Printf("save settings: %v", err)
		}
	case types.LevelSelectMenuItemBack:
		if confirmed {
			s.menuOpen = false
		}
	case types.LevelSelectMenuItemSettings:
		if confirmed {
			s.settingsOverlay.Open()
		}
	case types.LevelSelectMenuItemQuit:
		if confirmed {
			return types.StateTransition{Target: types.TransitionToQuit}
		}
	}

	return types.StateTransition{}
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
