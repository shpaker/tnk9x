package app

import (
	"context"
	"fmt"
	"log"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/tnk9x/internal/adapters/effects"
	"github.com/shpaker/tnk9x/internal/adapters/level_select"
	"github.com/shpaker/tnk9x/internal/adapters/scripting"
	"github.com/shpaker/tnk9x/internal/adapters/stage"
	"github.com/shpaker/tnk9x/internal/adapters/touch_controls"
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

// hudFontSize — размер пиксельного шрифта HUD: глифы PressStart2P
// нарисованы сеткой 8x8, целый размер даёт чёткие пиксели без фильтрации
const hudFontSize = 8

// screenScaleFactor — во сколько раз окно крупнее логического экрана NES
// (768x672 -> 256x224)
const screenScaleFactor = 3

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

	// Контекст приложения: отмена (например, SIGINT) завершает игровой цикл
	ctx context.Context

	session *session_entities.GameSessionEntity

	// Долгоживущая инфраструктура — создаётся один раз на приложение
	fileRepository    interfaces.IFileRepository
	tilesetRegistry   interfaces.ITilesetRepositoryRegistry
	spriteCache       *stage.SpriteCache
	mapsRepository    interfaces.IMapsDataRepository
	scriptsRepository interfaces.IScriptsRepository
	soundsRepository  interfaces.ISoundsRepository
	textFace          text.Face
	hudTextFace       text.Face
	scriptEngine      interfaces.IAIScriptEngine
	soundAdapter      *stage.SoundAdapter
	touchControls     *touch_controls.TouchControlsAdapter
	effectsRenderer   *effects.EffectsRendererAdapter

	// Включённость графических эффектов, переключается F2
	effectsSettings *types.EffectsSettingsEntity

	// Уровни кампании для превью экрана выбора; не мутируются —
	// для игры уровень каждый раз читается заново
	campaignLevels      map[int]*types.LevelEntity
	levelSelectRenderer *level_select.LevelSelectRendererAdapter

	// Stateless-сервисы
	randomService            interfaces.IRandomService
	boundaryCollisionService interfaces.IBoundaryCollisionService
	entitiesCollisionService interfaces.IEntitiesCollisionService
	wallCollisionService     interfaces.IWallCollisionService
	bulletCollisionService   interfaces.IBulletCollisionService
	spawnCollisionService    interfaces.ISpawnCollisionService
	tankBrakingService       interfaces.ITankBrakingService

	// Use Cases
	specsUseCases       interfaces.ISpecsUseCases
	waveUseCases        interfaces.IWaveUseCases
	progressionUseCases interfaces.IProgressionUseCases
	levelSelectUseCases interfaces.ILevelSelectUseCases
	debugUseCases       *use_cases.DebugUseCases
}

