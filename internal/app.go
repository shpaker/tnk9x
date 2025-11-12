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
	luaEngine     interfaces.ILuaEngine
	session       *session_entities.GameSessionEntity
	fontUseCases  interfaces.IFontUseCases
	debugUseCases *use_cases.DebugUseCases
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
	}

	app.State.Update()

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
	fontUseCases := use_cases.NewFontUseCases(fontsRepository)

	luaEngine := ai.NewLuaEngine()

	session := session_entities.NewGameSessionEntity()

	stageSelectorUseCases := use_cases.NewStageSelectorUseCases()
	stateTransitionUseCases := use_cases.NewStateTransitionUseCases()

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

func (app *App) createStateFromTarget(
	session *session_entities.GameSessionEntity,
	targetState *types.StateType,
) (interfaces.IState, error) {
	if targetState == nil {
		return nil, nil
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

func (app *App) createStageState(
	session *session_entities.GameSessionEntity,
) (interfaces.IState, error) {
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

func (app *App) createStageSelectState(
	session *session_entities.GameSessionEntity,
) (interfaces.IState, error) {
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
	ebiten.SetWindowTitle(
		fmt.Sprintf("%s v%s", app.config.GetGameTitle(), Version),
	)

	err := ebiten.RunGame(app)

	defer app.Close()

	return err
}

func (app *App) Close() {
	if app.luaEngine != nil {
		app.luaEngine.Close()
		app.luaEngine = nil
	}
}
