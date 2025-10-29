package internal

import (
	"context"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/shpaker/gonflict/internal/adapters"
	"github.com/shpaker/gonflict/internal/adapters/input_adapters"
	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/repositories/raw"
	"github.com/shpaker/gonflict/internal/services"
	"github.com/shpaker/gonflict/internal/states"
	"github.com/shpaker/gonflict/internal/use_cases"
)

type App struct {
	config *Config
	states.State
}

// ebiten game interface
func (app *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return app.config.ScreenWidth(), app.config.ScreenHeight()
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

	// Создаем репозиторий карт уровней
	mapsRepo := processed.NewMapsDataRepository(
		fileRepo,
		tilesetRegistry.Blocks(),
	)

	// Используем GameConfig напрямую из Config
	gameConfig := &cfg.GameConfig

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

	// Создаем сервисы коллизий
	boundaryCollisionService := services.NewBoundaryCollisionService(
		use_cases.MapWidthHeight,
		use_cases.TankSpriteSize,
	)
	wallCollisionService := services.NewWallCollisionService(
		use_cases.TankSpriteSize,
		use_cases.TileMinSize,
	)
	coordinateService := services.NewCoordinateService()
	tankBrakingService := services.NewTankBrakingService()

	// Создаем GameStateUseCasesFacade
	gameStateServices, err := states.NewGameStateUseCasesFacade(
		mapsRepo,
		scriptsRepo,
		gameConfig.LevelNumber,
		tilesetRegistry.Blocks(),
		tilesetRegistry.Player(),
		tilesetRegistry.Bullet(),
		tilesetRegistry.Spawner(),
		tilesetRegistry.Explosion(),
		gameConfig,
		gameRepo,
		boundaryCollisionService,
		wallCollisionService,
		coordinateService,
		tankBrakingService,
	)
	if err != nil {
		fmt.Printf("Error creating GameStateUseCasesFacade: %v\n", err)
		panic(err)
	}

	// Создаем RendererAdapter
	rendererAdapter := adapters.NewRendererAdapter(
		gameStateServices.MapUseCases(),
		gameStateServices.TankUseCases(),
		gameStateServices.BulletUseCases(),
		gameStateServices.GetEnemyTanks(),
		gameStateServices.GetEnemyUseCases(),
		mapTilesUseCases,
		playerTilesUseCases,
		bulletTilesUseCases,
		spawnerTilesUseCases,
		explosionTilesUseCases,
	)

	// Создаем InputAdapter
	inputAdapter := input_adapters.NewKeyboardInputAdapter(
		gameStateServices.TankUseCases(),
		ebiten.KeyW,     // up
		ebiten.KeyS,     // down
		ebiten.KeyA,     // left
		ebiten.KeyD,     // right
		ebiten.KeySpace, // shoot
	)

	// Создаем GameState с переданными зависимостями
	gameState, err := states.NewGameState(
		mapsRepo,
		scriptsRepo,
		tilesetRegistry,
		gameConfig,
		gameRepo,
		rendererAdapter,
		inputAdapter,
		gameStateServices,
	)
	if err != nil {
		// Логируем ошибку и падаем
		fmt.Printf("Error creating GameState: %v\n", err)
		panic(err)
	}

	return &App{
		config: cfg,
		State:  gameState,
	}
}

func (app *App) Run(ctx context.Context) error {
	ebiten.SetWindowSize(
		app.config.ScreenWidth()*3,
		app.config.ScreenHeight()*3,
	)
	ebiten.SetWindowTitle(app.config.Name)

	if err := ebiten.RunGame(app); err != nil {
		return err
	}

	return nil
}
