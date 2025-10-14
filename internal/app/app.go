package app

import (
	"context"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/shpaker/gonflict/internal/config"
	"github.com/shpaker/gonflict/internal/interfaces"
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
	return app.config.AppConfig.ScreenWidth, app.config.AppConfig.ScreenHeight
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
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

func New(cfg *config.Config) *App {
	svc := services.NewWindowService()
	return &App{
		config:  cfg,
		service: svc,
		State:   &states.GameState{},
	}
}

func (app *App) Run(ctx context.Context) error {
	ebiten.SetWindowSize(app.config.AppConfig.ScreenWidth, app.config.AppConfig.ScreenHeight)
	ebiten.SetWindowTitle(app.config.AppConfig.Name)

	if err := ebiten.RunGame(app); err != nil {
		return err
	}

	return nil
}
