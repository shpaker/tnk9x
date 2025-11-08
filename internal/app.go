package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	game "github.com/shpaker/gonflict/internal/adapters/game"
	"github.com/shpaker/gonflict/internal/adapters/game/input_adapters"
	"github.com/shpaker/gonflict/internal/adapters/game/input_adapters/ai"
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
	config    *Config
	State     interfaces.IState
	luaEngine interfaces.ILuaEngine               // Lua Engine для AI (существует весь срок жизни App)
	session   *session_entities.GameSessionEntity // Сессия игры
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
	ebitenutil.DebugPrintAt(
		screen,
		fmt.Sprintf(
			"TPS: %.2f,\nFPS: %.2f",
			ebiten.ActualTPS(),
			ebiten.ActualFPS(),
		),
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
		tilesetRegistry.Blocks(),
	)

	// Создаем репозиторий шрифтов
	fontsRepository := processed.NewFontsRepository(fileRepository)

	// Создаем Lua engine для AI (общий для всех GameState, существует весь срок жизни App)
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
		fontsRepository,
		mapsRepository,
	)
	if err != nil {
		fmt.Printf("Error creating StageSelectState: %v\n", err)
		panic(err)
	}

	return &App{
		config:    cfg,
		State:     stageSelectState,
		luaEngine: luaEngine,
		session:   session,
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
	case types.StateTypeGame:
		return app.createGameState(session)
	case types.StateTypeStageSelect:
		return app.createStageSelectState(session)
	default:
		return nil, nil
	}
}

// createGameState создает игровое состояние
func (app *App) createGameState(
	session *session_entities.GameSessionEntity,
) (interfaces.IState, error) {
	// Создаем все необходимые зависимости для GameState
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
		tilesetRegistry.Blocks(),
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

	// Создаем временные адаптеры
	mapOffsetX := int(app.config.MapOffsets[0])
	mapOffsetY := int(app.config.MapOffsets[1])
	mapWidthHeightForAdapter := app.config.MapBlocksCount.Width * int(
		app.config.TileBaseSize,
	)

	// Создаем временные TilesUseCases для адаптеров
	mapTileService := services.NewTileService(tilesetRegistry.Blocks())
	playerTileService := services.NewTileService(tilesetRegistry.Player())
	bulletTileService := services.NewTileService(tilesetRegistry.Bullet())
	spawnerTileService := services.NewTileService(tilesetRegistry.Spawner())
	explosionTileService := services.NewTileService(tilesetRegistry.Explosion())
	hqTileService := services.NewTileService(tilesetRegistry.HQ())
	animationService := services.NewAnimationService()

	mapTilesUseCases := use_cases.NewTilesUseCases(
		tilesetRegistry.Blocks(),
		mapTileService,
		animationService,
	)
	playerTilesUseCases := use_cases.NewTilesUseCases(
		tilesetRegistry.Player(),
		playerTileService,
		animationService,
	)
	bulletTilesUseCases := use_cases.NewTilesUseCases(
		tilesetRegistry.Bullet(),
		bulletTileService,
		animationService,
	)
	spawnerTilesUseCases := use_cases.NewTilesUseCases(
		tilesetRegistry.Spawner(),
		spawnerTileService,
		animationService,
	)
	explosionTilesUseCases := use_cases.NewTilesUseCases(
		tilesetRegistry.Explosion(),
		explosionTileService,
		animationService,
	)
	hqTilesUseCases := use_cases.NewTilesUseCases(
		tilesetRegistry.HQ(),
		hqTileService,
		animationService,
	)

	tempRendererAdapter := game.NewGameRendererAdapter(
		nil, nil, nil, nil,
		mapTilesUseCases,
		playerTilesUseCases,
		bulletTilesUseCases,
		spawnerTilesUseCases,
		explosionTilesUseCases,
		hqTilesUseCases,
		nil,
		int(app.config.TileBaseSize),
		mapOffsetX,
		mapOffsetY,
		mapWidthHeightForAdapter,
	)

	tempInputAdapter := input_adapters.NewGameKeyboardInputAdapter(
		nil, nil,
		ebiten.KeyW,
		ebiten.KeyS,
		ebiten.KeyA,
		ebiten.KeyD,
		ebiten.KeySpace,
	)

	// Создаем GameState
	gameStatePtr, err := states.NewGameState(
		mapsRepository,
		scriptsRepository,
		session.Level,
		tilesetRegistry,
		app.config,
		gameRepository,
		boundaryCollisionService,
		wallCollisionService,
		coordinateService,
		tankBrakingService,
		tempRendererAdapter,
		tempInputAdapter,
		app.luaEngine,
		session,
	)
	if err != nil {
		return nil, err
	}

	// Обновляем адаптеры с реальными данными
	rendererAdapter := game.NewGameRendererAdapter(
		gameStatePtr.MapUseCases,
		gameStatePtr.TankCommonUseCases,
		gameStatePtr.TankRenderUseCases,
		gameStatePtr.BulletUseCases,
		mapTilesUseCases,
		playerTilesUseCases,
		bulletTilesUseCases,
		spawnerTilesUseCases,
		explosionTilesUseCases,
		hqTilesUseCases,
		gameStatePtr.HQUseCases,
		int(app.config.TileBaseSize),
		mapOffsetX,
		mapOffsetY,
		mapWidthHeightForAdapter,
	)

	inputAdapter := input_adapters.NewGameKeyboardInputAdapter(
		gameStatePtr.TankActionsUseCases,
		nil,
		ebiten.KeyW,
		ebiten.KeyS,
		ebiten.KeyA,
		ebiten.KeyD,
		ebiten.KeySpace,
	)

	gameStatePtr.InputAdapter = inputAdapter
	gameStatePtr.RendererAdapter = rendererAdapter

	return gameStatePtr, nil
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
		tilesetRegistry.Blocks(),
	)

	fontsRepository := processed.NewFontsRepository(fileRepository)

	stageSelectorUseCases := use_cases.NewStageSelectorUseCases()
	stateTransitionUseCases := use_cases.NewStateTransitionUseCases()

	return states.NewStageSelectState(
		app.config,
		stageSelectorUseCases,
		stateTransitionUseCases,
		session,
		fontsRepository,
		mapsRepository,
	)
}

func (app *App) Run(ctx context.Context) error {
	ebiten.SetWindowSize(
		app.config.ScreenWidth(),
		app.config.ScreenHeight(),
	)
	ebiten.SetWindowTitle(app.config.Name)

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