func New(cfg *Config) *App {
	// Создаем audio context один раз на приложение
	audioContext := audio.NewContext(audioSampleRate)

	fileRepository := raw.NewFileRepository(assetsFS())

	tilesetRegistry, err := processed.NewTilesetRepositoryRegistry(
		fileRepository,
	)
	if err != nil {
		fmt.Printf("Error creating tileset registry: %v\n", err)
		panic(err)
	}

	// Fail-fast проверка спрайтов: все идентификаторы, на которые
	// ссылается код, должны существовать в тайлсетах — падение
	// до открытия окна с полным списком проблем
	spriteManifest := processed.RequiredSprites().
		Merge(use_cases.RequiredSprites()).
		Merge(stage.RequiredSprites()).
		Merge(stageFactoryRequiredSprites())
	spriteValidation := use_cases.NewSpriteValidationUseCases(tilesetRegistry)
	if err := spriteValidation.Validate(spriteManifest); err != nil {
		fmt.Printf("Error validating sprites: %v\n", err)
		panic(err)
	}

	// Кэш GPU-спрайтов общий для всех уровней; прогрев при старте
	// делает стоимость кадра детерминированной
	spriteCache := stage.NewSpriteCache(
		use_cases.NewSpriteUseCases(tilesetRegistry),
	)
	spriteCache.Preload(types.AllTilesetTypes())

	mapsRepository := processed.NewMapsDataRepository(
		fileRepository,
		tilesetRegistry,
		cfg.LevelDefaults,
	)

	// Fail-fast проверка кампании: все уровни существуют и разбираются
	campaign, err := processed.NewCampaignRepository(fileRepository).
		GetCampaign(cfg.Campaign)
	if err != nil {
		fmt.Printf("Error loading campaign: %v\n", err)
		panic(err)
	}
	campaignLevels, err := loadCampaignLevels(
		campaign,
		mapsRepository,
		int(cfg.GetTileBaseSize()),
	)
	if err != nil {
		fmt.Printf("Error loading campaign levels: %v\n", err)
		panic(err)
	}

	progressRepository := newProgressRepository(cfg)
	progressionUseCases := use_cases.NewProgressionUseCases(
		campaign,
		loadProgress(progressRepository),
		progressRepository,
	)
	scriptsRepository := processed.NewScriptsRepository(fileRepository)
	soundsRepository := processed.NewSoundsRepository(fileRepository)

	// Звуковой адаптер общий для всех уровней: PCM декодируется один
	// раз при старте, плееры живут на единственном audio.Context
	soundAdapter, err := stage.NewSoundAdapter(
		soundsRepository,
		audioContext,
		cfg.GetVolume(),
	)
	if err != nil {
		fmt.Printf("Error creating sound adapter: %v\n", err)
		panic(err)
	}

	// Шейдеры компилируются при старте: ошибка в исходнике
	// останавливает запуск до открытия окна
	effectsRenderer, err := effects.NewEffectsRendererAdapter(
		processed.NewShadersRepository(fileRepository),
	)
	if err != nil {
		fmt.Printf("Error creating effects renderer: %v\n", err)
		panic(err)
	}

	fontsRepository := processed.NewFontsRepository(fileRepository)
	textFace, err := buildTextFace(fontsRepository, cfg.GetTitleFontSize())
	if err != nil {
		fmt.Printf("Error creating text face: %v\n", err)
		panic(err)
	}

	hudTextFace, err := buildTextFace(fontsRepository, hudFontSize)
	if err != nil {
		fmt.Printf("Error creating HUD text face: %v\n", err)
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
		spriteCache:       spriteCache,
		mapsRepository:    mapsRepository,
		scriptsRepository: scriptsRepository,
		soundsRepository:  soundsRepository,
		textFace:          textFace,
		hudTextFace:       hudTextFace,
		scriptEngine: scripting.NewLuaEngine(
			services.NewNavigationService(),
		),
		soundAdapter:    soundAdapter,
		effectsRenderer: effectsRenderer,
		effectsSettings: types.NewEffectsSettingsEntity(
			cfg.GetEffectsEnabled(),
		),
		touchControls: touch_controls.NewTouchControlsAdapter(
			cfg.ScreenWidth()/screenScaleFactor,
			cfg.ScreenHeight()/screenScaleFactor,
		),
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
		tankBrakingService:  services.NewTankBrakingService(),
		randomService:       services.NewRandomService(),
		specsUseCases:       use_cases.NewSpecsUseCases(),
		waveUseCases:        use_cases.NewWaveUseCases(),
		progressionUseCases: progressionUseCases,
		levelSelectUseCases: use_cases.NewLevelSelectUseCases(
			progressionUseCases,
		),
		campaignLevels: campaignLevels,
		levelSelectRenderer: level_select.NewLevelSelectRendererAdapter(
			level_select.LevelSelectRendererDependencies{
				FontFace:        textFace,
				TitleFontSize:   int(cfg.GetTitleFontSize()),
				RegularFontSize: int(cfg.GetRegularFontSize()),
				HQPosition: types.Position{
					X: float64(cfg.GetHQPosition()[0]),
					Y: float64(cfg.GetHQPosition()[1]),
				},
				EnemySpawners: cfg.GetEnemySpawners(),
				CellSizePx:    int(cfg.GetBaseSizePx()),
			},
		),
		debugUseCases: use_cases.NewDebugUseCases(Version),
	}

	app.state = app.newLevelSelectState(0)

	return app
}

