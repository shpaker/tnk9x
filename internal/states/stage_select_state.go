package states

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	stage_select "github.com/shpaker/tnk9x/internal/adapters/stage_select"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

type StageSelectState struct {
	config           interfaces.IConfigProvider
	selector         *types.StageSelectorEntity
	selectorUseCases interfaces.IStageSelectorUseCases
	inputAdapter     *stage_select.StageSelectKeyboardInputAdapter
	rendererAdapter  *stage_select.StageSelectRendererAdapter
	touchControls    interfaces.ITouchControlsAdapter
	activeMenuItem   stageSelectMenuItem
	playerCount      int
	maxActiveEnemies uint
	// quitAvailable — строка QUIT есть только там, где приложение
	// может завершиться (десктоп); в js-сборке скрыта
	quitAvailable bool
}

type stageSelectMenuItem int

const (
	stageSelectMenuItemLevel stageSelectMenuItem = iota
	stageSelectMenuItemPlayers
	stageSelectMenuItemMaxEnemies
	stageSelectMenuItemQuit
)

func NewStageSelectState(
	config interfaces.IConfigProvider,
	selectorUseCases interfaces.IStageSelectorUseCases,
	textFace text.Face,
	mapsRepository interfaces.IMapsDataRepository,
	touchControls interfaces.ITouchControlsAdapter,
	quitAvailable bool,
) (*StageSelectState, error) {
	levelsCount, err := mapsRepository.GetLevelsCount()
	if err != nil {
		return nil, err
	}
	maxStages := uint(levelsCount)

	selector := types.NewStageSelector(maxStages)

	rendererAdapter := stage_select.NewStageSelectRendererAdapter(
		stage_select.StageSelectRendererDependencies{
			Selector:         selector,
			SelectorUseCases: selectorUseCases,
			FontFace:         textFace,
			TitleFontSize:    int(config.GetTitleFontSize()),
			RegularFontSize:  int(config.GetRegularFontSize()),
			SubtitleFontSize: int(config.GetSubtitleFontSize()),
			GameTitle:        config.GetGameTitle(),
		},
	)

	inputAdapter := stage_select.NewStageSelectKeyboardInputAdapter(
		selector,
		selectorUseCases,
		ebiten.KeyLeft,
		ebiten.KeyRight,
	)

	state := &StageSelectState{
		config:           config,
		selector:         selector,
		selectorUseCases: selectorUseCases,
		inputAdapter:     inputAdapter,
		rendererAdapter:  rendererAdapter,
		touchControls:    touchControls,
		activeMenuItem:   stageSelectMenuItemLevel,
		playerCount:      1,
		maxActiveEnemies: 5,
		quitAvailable:    quitAvailable,
	}

	return state, nil
}

func (s *StageSelectState) Update() types.StateTransition {
	s.handleMenuNavigation()

	if s.activeMenuItem == stageSelectMenuItemLevel {
		s.inputAdapter.Update(0)
	} else if s.activeMenuItem == stageSelectMenuItemPlayers {
		s.handlePlayerCountSelection()
	} else if s.activeMenuItem == stageSelectMenuItemMaxEnemies {
		s.handleMaxActiveEnemiesSelection()
	}

	transition := s.handleMenuTap()
	if transition.Target != types.TransitionNone {
		return transition
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if s.activeMenuItem == stageSelectMenuItemQuit {
			return types.StateTransition{
				Target: types.TransitionToQuit,
			}
		}

		return s.startTransition()
	}

	return types.StateTransition{}
}

func (s *StageSelectState) Draw(screen *ebiten.Image) {
	s.rendererAdapter.DrawAll(screen, types.StageSelectViewData{
		LevelActive:      s.activeMenuItem == stageSelectMenuItemLevel,
		PlayersActive:    s.activeMenuItem == stageSelectMenuItemPlayers,
		MaxEnemiesActive: s.activeMenuItem == stageSelectMenuItemMaxEnemies,
		QuitVisible:      s.quitAvailable,
		QuitActive:       s.activeMenuItem == stageSelectMenuItemQuit,
		PlayerCount:      uint(s.playerCount),
		MaxActiveEnemies: s.maxActiveEnemies,
		TouchActive:      s.touchControls.IsTouchActive(),
	})
}

// startTransition собирает переход к запуску выбранного уровня
func (s *StageSelectState) startTransition() types.StateTransition {
	selectedLevel := s.selectorUseCases.Select(s.selector)

	return types.StateTransition{
		Target:           types.TransitionToStage,
		Level:            selectedLevel,
		PlayerCount:      uint(s.playerCount),
		MaxActiveEnemies: s.maxActiveEnemies,
	}
}

