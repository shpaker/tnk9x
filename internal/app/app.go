package app

import (
	"context"
	"fmt"
	"log"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/tnk9x/internal/adapters/bindings"
	"github.com/shpaker/tnk9x/internal/adapters/effects"
	"github.com/shpaker/tnk9x/internal/adapters/level_select"
	"github.com/shpaker/tnk9x/internal/adapters/menu_input"
	"github.com/shpaker/tnk9x/internal/adapters/scripting"
	"github.com/shpaker/tnk9x/internal/adapters/settings"
	"github.com/shpaker/tnk9x/internal/adapters/stage"
	"github.com/shpaker/tnk9x/internal/adapters/touch_controls"
	"github.com/shpaker/tnk9x/internal/adapters/window"
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
	config *Config
	state  gameState

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
	menuInput         interfaces.IMenuInputAdapter
	effectsRenderer   *effects.EffectsRendererAdapter
	windowAdapter     interfaces.IWindowAdapter

	// Пользовательские настройки (графика, полный экран, громкость,
	// число игроков) и раскладка управления; их экраны общие
	// для выбора уровня и паузы, хоткеи работают на всех экранах
	settings         *types.SettingsEntity
	controls         *types.ControlsEntity
	settingsUseCases interfaces.ISettingsUseCases
	settingsOverlay  *states.SettingsOverlay
	hotkeysHandler   *states.HotkeysHandler

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

	// Прогресс и настройки — в одном пользовательском хранилище
	storageRepository := newStorageRepository(cfg)
	progressRepository := processed.NewProgressRepository(storageRepository)
	progressionUseCases := use_cases.NewProgressionUseCases(
		campaign,
		loadProgress(progressRepository),
		progressRepository,
	)
	settingsRepository := processed.NewSettingsRepository(storageRepository)
	userSettings := loadSettings(settingsRepository)
	controlsRepository := processed.NewControlsRepository(storageRepository)
	userControls := loadControls(controlsRepository)

	scriptsRepository := processed.NewScriptsRepository(fileRepository)
	soundsRepository := processed.NewSoundsRepository(fileRepository)

	// Звуковой адаптер общий для всех уровней: PCM декодируется один
	// раз при старте, плееры живут на единственном audio.Context
	soundAdapter, err := stage.NewSoundAdapter(
		soundsRepository,
		audioContext,
		userSettings.GetVolume(),
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

	// Ввод меню и экраны настроек общие для выбора уровня и паузы;
	// FULLSCREEN в настройках — только там, где окно можно
	// развернуть при старте (в браузере — только жестом)
	touchControls := touch_controls.NewTouchControlsAdapter(
		cfg.ScreenWidth()/screenScaleFactor,
		cfg.ScreenHeight()/screenScaleFactor,
		userSettings,
	)
	menuInput := menu_input.NewMenuInputAdapter(touchControls)
	windowAdapter := window.NewWindowAdapter()
	settingsUseCases := use_cases.NewSettingsUseCases(
		settingsRepository,
		runtime.GOOS != "js",
	)
	controlsOverlay := states.NewControlsOverlay(
		use_cases.NewControlsUseCases(controlsRepository),
		settings.NewControlsRendererAdapter(
			textFace,
			int(cfg.GetTitleFontSize()),
			int(cfg.GetRegularFontSize()),
		),
		menuInput,
		bindings.NewCaptureAdapter(),
		userControls,
	)
	settingsOverlay := states.NewSettingsOverlay(
		settingsUseCases,
		settings.NewSettingsRendererAdapter(
			textFace,
			int(cfg.GetTitleFontSize()),
			int(cfg.GetRegularFontSize()),
		),
		menuInput,
		soundAdapter,
		windowAdapter,
		controlsOverlay,
		userSettings,
	)

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
		soundAdapter:     soundAdapter,
		touchControls:    touchControls,
		menuInput:        menuInput,
		effectsRenderer:  effectsRenderer,
		windowAdapter:    windowAdapter,
		settings:         userSettings,
		controls:         userControls,
		settingsUseCases: settingsUseCases,
		settingsOverlay:  settingsOverlay,
		hotkeysHandler: states.NewHotkeysHandler(
			settingsUseCases,
			bindings.NewHotkeysAdapter(userControls),
			windowAdapter,
			settingsOverlay,
			userSettings,
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

// newStorageRepository — пользовательское хранилище прогресса
// и настроек: файлы в каталоге конфигурации ОС на десктопе,
// localStorage в браузере
func newStorageRepository(cfg *Config) interfaces.IStorageRepository {
	dir, err := raw.DefaultStorageDir(cfg.GetGameTitle())
	if err != nil {
		log.Printf("storage unavailable: %v", err)
		dir = "."
	}
	return raw.NewStorageRepository(dir)
}

// loadSettings читает настройки; повреждённое или недоступное
// сохранение не мешает игре — начинаем с настроек по умолчанию
func loadSettings(
	repository interfaces.ISettingsRepository,
) *types.SettingsEntity {
	settings, err := repository.GetSettings()
	if err != nil {
		log.Printf("load settings: %v", err)
	}
	return settings
}

// loadControls читает раскладку; повреждённое сохранение не мешает
// игре, а неизвестные движку имена заменяются умолчаниями
func loadControls(
	repository interfaces.IControlsRepository,
) *types.ControlsEntity {
	controls, err := repository.GetControls()
	if err != nil {
		log.Printf("load controls: %v", err)
	}
	bindings.RepairUnknown(controls)
	return controls
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
	return states.NewLevelSelectState(states.LevelSelectStateDependencies{
		LevelSelectUseCases: app.levelSelectUseCases,
		SettingsUseCases:    app.settingsUseCases,
		Renderer:            app.levelSelectRenderer,
		MenuInput:           app.menuInput,
		SettingsOverlay:     app.settingsOverlay,
		Settings:            app.settings,
		Levels:              app.campaignLevels,
		LastLevel:           lastLevel,
		QuitAvailable:       runtime.GOOS != "js",
	})
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

	// Ввод опрашивается один раз на кадр до обновления состояния:
	// тач — первым, меню собирает в том числе его события; затем
	// глобальные хоткеи
	app.touchControls.Update()
	app.menuInput.Update()
	app.hotkeysHandler.Update()

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
		stageSession.SetPlayerCount(app.settings.GetPlayers())
		stageSession.SetStageNumber(transition.Level)
		stageSession.SetCarryOver(transition.CarryOver)
		app.session.Level = int(transition.Level)

		stageState, err := app.newStageState()
		if err != nil {
			return err
		}
		app.state = stageState
	case types.TransitionToLevelSelect:
		app.state = app.newLevelSelectState(int(transition.Level))
	case types.TransitionToQuit:
		// Пункт QUIT меню экрана выбора уровня; в js-сборке выхода
		// нет, поэтому переход возможен только на десктопе
		return ebiten.Termination
	}

	return nil
}

func (app *App) Draw(screen *ebiten.Image) {
	app.state.Draw(screen)
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
	// В браузере полный экран разрешён только по жесту пользователя
	if runtime.GOOS != "js" {
		app.windowAdapter.SetFullscreen(app.settings.IsFullscreen())
	}

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
