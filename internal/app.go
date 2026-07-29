package internal

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
	game_repos "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/repositories/processed"
	"github.com/shpaker/tnk9x/internal/repositories/raw"
	"github.com/shpaker/tnk9x/internal/services"
	collision_services "github.com/shpaker/tnk9x/internal/services/collision_services"
	"github.com/shpaker/tnk9x/internal/states"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"

	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

const audioSampleRate = 44100

// gameState — контракт состояния приложения, определён у потребителя;
// Update возвращает запрос перехода (нулевое значение — остаться)
type gameState interface {
	Update() types.StateTransition
	Draw(screen *ebiten.Image)
}

type App struct {
	config        *Config
	state         gameState
	stageState    *states.StageState
	scriptEngine  interfaces.IAIScriptEngine
	session       *session_entities.GameSessionEntity
	textFace      text.Face
	debugUseCases *use_cases.DebugUseCases
	audioContext  *audio.Context
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

		stageState, err := app.createStageState(app.session)
		if err != nil {
			return err
		}
		app.state = stageState
		app.stageState = stageState
	case types.TransitionToStageSelect:
		selectState, err := app.createStageSelectState()
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
	fontsRepository := processed.NewFontsRepository(fileRepository)
	textFace, err := buildTextFace(fontsRepository, cfg.GetTitleFontSize())
	if err != nil {
		fmt.Printf("Error creating text face: %v\n", err)
		panic(err)
	}

	scriptEngine := scripting.NewLuaEngine()

	session := session_entities.NewGameSessionEntity()

	stageSelectorUseCases := use_cases.NewStageSelectorUseCases()

	stageSelectState, err := states.NewStageSelectState(
		cfg,
		stageSelectorUseCases,
		textFace,
		mapsRepository,
	)
	if err != nil {
		fmt.Printf("Error creating StageSelectState: %v\n", err)
		panic(err)
	}

	return &App{
		config:        cfg,
		state:         stageSelectState,
		scriptEngine:  scriptEngine,
		session:       session,
		textFace:      textFace,
		debugUseCases: use_cases.NewDebugUseCases(Version),
		audioContext:  audioContext,
	}
}

func (app *App) createStageState(
	session *session_entities.GameSessionEntity,
) (*states.StageState, error) {
	fileRepository := raw.NewFileRepository("assets")

	tilesetRegistry, err := processed.NewTilesetRepositoryRegistry(
		fileRepository,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tileset registry: %w", err)
	}

	scriptsRepository := processed.NewScriptsRepository(fileRepository)

	mapsRepository := processed.NewMapsDataRepository(
		fileRepository,
		tilesetRegistry,
	)

	gameRepository := game_repos.NewGameRepositoriesRegistry()

	boundaryCollisionService := collision_services.NewBoundaryCollisionService(
		types.Size{
			Width: app.config.MapBlocksCount.Width * int(
				app.config.TileBaseSize,
			),
			Height: app.config.MapBlocksCount.Height * int(
				app.config.TileBaseSize,
			),
		},
	)
	entitiesCollisionService := collision_services.NewEntitiesCollisionService()
	wallCollisionService := collision_services.NewWallCollisionService(
		entitiesCollisionService,
	)
	tankBrakingService := services.NewTankBrakingService()

	stageStatePtr, err := states.NewStageState(
		mapsRepository,
		scriptsRepository,
		session.Level,
		tilesetRegistry,
		app.config,
		app.textFace,
		gameRepository,
		boundaryCollisionService,
		wallCollisionService,
		tankBrakingService,
		app.scriptEngine,
		session.StageSession(),
		fileRepository,
		app.audioContext,
	)
	if err != nil {
		return nil, err
	}

	// Устанавливаем флаг дебаг-режима
	stageStatePtr.SetDebugEnabled(Debug)

	return stageStatePtr, nil
}

func (app *App) createStageSelectState() (*states.StageSelectState, error) {
	fileRepository := raw.NewFileRepository("assets")

	tilesetRegistry, err := processed.NewTilesetRepositoryRegistry(
		fileRepository,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tileset registry: %w", err)
	}

	mapsRepository := processed.NewMapsDataRepository(
		fileRepository,
		tilesetRegistry,
	)

	stageSelectorUseCases := use_cases.NewStageSelectorUseCases()

	return states.NewStageSelectState(
		app.config,
		stageSelectorUseCases,
		app.textFace,
		mapsRepository,
	)
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
