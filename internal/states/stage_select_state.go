package states

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/opentype"

	stage_select "github.com/shpaker/gonflict/internal/adapters/stage_select"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"

	"github.com/shpaker/gonflict/internal/types/session_entities"
)

// StageSelectState представляет состояние выбора уровня
type StageSelectState struct {
	config             interfaces.IConfigProvider
	selector           *types.StageSelectorEntity
	selectorUseCases   *use_cases.StageSelectorUseCases
	transitionUseCases *use_cases.StateTransitionUseCases
	Session            *session_entities.GameSessionEntity
	inputAdapter       *stage_select.StageSelectKeyboardInputAdapter
	rendererAdapter    *stage_select.StageSelectRendererAdapter
	isSetUp            bool // Флаг для отслеживания, был ли вызван SetUp
	activeMenuItem     stageSelectMenuItem
	playerCount        int
	maxActiveEnemies   uint
}

type stageSelectMenuItem int

const (
	stageSelectMenuItemLevel stageSelectMenuItem = iota
	stageSelectMenuItemPlayers
	stageSelectMenuItemMaxEnemies
)

// NewStageSelectState создает новое состояние выбора уровня
func NewStageSelectState(
	config interfaces.IConfigProvider,
	selectorUseCases *use_cases.StageSelectorUseCases,
	transitionUseCases *use_cases.StateTransitionUseCases,
	session *session_entities.GameSessionEntity,
	fontUseCases interfaces.IFontUseCases,
	mapsRepository interfaces.IMapsDataRepository,
) (*StageSelectState, error) {
	// Получаем размеры экрана через type assertion к конкретному типу Config
	// (в интерфейсе нет этих методов, но они есть в реализации)
	type screenConfig interface {
		ScreenWidth() int
		ScreenHeight() int
	}
	var screenWidth, screenHeight int
	if screenCfg, ok := config.(screenConfig); ok {
		screenWidth = screenCfg.ScreenWidth()
		screenHeight = screenCfg.ScreenHeight()
	} else {
		// Fallback значения
		screenWidth = 800
		screenHeight = 600
	}

	// Получаем максимальное количество уровней из репозитория карт
	levelsCount, err := mapsRepository.GetLevelsCount()
	if err != nil {
		return nil, err
	}
	maxStages := uint(levelsCount)

	// Создаем селектор уровней
	selector := types.NewStageSelector(maxStages)

	baseFont, err := fontUseCases.GetFont()
	if err != nil {
		return nil, err
	}

	titleSize := config.GetTitleFontSize()
	if titleSize == 0 {
		titleSize = 32
	}
	regularSize := config.GetRegularFontSize()
	if regularSize == 0 {
		regularSize = titleSize
	}
	subtitleSize := config.GetSubtitleFontSize()
	if subtitleSize == 0 {
		subtitleSize = regularSize
	}

	fontFace, err := opentype.NewFace(baseFont, &opentype.FaceOptions{
		Size: float64(titleSize),
		DPI:  72,
	})
	if err != nil {
		return nil, err
	}

	textFace := text.NewGoXFace(fontFace)

	// Создаем адаптер рендеринга (вся работа с графикой, включая загрузку шрифта, в адаптере)
	rendererAdapter, err := stage_select.NewStageSelectRendererAdapter(
		selector,
		selectorUseCases,
		textFace,
		screenWidth,
		screenHeight,
		int(titleSize),
		int(regularSize),
		int(subtitleSize),
		config.GetGameTitle(),
	)
	if err != nil {
		return nil, err
	}

	inputAdapter := stage_select.NewStageSelectKeyboardInputAdapter(
		selector,
		selectorUseCases,
		ebiten.KeyLeft,
		ebiten.KeyRight,
		ebiten.KeyEnter,
	)

	state := &StageSelectState{
		config:             config,
		selector:           selector,
		selectorUseCases:   selectorUseCases,
		transitionUseCases: transitionUseCases,
		Session:            session,
		inputAdapter:       inputAdapter,
		rendererAdapter:    rendererAdapter,
		activeMenuItem:     stageSelectMenuItemLevel,
		playerCount:        1,
		maxActiveEnemies:   5,
	}

	return state, nil
}

// SetUp запускается один раз на старте состояния
func (s *StageSelectState) SetUp() {
	// Метод для инициализации состояния при первом вызове Update
}

// Update обновляет состояние выбора уровня
func (s *StageSelectState) Update() {
	// Вызываем SetUp один раз на старте состояния
	if !s.isSetUp {
		s.SetUp()
		s.isSetUp = true
	}

	s.handleMenuNavigation()

	// Обрабатываем ввод в зависимости от активного пункта меню
	if s.activeMenuItem == stageSelectMenuItemLevel {
		s.inputAdapter.Update(0)
	} else if s.activeMenuItem == stageSelectMenuItemPlayers {
		s.handlePlayerCountSelection()
	} else if s.activeMenuItem == stageSelectMenuItemMaxEnemies {
		s.handleMaxActiveEnemiesSelection()
	}

	// Обрабатываем переход к игре по Enter
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		selectedLevel := s.selectorUseCases.Select(s.selector)
		// Устанавливаем настройки в StageSessionEntity перед переходом
		if s.Session != nil && s.Session.StageSession() != nil {
			s.Session.StageSession().SetMaxActiveEnemies(s.maxActiveEnemies)
			s.Session.StageSession().SetPlayerCount(uint(s.playerCount))
		}
		s.transitionUseCases.ToGame(s.Session, selectedLevel)
	}
}

// Draw отрисовывает экран выбора уровня
func (s *StageSelectState) Draw(screen *ebiten.Image) {
	if s.rendererAdapter != nil {
		isLevelActive := s.activeMenuItem == stageSelectMenuItemLevel
		isPlayersActive := s.activeMenuItem == stageSelectMenuItemPlayers
		s.rendererAdapter.DrawAll(
			screen,
			isLevelActive,
			s.playerCount,
			isPlayersActive,
			s.maxActiveEnemies,
		)
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
		if s.maxActiveEnemies > 3 {
			s.maxActiveEnemies--
		} else {
			s.maxActiveEnemies = 10
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyRight) ||
		inpututil.IsKeyJustPressed(ebiten.KeyD) {
		if s.maxActiveEnemies < 10 {
			s.maxActiveEnemies++
		} else {
			s.maxActiveEnemies = 3
		}
	}
}
