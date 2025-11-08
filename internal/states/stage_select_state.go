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
}

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

	fontFace, err := opentype.NewFace(baseFont, &opentype.FaceOptions{
		Size: 32,
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
	)
	if err != nil {
		return nil, err
	}

	inputAdapter := stage_select.NewStageSelectKeyboardInputAdapter(
		selector,
		selectorUseCases,
		ebiten.KeyW,     // up
		ebiten.KeyS,     // down
		ebiten.KeyEnter, // enter
	)

	state := &StageSelectState{
		config:             config,
		selector:           selector,
		selectorUseCases:   selectorUseCases,
		transitionUseCases: transitionUseCases,
		Session:            session,
		inputAdapter:       inputAdapter,
		rendererAdapter:    rendererAdapter,
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

	// Обрабатываем ввод
	s.inputAdapter.Update(0)

	// Обрабатываем переход к игре по Enter
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		selectedLevel := s.selectorUseCases.Select(s.selector)
		s.transitionUseCases.ToGame(s.Session, selectedLevel)
	}
}

// Draw отрисовывает экран выбора уровня
func (s *StageSelectState) Draw(screen *ebiten.Image) {
	if s.rendererAdapter != nil {
		s.rendererAdapter.DrawAll(screen)
	}
}
