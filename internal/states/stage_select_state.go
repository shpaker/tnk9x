package states

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	stage_select "github.com/shpaker/gonflict/internal/adapters/stage_select"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// StageSelectState представляет состояние выбора уровня
type StageSelectState struct {
	config             interfaces.IConfigProvider
	selector           *types.StageSelectorEntity
	selectorUseCases   *use_cases.StageSelectorUseCases
	transitionUseCases *use_cases.StateTransitionUseCases
	session            *types.SessionEntity
	inputAdapter       *stage_select.StageSelectKeyboardInputAdapter
	rendererAdapter    *stage_select.StageSelectRendererAdapter
}

// NewStageSelectState создает новое состояние выбора уровня
func NewStageSelectState(
	config interfaces.IConfigProvider,
	selectorUseCases *use_cases.StageSelectorUseCases,
	transitionUseCases *use_cases.StateTransitionUseCases,
	session *types.SessionEntity,
	fontsRepository interfaces.IFontsRepository,
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

	// Создаем адаптер рендеринга (вся работа с графикой, включая загрузку шрифта, в адаптере)
	rendererAdapter, err := stage_select.NewStageSelectRendererAdapter(
		selector,
		selectorUseCases,
		fontsRepository,
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
		session:            session,
		inputAdapter:       inputAdapter,
		rendererAdapter:    rendererAdapter,
	}

	return state, nil
}

// Update обновляет состояние выбора уровня
func (s *StageSelectState) Update() {
	// Обрабатываем ввод
	s.inputAdapter.Update(0)

	// Обрабатываем переход к игре по Enter
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		selectedLevel := s.selectorUseCases.Select(s.selector)
		s.transitionUseCases.ToGame(s.session, selectedLevel)
	}
}

// Draw отрисовывает экран выбора уровня
func (s *StageSelectState) Draw(screen *ebiten.Image) {
	if s.rendererAdapter != nil {
		s.rendererAdapter.DrawAll(screen)
	}
}