// handleMenuTap — тап-управление меню: тап по строке фокусирует её,
// повторный тап меняет значение (левая половина экрана — назад,
// правая — вперёд), нижняя полоса запускает игру
func (s *StageSelectState) handleMenuTap() types.StateTransition {
	pos, ok := s.touchControls.TapJustPressed()
	if !ok {
		return types.StateTransition{}
	}

	hit := s.rendererAdapter.HitTest(pos)
	if hit == stage_select.MenuHitStart {
		return s.startTransition()
	}
	// QUIT — командная зона, тап активирует её сразу
	if hit == stage_select.MenuHitQuit {
		return types.StateTransition{Target: types.TransitionToQuit}
	}

	item, next, ok := menuHitTarget(hit)
	if !ok {
		return types.StateTransition{}
	}
	if s.activeMenuItem != item {
		s.activeMenuItem = item

		return types.StateTransition{}
	}
	s.applyMenuStep(item, next)

	return types.StateTransition{}
}

// menuHitTarget сопоставляет тап-зоне строку меню и направление шага
func menuHitTarget(
	hit stage_select.MenuHit,
) (stageSelectMenuItem, bool, bool) {
	switch hit {
	case stage_select.MenuHitLevelPrev:
		return stageSelectMenuItemLevel, false, true
	case stage_select.MenuHitLevelNext:
		return stageSelectMenuItemLevel, true, true
	case stage_select.MenuHitPlayersPrev:
		return stageSelectMenuItemPlayers, false, true
	case stage_select.MenuHitPlayersNext:
		return stageSelectMenuItemPlayers, true, true
	case stage_select.MenuHitMaxEnemiesPrev:
		return stageSelectMenuItemMaxEnemies, false, true
	case stage_select.MenuHitMaxEnemiesNext:
		return stageSelectMenuItemMaxEnemies, true, true
	}

	return 0, false, false
}

// applyMenuStep меняет значение активной строки меню на один шаг
func (s *StageSelectState) applyMenuStep(
	item stageSelectMenuItem,
	next bool,
) {
	switch item {
	case stageSelectMenuItemLevel:
		if next {
			s.selectorUseCases.Next(s.selector)
		} else {
			s.selectorUseCases.Previous(s.selector)
		}
	case stageSelectMenuItemPlayers:
		s.playerCount = 1
		if next {
			s.playerCount = 2
		}
	case stageSelectMenuItemMaxEnemies:
		if next {
			s.maxActiveEnemies = s.selectorUseCases.NextMaxActiveEnemies(
				s.maxActiveEnemies,
			)
		} else {
			s.maxActiveEnemies = s.selectorUseCases.PreviousMaxActiveEnemies(
				s.maxActiveEnemies,
			)
		}
	}
}

func (s *StageSelectState) handleMenuNavigation() {
	moveUp := inpututil.IsKeyJustPressed(ebiten.KeyUp) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW)
	moveDown := inpututil.IsKeyJustPressed(ebiten.KeyDown) ||
		inpututil.IsKeyJustPressed(ebiten.KeyS)

	if moveUp {
		switch s.activeMenuItem {
		case stageSelectMenuItemPlayers:
			s.activeMenuItem = stageSelectMenuItemLevel
		case stageSelectMenuItemMaxEnemies:
			s.activeMenuItem = stageSelectMenuItemPlayers
		case stageSelectMenuItemQuit:
			s.activeMenuItem = stageSelectMenuItemMaxEnemies
		}
	}
	if moveDown {
		switch s.activeMenuItem {
		case stageSelectMenuItemLevel:
			s.activeMenuItem = stageSelectMenuItemPlayers
		case stageSelectMenuItemPlayers:
			s.activeMenuItem = stageSelectMenuItemMaxEnemies
		case stageSelectMenuItemMaxEnemies:
			if s.quitAvailable {
				s.activeMenuItem = stageSelectMenuItemQuit
			}
		}
	}
}

func (s *StageSelectState) handlePlayerCountSelection() {
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) ||
		inpututil.IsKeyJustPressed(ebiten.KeyA) {
		s.playerCount = 1
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyRight) ||
		inpututil.IsKeyJustPressed(ebiten.KeyD) {
		s.playerCount = 2
	}
}

func (s *StageSelectState) handleMaxActiveEnemiesSelection() {
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) ||
		inpututil.IsKeyJustPressed(ebiten.KeyA) {
		s.maxActiveEnemies = s.selectorUseCases.PreviousMaxActiveEnemies(
			s.maxActiveEnemies,
		)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyRight) ||
		inpututil.IsKeyJustPressed(ebiten.KeyD) {
		s.maxActiveEnemies = s.selectorUseCases.NextMaxActiveEnemies(
			s.maxActiveEnemies,
		)
	}
}
