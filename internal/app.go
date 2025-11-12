package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/gonflict/internal/adapters/stage/input_adapters/ai"
	"github.com/shpaker/gonflict/internal/interfaces"
	game_repos "github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/repositories/raw"
	"github.com/shpaker/gonflict/internal/services"
	collision_services "github.com/shpaker/gonflict/internal/services/collision_services"
	"github.com/shpaker/gonflict/internal/states"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"

	"github.com/shpaker/gonflict/internal/types/session_entities"
)

type App struct {
	config        *Config
	State         interfaces.IState
	luaEngine     interfaces.ILuaEngine               // Lua Engine для AI (существует весь срок жизни App)
	session       *session_entities.GameSessionEntity // Сессия игры
	fontUseCases  interfaces.IFontUseCases
	debugUseCases *use_cases.DebugUseCases
}

// ebiten game interface
func (app *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return app.config.ScreenWidth() / 3, app.config.ScreenHeight() / 3
}

func (app *App) Update() error {
	// Обработка ESC для выхода из приложения
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("exit application")
	}

	// Переключаем режим отладки по F1
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		Debug = !Debug
	}

	// Обновляем текущее состояние
	app.State.Update()

	// Проверяем переходы через session.GetTargetState() (автоматически обнуляется)
	targetState := app.session.GetTargetState()
	if targetState != nil {
		newState, err := app.createStateFromTarget(app.session, targetState)
		if err != nil {
			return err
		}
		if newState != nil {
			app.State = newState
		}
	}

	return nil
}

func (app *App) Draw(screen *ebiten.Image) {
	app.State.Draw(screen)
	if app.debugUseCases == nil {
		return
	}

	if !Debug {
		return
	}

	debugText := app.debugUseCases.BuildDebugInfo()
	if debugText == "" {
		return
	}

	ebitenutil.DebugPrintAt(
		screen,
		debugText,
		0,
		0,
	)
}

func New(cfg *Config) *App {
	// Создаем файловый репозиторий
	fileRepository := raw.NewFileRepository("assets")

	// Создаем реестр тайлсетов
	tilesetRegistry, err := processed.NewTilesetRepositoryRegistry(
		fileRepository,
	)
	if err != nil {
		fmt.Printf("Error creating tileset registry: %v\n", err)
		panic(err)
	}

	// Создаем репозиторий карт
	mapsRepository := processed.NewMapsDataRepository(
		fileRepository,
		tilesetRegistry,
	)
	fontsRepository := processed.NewFontsRepository(fileRepository)
	fontUseCases := use_cases.NewFontUseCases(fontsRepository)

	// Создаем Lua engine для AI (общий для всех StageState, существует весь срок жизни App)
	luaEngine := ai.NewLuaEngine()

	// Создаем GameSessionEntity
	session := session_entities.NewGameSessionEntity()

	// Создаем Use Cases для выбор уровня
	stageSelectorUseCases := use_cases.NewStageSelectorUseCases()
	stateTransitionUseCases := use_cases.NewStateTransitionUseCases()

	// Создаем StageSelectState как дефолтное состояние
	stageSelectState, err := states.NewStageSelectState(
		cfg,
		stageSelectorUseCases,
		stateTransitionUseCases,
		session,
		fontUseCases,
		mapsRepository,
	)
	if err != nil {
		fmt.Printf("Error creating StageSelectState: %v\n", err)
		panic(err)
	}

	return &App{
		config:       cfg,
		State:        stageSelectState,
		luaEngine:    luaEngine,
		session:      session,
		fontUseCases: fontUseCases,
		debugUseCases: use_cases.NewDebugUseCases(
			session,
			Version,
		),
	}
}

// createStateFromTarget создает состояние на основе TargetState
func (app *App) createStateFromTarget(
	session *session_entities.GameSessionEntity,
	targetState *types.StateType,
) (interfaces.IState, error) {
	if targetState == nil {
		return nil, nil // Нет перехода
	}

	switch *targetState {
	case types.StateTypeStage:
		return app.createStageState(session)
	case types.StateTypeStageSelect:
		return app.createStageSelectState(session)
	default:
		return nil, nil
	}
}

