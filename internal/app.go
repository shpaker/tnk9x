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
	new_state, err := app.State.Update()
	if err != nil {
		return err
	}
	if new_state != nil {
		app.State = new_state
	}
	return nil
}

func (app *App) Draw(screen *ebiten.Image) {
	app.State.Draw(screen)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %.2f,\nFPS: %.2f", ebiten.ActualTPS(), ebiten.ActualFPS()), 0, 0)
}

func New(cfg *Config) *App {
	// Создаем файловый репозиторий
	fileRepo := raw.NewFileRepository("assets")

	// Создаем tilesetRepository для работы с изображениями блоков
	tilesetRepo, err := processed.NewTilesetDataRepository(fileRepo, "tiles/blocks")
	if err != nil {
		fmt.Printf("Ошибка создания TilesetRepository: %v\n", err)
		panic(err)
	}

	// Создаем tilesetRepository для работы с изображениями игрока
	playerTilesetRepo, err := processed.NewTilesetDataRepository(fileRepo, "tiles/player")
	if err != nil {
		fmt.Printf("Ошибка создания PlayerTilesetRepository: %v\n", err)
		panic(err)
	}

	// Создаем tilesetRepository для работы с изображениями пуль
	bulletTilesetRepo, err := processed.NewTilesetDataRepository(fileRepo, "tiles/bullet")
	if err != nil {
		fmt.Printf("Ошибка создания BulletTilesetRepository: %v\n", err)
		panic(err)
	}

	// Создаем tilesetRepository для работы с анимацией спавна
	spawnerTilesetRepo, err := processed.NewTilesetDataRepository(fileRepo, "tiles/spawner")
	if err != nil {
		fmt.Printf("Ошибка создания SpawnerTilesetRepository: %v\n", err)
		panic(err)
	}

	// Создаем репозиторий карт уровней
	mapsRepo := processed.NewMapsDataRepository(fileRepo, tilesetRepo)

	// Создаем игровую конфигурацию из общей конфигурации
	gameConfig := &states.GameConfig{
		SpawnDurationMs: cfg.SpawnDurationMs,
		EnemySpawners:   cfg.EnemySpawners,
	}

	// Создаем GameState с переданными репозиториями
	gameState, err := states.NewGameState(
		mapsRepo,
		tilesetRepo,        // Используем tilesetRepo для блоков карты
		playerTilesetRepo,  // Используем playerTilesetRepo для игрока
		bulletTilesetRepo,  // Используем bulletTilesetRepo для пуль
		spawnerTilesetRepo, // Используем spawnerTilesetRepo для спавна
		gameConfig,         // Передаем игровую конфигурацию
	)
	if err != nil {
		// Логируем ошибку и падаем
		fmt.Printf("Ошибка создания GameState: %v\n", err)
		panic(err)
	}

	return &App{
		config: cfg,
		State:  gameState,
	}
}

func (app *App) Run(ctx context.Context) error {
	ebiten.SetWindowSize(app.config.ScreenWidth()*3, app.config.ScreenHeight()*3)
	ebiten.SetWindowTitle(app.config.Name)

	if err := ebiten.RunGame(app); err != nil {
		return err
	}

	return nil
}
