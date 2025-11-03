package internal

import (
	"context"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/shpaker/gonflict/internal/adapters"
	"github.com/shpaker/gonflict/internal/adapters/input_adapters"
	"github.com/shpaker/gonflict/internal/adapters/input_adapters/ai"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/repositories/raw"
	"github.com/shpaker/gonflict/internal/services"
	"github.com/shpaker/gonflict/internal/states"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

type App struct {
	config    *Config
	State     states.State
	luaEngine interfaces.ILuaEngine // Lua Engine для AI (существует весь срок жизни App)
}

// ebiten game interface
func (app *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return app.config.ScreenWidth() / 3, app.config.ScreenHeight() / 3
}

func (app *App) Update() error {
	newState, err := app.State.Update()
	if err != nil {
		return err
	}
	if newState != nil {
		app.State = newState
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
	fileRepo := raw.NewFileRepository("assets")

	// Создаем реестр всех тайлсетов
	tilesetRegistry, err := processed.NewTilesetRepositoryRegistry(fileRepo)
	if err != nil {
		fmt.Printf("Error creating TilesetRepositoryRegistry: %v\n", err)
		panic(err)
	}

	// Создаем репозиторий скриптов
	scriptsRepo := processed.NewScriptsRepository(fileRepo)

	// Создаем Lua engine для AI (общий для всех GameState, существует весь срок жизни App)
	luaEngine := ai.NewLuaEngine()

	// Создаем репозиторий карт уровней
	mapsRepo := processed.NewMapsDataRepository(
		fileRepo,
		tilesetRegistry.Blocks(),
		cfg.MapBlocksCount.Width,
	)

	// Создаем реестр игровых репозиториев
	gameRepo := game.NewGameRepositoriesRegistry()

	// Создаем временный GameStateUseCasesFacade для получения Use Cases (будет пересоздан позже)
	// Но для создания RendererAdapter нам нужны Use Cases, поэтому создадим их через фасад
	// Вместо этого создадим TilesUseCases напрямую для RendererAdapter
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

	// Создаем сервисы коллизий
	mapWidthHeight := cfg.MapBlocksCount.Width * int(cfg.TileBaseSize)
	boundaryCollisionService := services.NewBoundaryCollisionService(
		mapWidthHeight,
		int(cfg.BaseSizePx),
	)
	wallCollisionService := services.NewWallCollisionService(
		int(cfg.BaseSizePx),
		int(cfg.TileBaseSize),
	)
	coordinateService := services.NewCoordinateService()
	tankBrakingService := services.NewTankBrakingService()

	// Сначала создаем временный RendererAdapter и InputAdapter
	// (они нужны для создания GameState, но сам GameState создаст их правильно)
	mapOffsetX := int(cfg.MapOffsets[0])
	mapOffsetY := int(cfg.MapOffsets[1])
	tempRendererAdapter := adapters.NewGameStateRendererAdapter(
		nil, // mapUseCases
		nil, // playerTank
		nil, // tankRenderUseCases
		nil, // bulletUseCases
		nil, // enemyTanks
		mapTilesUseCases,
		playerTilesUseCases,
		bulletTilesUseCases,
		spawnerTilesUseCases,
		explosionTilesUseCases,
		hqTilesUseCases,
		nil, // hq
		nil, // hqUseCases
		int(cfg.TileBaseSize),
		mapOffsetX,
		mapOffsetY,
		mapWidthHeight,
	)
	tempInputAdapter := input_adapters.NewKeyboardInputAdapter(
		nil, nil,
		ebiten.KeyW,     // up
		ebiten.KeyS,     // down
		ebiten.KeyA,     // left
		ebiten.KeyD,     // right
		ebiten.KeySpace, // shoot
	)

	// Создаем GameState
	gameStatePtr, err := states.NewGameState(
		mapsRepo,
		scriptsRepo,
		cfg.LevelNumber,
		tilesetRegistry,
		cfg,
		gameRepo,
		boundaryCollisionService,
		wallCollisionService,
		coordinateService,
		tankBrakingService,
		tempRendererAdapter,
		tempInputAdapter,
		luaEngine,
	)
	if err != nil {
		fmt.Printf("Error creating GameState: %v\n", err)
		panic(err)
	}

	// Извлекаем срезы из массивов для адаптеров
	enemyTanksSlice := make([]*types.TankEntity, 0, 3)
	for _, tank := range gameStatePtr.EnemyTanks {
		if tank != nil {
			enemyTanksSlice = append(enemyTanksSlice, tank)
		}
	}

	// Теперь создаем правильный RendererAdapter с реальными данными
	rendererAdapter := adapters.NewGameStateRendererAdapter(
		gameStatePtr.MapUseCases,
		gameStatePtr.PlayerTank,
		gameStatePtr.TankRenderUseCases,
		gameStatePtr.BulletUseCases,
		enemyTanksSlice,
		mapTilesUseCases,
		playerTilesUseCases,
		bulletTilesUseCases,
		spawnerTilesUseCases,
		explosionTilesUseCases,
		hqTilesUseCases,
		gameStatePtr.HQEntity,
		gameStatePtr.HQUseCases,
		int(cfg.TileBaseSize),
		mapOffsetX,
		mapOffsetY,
		mapWidthHeight,
	)

	// Создаем InputAdapter используя TankActionsUseCases из gameState
	inputAdapter := input_adapters.NewKeyboardInputAdapter(
		gameStatePtr.TankActionsUseCases,
		gameStatePtr.PlayerTank,
		ebiten.KeyW,     // up
		ebiten.KeyS,     // down
		ebiten.KeyA,     // left
		ebiten.KeyD,     // right
		ebiten.KeySpace, // shoot
	)

	// Обновляем адаптеры в gameState
	gameStatePtr.InputAdapter = inputAdapter
	gameStatePtr.RendererAdapter = rendererAdapter

	return &App{
		config:    cfg,
		State:     gameStatePtr,
		luaEngine: luaEngine,
	}
}

func (app *App) Run(ctx context.Context) error {
	ebiten.SetWindowSize(
		app.config.ScreenWidth(),
		app.config.ScreenHeight(),
	)
	ebiten.SetWindowTitle(app.config.Name)

	err := ebiten.RunGame(app)

	// Закрываем luaEngine при завершении App
	app.Close()

	return err
}

// Close освобождает ресурсы App, включая Lua engine
func (app *App) Close() {
	if app.luaEngine != nil {
		app.luaEngine.Close()
		app.luaEngine = nil
	}
}
