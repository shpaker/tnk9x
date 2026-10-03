package states

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// mainMenuFadeInTicks — появление меню из чёрного
const mainMenuFadeInTicks = 24

// MainMenuRenderer — контракт рендера главного меню, определён
// у потребителя
type MainMenuRenderer interface {
	Draw(screen *ebiten.Image, view types.MainMenuViewData)
	// HitRow — пункт последней отрисовки под точкой
	HitRow(position types.Position) (int, bool)
}

// MenuScene — живая сцена за главным меню; контракт определён
// у потребителя
type MenuScene interface {
	Update()
	Draw(screen *ebiten.Image)
	// StopSounds глушит звуки сцены при уходе из меню
	StopSounds()
}

// MainMenuStateDependencies — готовый граф зависимостей главного
// меню; собирается composition root'ом, все поля обязательны
type MainMenuStateDependencies struct {
	// Use Cases
	SettingsUseCases interfaces.ISettingsUseCases

	// Adapters
	Renderer  MainMenuRenderer
	MenuInput interfaces.IMenuInputAdapter

	// Presentation
	// Scene — сцена за меню: название из блоков и танки с ИИ
	Scene           MenuScene
	SettingsOverlay *SettingsOverlay
	ShopOverlay     *ShopOverlay

	// Entities
	Settings *types.SettingsEntity

	// QuitAvailable — выход из игры есть только там, где приложение
	// может завершиться (десктоп)
	QuitAvailable bool
	// ShopAvailable — магазин есть только там, где площадка умеет
	// покупки и продаёт хоть что-то
	ShopAvailable bool
}

// MainMenuState — главное меню: выбор режима на одного или двоих
// (у каждого своё прохождение), магазин, настройки и выход;
// появляется из чёрного
type MainMenuState struct {
	// Use Cases
	settingsUseCases interfaces.ISettingsUseCases
	// Adapters
	renderer  MainMenuRenderer
	menuInput interfaces.IMenuInputAdapter
	// Presentation
	scene           MenuScene
	settingsOverlay *SettingsOverlay
	shopOverlay     *ShopOverlay
	// Entities
	settings *types.SettingsEntity

	items       []types.MainMenuItem
	activeIndex int
	ticks       uint
}

func NewMainMenuState(deps MainMenuStateDependencies) *MainMenuState {
	items := []types.MainMenuItem{
		types.MainMenuItemOnePlayer,
		types.MainMenuItemTwoPlayers,
	}
	if deps.ShopAvailable {
		items = append(items, types.MainMenuItemShop)
	}
	items = append(items, types.MainMenuItemSettings)
	if deps.QuitAvailable {
		items = append(items, types.MainMenuItemQuit)
	}

	// Курсор встаёт на последний выбранный режим
	activeIndex := 0
	if deps.Settings.GetPlayers() > 1 {
		activeIndex = 1
	}

	return &MainMenuState{
		settingsUseCases: deps.SettingsUseCases,
		renderer:         deps.Renderer,
		menuInput:        deps.MenuInput,
		scene:            deps.Scene,
		settingsOverlay:  deps.SettingsOverlay,
		shopOverlay:      deps.ShopOverlay,
		settings:         deps.Settings,
		items:            items,
		activeIndex:      activeIndex,
	}
}

func (s *MainMenuState) Update() types.StateTransition {
	// Сцена живёт и под открытыми настройками и магазином
	s.scene.Update()
	s.ticks++
	if s.settingsOverlay.IsOpen() {
		s.settingsOverlay.Update()
		return types.StateTransition{}
	}
	if s.shopOverlay.IsOpen() {
		s.shopOverlay.Update()
		return types.StateTransition{}
	}

	moveUp, moveDown := s.menuInput.Steps()
	s.activeIndex = stepIndex(s.activeIndex, len(s.items), moveUp, moveDown)
	activeIndex, clicked := pointerIndex(
		s.menuInput, s.renderer.HitRow, s.activeIndex,
	)
	s.activeIndex = activeIndex
	if !s.menuInput.Confirmed() && !clicked {
		return types.StateTransition{}
	}

	switch s.items[s.activeIndex] {
	case types.MainMenuItemOnePlayer:
		return s.play(1)
	case types.MainMenuItemTwoPlayers:
		return s.play(2)
	case types.MainMenuItemShop:
		s.shopOverlay.Open()
	case types.MainMenuItemSettings:
		s.settingsOverlay.Open()
	case types.MainMenuItemQuit:
		s.scene.StopSounds()
		return types.StateTransition{Target: types.TransitionToQuit}
	}
	return types.StateTransition{}
}

func (s *MainMenuState) Draw(screen *ebiten.Image) {
	fade := 0.0
	if s.ticks < mainMenuFadeInTicks {
		fade = 1 - float64(s.ticks)/mainMenuFadeInTicks
	}
	s.scene.Draw(screen)
	s.renderer.Draw(screen, types.MainMenuViewData{
		Items:       s.items,
		ActiveIndex: s.activeIndex,
		Fade:        fade,
	})
	if s.settingsOverlay.IsOpen() {
		s.settingsOverlay.Draw(screen)
	}
	if s.shopOverlay.IsOpen() {
		s.shopOverlay.Draw(screen)
	}
}

// play выбирает режим и ведёт к выбору уровня его кампании
func (s *MainMenuState) play(players uint) types.StateTransition {
	if err := s.settingsUseCases.SetPlayers(s.settings, players); err != nil {
		log.Printf("save settings: %v", err)
	}
	s.scene.StopSounds()
	return types.StateTransition{Target: types.TransitionToLevelSelect}
}
