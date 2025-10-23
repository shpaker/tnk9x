package app

import (
	"context"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/shpaker/gonflict/internal/config"
	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/repositories/raw"
	services "github.com/shpaker/gonflict/internal/services"
	"github.com/shpaker/gonflict/internal/states"
)

type App struct {
	config  *config.Config
	service *services.WindowService
	interfaces.State
}

// ebiten game interface
func (app *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return constants.ScreenWidth, constants.ScreenHeight
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

func New(cfg *config.Config) *App {
	svc := services.NewWindowService()

	// Создаем файловый репозиторий
	fileRepo := raw.NewFileRepository("assets")

	// Создаем репозиторий спрайтов
	spritesRepo, err := processed.NewSpritesRepository(fileRepo)
	if err != nil {
		// Логируем ошибку и падаем
		fmt.Printf("Ошибка создания SpritesRepository: %v\n", err)
		panic(err)
	}

	// Создаем репозиторий карт уровней
	mapsRepo := processed.NewMapsDataRepository(fileRepo, spritesRepo)

	// Создаем сервис игрока
	playerService := services.NewPlayerService(spritesRepo)

	// Создаем сервис пуль
	bulletsService := services.NewBulletsService()

	// Создаем сервис контроллера
	controllerService := services.NewControllerService(
		playerService,
		bulletsService,
		ebiten.KeyW,     // up
		ebiten.KeyS,     // down
		ebiten.KeyA,     // left
		ebiten.KeyD,     // right
		ebiten.KeySpace, // shoot
	)

	// Создаем GameState с переданными сервисами
	gameState, err := states.NewGameState(
		mapsRepo,
		playerService,
		controllerService,
		bulletsService,
	)
	if err != nil {
		// Логируем ошибку и падаем
		fmt.Printf("Ошибка создания GameState: %v\n", err)
		panic(err)
	}

	return &App{
		config:  cfg,
		service: svc,
		State:   gameState,
	}
}

func (app *App) Run(ctx context.Context) error {
	ebiten.SetWindowSize(constants.ScreenWidth*3, constants.ScreenHeight*3)
	ebiten.SetWindowTitle(app.config.AppConfig.Name)

	if err := ebiten.RunGame(app); err != nil {
		return err
	}

	return nil
}