// createStageState создает игровое состояние
func (app *App) createStageState(
	session *session_entities.GameSessionEntity,
) (interfaces.IState, error) {
	// Создаем все необходимые зависимости для StageState
	fileRepository := raw.NewFileRepository("assets")

	// Создаем реестр тайлсетов
	tilesetRegistry, err := processed.NewTilesetRepositoryRegistry(
		fileRepository,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tileset registry: %w", err)
	}

	// Создаем репозиторий скриптов
	scriptsRepository := processed.NewScriptsRepository(fileRepository)

	// Создаем репозиторий карт
	mapsRepository := processed.NewMapsDataRepository(
		fileRepository,
		tilesetRegistry,
	)
	// Создаем реестр игровых репозиториев
	gameRepository := game_repos.NewGameRepositoriesRegistry()

	// Создаем сервисы
	mapWidthHeight := app.config.MapBlocksCount.Width * int(
		app.config.TileBaseSize,
	)
	boundaryCollisionService := collision_services.NewBoundaryCollisionService(
		mapWidthHeight,
		int(app.config.BaseSizePx),
	)
	entitiesCollisionService := collision_services.NewEntitiesCollisionService()
	coordinateService := services.NewCoordinateService()
	wallCollisionService := collision_services.NewWallCollisionService(
		int(app.config.BaseSizePx),
		int(app.config.TileBaseSize),
		coordinateService,
		entitiesCollisionService,
	)
	tankBrakingService := services.NewTankBrakingService()

	fontUseCases := app.fontUseCases
	if fontUseCases == nil {
		fontUseCases = use_cases.NewFontUseCases(
			processed.NewFontsRepository(fileRepository),
		)
	}

	// Создаем StageState
	stageStatePtr, err := states.NewStageState(
		mapsRepository,
		scriptsRepository,
		session.Level,
		tilesetRegistry,
		app.config,
		fontUseCases,
		gameRepository,
		boundaryCollisionService,
		wallCollisionService,
		coordinateService,
		tankBrakingService,
		app.luaEngine,
		session,
	)
	if err != nil {
		return nil, err
	}

	return stageStatePtr, nil
}

// createStageSelectState создает состояние выбора уровня
func (app *App) createStageSelectState(
	session *session_entities.GameSessionEntity,
) (interfaces.IState, error) {
	fileRepository := raw.NewFileRepository("assets")

	// Создаем реестр тайлсетов
	tilesetRegistry, err := processed.NewTilesetRepositoryRegistry(
		fileRepository,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tileset registry: %w", err)
	}

	// Создаем репозиторий карт
	mapsRepository := processed.NewMapsDataRepository(
		fileRepository,
		tilesetRegistry,
	)

	fontUseCases := app.fontUseCases
	if fontUseCases == nil {
		fontUseCases = use_cases.NewFontUseCases(
			processed.NewFontsRepository(fileRepository),
		)
		app.fontUseCases = fontUseCases
	}

	stageSelectorUseCases := use_cases.NewStageSelectorUseCases()
	stateTransitionUseCases := use_cases.NewStateTransitionUseCases()

	return states.NewStageSelectState(
		app.config,
		stageSelectorUseCases,
		stateTransitionUseCases,
		session,
		fontUseCases,
		mapsRepository,
	)
}

func (app *App) Run(ctx context.Context) error {
	ebiten.SetWindowSize(
		app.config.ScreenWidth(),
		app.config.ScreenHeight(),
	)
	ebiten.SetWindowTitle(fmt.Sprintf("%s v%s", app.config.Name, Version))

	err := ebiten.RunGame(app)

	// Закрываем luaEngine при завершении App
	defer app.Close()

	return err
}

// Close освобождает ресурсы App, включая Lua engine
func (app *App) Close() {
	if app.luaEngine != nil {
		app.luaEngine.Close()
		app.luaEngine = nil
	}
}
