package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/tnk9x/internal/adapters/scripting"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/repositories/processed"
	"github.com/shpaker/tnk9x/internal/repositories/raw"
	"github.com/shpaker/tnk9x/internal/services"
	collision_services "github.com/shpaker/tnk9x/internal/services/collision_services"
	"github.com/shpaker/tnk9x/internal/states"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

const audioSampleRate = 44100

// gameState — контракт состояния приложения, определён у потребителя;
// Update возвращает запрос перехода (нулевое значение — остаться)
type gameState interface {
	Update() types.StateTransition
	Draw(screen *ebiten.Image)
}

type App struct {
	config     *Config
	state      gameState
	stageState *states.StageState

	session *session_entities.GameSessionEntity

	// Долгоживущая инфраструктура — создаётся один раз на приложение
	fileRepository    interfaces.IFileRepository
	tilesetRegistry   interfaces.ITilesetRepositoryRegistry
	mapsRepository    interfaces.IMapsDataRepository
	scriptsRepository interfaces.IScriptsRepository
	soundsRepository  interfaces.ISoundsRepository
	textFace          text.Face
	scriptEngine      interfaces.IAIScriptEngine
	audioContext      *audio.Context

	// Stateless-сервисы
	boundaryCollisionService interfaces.IBoundaryCollisionService
	entitiesCollisionService interfaces.IEntitiesCollisionService
	wallCollisionService     interfaces.IWallCollisionService
	bulletCollisionService   interfaces.IBulletCollisionService
	spawnCollisionService    interfaces.ISpawnCollisionService
	tankBrakingService       interfaces.ITankBrakingService

	// Use Cases
	specsUseCases interfaces.ISpecsUseCases
	debugUseCases *use_cases.DebugUseCases
}

func New(cfg *Config) *App {
	// Создаем audio context один раз на приложение
	audioContext := audio.NewContext(audioSampleRate)

	fileRepository := raw.NewFileRepository("assets")

	tilesetRegistry, err := processed.NewTilesetRepositoryRegistry(
		fileRepository,
	)
	if err != nil {
		fmt.Printf("Error creating tileset registry: %v\n", err)
		panic(err)
	}

	mapsRepository := processed.NewMapsDataRepository(
		fileRepository,
		tilesetRegistry,
	)
	scriptsRepository := processed.NewScriptsRepository(fileRepository)
	soundsRepository := processed.NewSoundsRepository(fileRepository)

	fontsRepository := processed.NewFontsRepository(fileRepository)
	textFace, err := buildTextFace(fontsRepository, cfg.GetTitleFontSize())
	if err != nil {
		fmt.Printf("Error creating text face: %v\n", err)
		panic(err)
	}

	mapSizePx := types.Size{
		Width:  cfg.MapBlocksCount.Width * int(cfg.TileBaseSize),
		Height: cfg.MapBlocksCount.Height * int(cfg.TileBaseSize),
	}
	entitiesCollisionService := collision_services.NewEntitiesCollisionService()

	app := &App{
		config:            cfg,
		session:           session_entities.NewGameSessionEntity(),
		fileRepository:    fileRepository,
		tilesetRegistry:   tilesetRegistry,
		mapsRepository:    mapsRepository,
		scriptsRepository: scriptsRepository,
		soundsRepository:  soundsRepository,
		textFace:          textFace,
		scriptEngine:      scripting.NewLuaEngine(),
		audioContext:      audioContext,
		boundaryCollisionService: collision_services.NewBoundaryCollisionService(
			mapSizePx,
		),
		entitiesCollisionService: entitiesCollisionService,
		wallCollisionService: collision_services.NewWallCollisionService(
			entitiesCollisionService,
		),
		bulletCollisionService: collision_services.NewBulletCollisionService(
			int(cfg.GetTileBaseSize()),
			entitiesCollisionService,
		),
		spawnCollisionService: collision_services.NewSpawnCollisionService(
			entitiesCollisionService,
		),
		tankBrakingService: services.NewTankBrakingService(),
		specsUseCases:      use_cases.NewSpecsUseCases(),
		debugUseCases:      use_cases.NewDebugUseCases(Version),
	}

	selectState, err := app.newStageSelectState()
	if err != nil {
		fmt.Printf("Error creating StageSelectState: %v\n", err)
		panic(err)
	}
	app.state = selectState

	return app
}

func (app *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return app.config.ScreenWidth() / 3, app.config.ScreenHeight() / 3
}

func (app *App) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("exit application")
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		Debug = !Debug
		// Обновляем флаг дебаг-режима в текущем игровом состоянии
		if app.stageState != nil {
			app.stageState.SetDebugEnabled(Debug)
		}
	}

	transition := app.state.Update()

	return app.applyTransition(transition)
}

// applyTransition применяет запрошенный стейтом переход:
// записывает параметры в сессию и собирает новое состояние
func (app *App) applyTransition(transition types.StateTransition) error {
	switch transition.Target {
	case types.TransitionNone:
		return nil
	case types.TransitionToStage:
		stageSession := app.session.StageSession()
		if stageSession != nil {
			stageSession.SetMaxActiveEnemies(transition.MaxActiveEnemies)
			stageSession.SetPlayerCount(transition.PlayerCount)
		}
		app.session.Level = int(transition.Level)

		stageState, err := app.newStageState()
		if err != nil {
			return err
		}
		stageState.SetDebugEnabled(Debug)
		app.state = stageState
		app.stageState = stageState
	case types.TransitionToStageSelect:
		selectState, err := app.newStageSelectState()
		if err != nil {
			return err
		}
		app.state = selectState
		app.stageState = nil
	}

	return nil
}

func (app *App) Draw(screen *ebiten.Image) {
	app.state.Draw(screen)
	if app.debugUseCases == nil {
		return
	}

	if !Debug {
		return
	}

	debugText := app.debugUseCases.BuildDebugInfo(app.collectDebugData())
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

// collectDebugData собирает метрики движка и данные сессии для HUD
func (app *App) collectDebugData() types.DebugInfoData {
	data := types.DebugInfoData{
		FPS: ebiten.ActualFPS(),
		TPS: ebiten.ActualTPS(),
	}

	stageSession := app.session.StageSession()
	if stageSession == nil {
		return data
	}

	data.Player1Lives = stageSession.GetPlayerLives(types.PlayerTankNumPlayer1)
	data.Player1InitialLives = stageSession.GetPlayerInitialLives(
		types.PlayerTankNumPlayer1,
	)
	data.Player2Lives = stageSession.GetPlayerLives(types.PlayerTankNumPlayer2)
	data.Player2InitialLives = stageSession.GetPlayerInitialLives(
		types.PlayerTankNumPlayer2,
	)
	data.TotalEnemies = stageSession.GetTotalEnemies()
	data.RemainingEnemies = stageSession.GetRemainingEnemies()

	return data
}

func (app *App) Run(ctx context.Context) error {
	ebiten.SetWindowSize(
		app.config.ScreenWidth(),
		app.config.ScreenHeight(),
	)
	ebiten.SetWindowTitle(
		fmt.Sprintf("%s v%s", app.config.GetGameTitle(), Version),
	)

	err := ebiten.RunGame(app)

	defer app.Close()

	return err
}

func (app *App) Close() {
	if app.scriptEngine != nil {
		app.scriptEngine.Close()
		app.scriptEngine = nil
	}
}
