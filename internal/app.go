package internal

import (
	"context"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/repositories/raw"
	"github.com/shpaker/gonflict/internal/states"
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

	// Создаем игровую конфигурацию из общей конфигурации
	gameConfig := &states.GameConfig{
		EnemySpawners:  cfg.EnemySpawners,
		PlayerSpawners: cfg.PlayerSpawners,
	}

	// Создаем GameState с переданным реестром
	gameState, err := states.NewGameState(
		mapsRepo,
		scriptsRepo,
		tilesetRegistry,
		gameConfig,
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