// loadCampaignLevels читает все уровни кампании: ошибка в любом
// файле останавливает запуск
func loadCampaignLevels(
	campaign *types.CampaignEntity,
	mapsRepository interfaces.IMapsDataRepository,
	tileBaseSize int,
) (map[int]*types.LevelEntity, error) {
	levels := make(map[int]*types.LevelEntity)
	for _, pack := range campaign.GetPacks() {
		for _, number := range pack.Levels {
			level, err := mapsRepository.GetLevel(number, tileBaseSize)
			if err != nil {
				return nil, err
			}
			levels[number] = level
		}
	}
	return levels, nil
}

// newProgressRepository — хранилище прогресса: файл в каталоге
// конфигурации ОС на десктопе, localStorage в браузере
func newProgressRepository(cfg *Config) interfaces.IProgressRepository {
	dir, err := raw.DefaultStorageDir(cfg.GetGameTitle())
	if err != nil {
		log.Printf("storage unavailable: %v", err)
		dir = "."
	}
	return processed.NewProgressRepository(raw.NewStorageRepository(dir))
}

// loadProgress читает прогресс; повреждённое или недоступное
// сохранение не мешает игре — начинаем с чистого прогресса
func loadProgress(
	repository interfaces.IProgressRepository,
) *types.ProgressEntity {
	progress, err := repository.GetProgress()
	if err != nil {
		log.Printf("load progress: %v", err)
	}
	return progress
}

// newLevelSelectState собирает экран выбора уровня; курсор встаёт
// на lastLevel или на последний открытый уровень
func (app *App) newLevelSelectState(lastLevel int) *states.LevelSelectState {
	// В браузере ebiten.Termination заморозил бы canvas,
	// поэтому выход есть только на десктопе
	return states.NewLevelSelectState(
		app.levelSelectUseCases,
		app.levelSelectRenderer,
		app.touchControls,
		app.campaignLevels,
		lastLevel,
		runtime.GOOS != "js",
	)
}

func (app *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return app.config.ScreenWidth() / screenScaleFactor,
		app.config.ScreenHeight() / screenScaleFactor
}

func (app *App) Update() error {
	if app.ctx != nil {
		select {
		case <-app.ctx.Done():
			return ebiten.Termination
		default:
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		Debug = !Debug
		// Обновляем флаг дебаг-режима в текущем игровом состоянии
		if app.stageState != nil {
			app.stageState.SetDebugEnabled(Debug)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		app.effectsSettings.Toggle()
	}

	// Сенсорный ввод опрашивается один раз на кадр до обновления
	// состояния: события кадра общие для стейта и адаптеров
	app.touchControls.Update()

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
		stageSession.SetPlayerCount(app.config.GetPlayerCount())
		stageSession.SetStageNumber(transition.Level)
		app.session.Level = int(transition.Level)

		stageState, err := app.newStageState()
		if err != nil {
			return err
		}
		stageState.SetDebugEnabled(Debug)
		app.state = stageState
		app.stageState = stageState
	case types.TransitionToLevelSelect:
		app.state = app.newLevelSelectState(int(transition.Level))
		app.stageState = nil
	case types.TransitionToQuit:
		// Выход по ESC с экрана выбора уровня; в js-сборке выхода
		// нет, поэтому переход возможен только на десктопе
		return ebiten.Termination
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
	app.ctx = ctx
	defer app.Close()

	ebiten.SetWindowSize(
		app.config.ScreenWidth(),
		app.config.ScreenHeight(),
	)
	ebiten.SetWindowTitle(
		fmt.Sprintf("%s v%s", app.config.GetGameTitle(), Version),
	)

	return ebiten.RunGame(app)
}

func (app *App) Close() {
	if app.scriptEngine != nil {
		app.scriptEngine.Close()
		app.scriptEngine = nil
	}
	if app.soundAdapter != nil {
		app.soundAdapter.StopAll()
	}
}
