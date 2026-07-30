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
	activeMenuItem   stageSelectMenuItem
	playerCount      int
	maxActiveEnemies uint
}

type stageSelectMenuItem int

const (
	stageSelectMenuItemLevel stageSelectMenuItem = iota
	stageSelectMenuItemPlayers
	stageSelectMenuItemMaxEnemies
)

func NewStageSelectState(
	config interfaces.IConfigProvider,
	selectorUseCases interfaces.IStageSelectorUseCases,
	textFace text.Face,
	mapsRepository interfaces.IMapsDataRepository,
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
		activeMenuItem:   stageSelectMenuItemLevel,
		playerCount:      1,
		maxActiveEnemies: 5,
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

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		selectedLevel := s.selectorUseCases.Select(s.selector)

		return types.StateTransition{
			Target:           types.TransitionToStage,
			Level:            selectedLevel,
			PlayerCount:      uint(s.playerCount),
			MaxActiveEnemies: s.maxActiveEnemies,
		}
	}

	return types.StateTransition{}
}

func (s *StageSelectState) Draw(screen *ebiten.Image) {
	s.rendererAdapter.DrawAll(screen, types.StageSelectViewData{
		LevelActive:      s.activeMenuItem == stageSelectMenuItemLevel,
		PlayersActive:    s.activeMenuItem == stageSelectMenuItemPlayers,
		MaxEnemiesActive: s.activeMenuItem == stageSelectMenuItemMaxEnemies,
		PlayerCount:      uint(s.playerCount),
		MaxActiveEnemies: s.maxActiveEnemies,
	})
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
		}
	}
	if moveDown {
		switch s.activeMenuItem {
		case stageSelectMenuItemLevel:
			s.activeMenuItem = stageSelectMenuItemPlayers
		case stageSelectMenuItemPlayers:
			s.activeMenuItem = stageSelectMenuItemMaxEnemies
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
