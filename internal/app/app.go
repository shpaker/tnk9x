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
	"github.com/shpaker/tnk9x/internal/adapters/main_menu"
	"github.com/shpaker/tnk9x/internal/adapters/menu_input"
	"github.com/shpaker/tnk9x/internal/adapters/platform"
	"github.com/shpaker/tnk9x/internal/adapters/scripting"
	"github.com/shpaker/tnk9x/internal/adapters/settings"
	"github.com/shpaker/tnk9x/internal/adapters/shop"
	"github.com/shpaker/tnk9x/internal/adapters/splash"
	"github.com/shpaker/tnk9x/internal/adapters/stage"
	"github.com/shpaker/tnk9x/internal/adapters/texts"
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

// screenScaleFactor — во сколько раз окно крупнее логического экрана
// (768x672 -> 256x224)
const screenScaleFactor = 3

// Логотип сплеша: картинка assets/images/<splashLogoName>.png, пока её
// нет — надпись splashLogoText
const (
	splashLogoName = "shpaker"
	splashLogoText = "SHPAKER"
)

// progressKeys — ключи прогресса по режимам: у одиночной игры и игры
// вдвоём раздельное прохождение
var progressKeys = [types.MaxPlayers]string{
	processed.ProgressKeyOnePlayer,
	processed.ProgressKeyTwoPlayers,
}

// gameState — контракт состояния приложения, определён у потребителя;
// Update возвращает запрос перехода (нулевое значение — остаться)
type gameState interface {
	Update() types.StateTransition
	Draw(screen *ebiten.Image)
}

// frameInput — источник ввода, опрашиваемый раз в кадр до состояния
type frameInput interface {
	Update()
}

// Необязательные возможности состояния, определены у потребителя:
// состояния сообщают факты, а площадке их передаёт App
type (
	// suspendable — состояние реагирует на приостановку игры
	// площадкой (уровень уходит на паузу)
	suspendable interface {
		Suspend()
	}
	// gameplayReporter — состояние сообщает, идёт ли активный геймплей
	gameplayReporter interface {
		IsGameplayActive() bool
	}
)

var (
	_ suspendable      = (*states.StageState)(nil)
	_ gameplayReporter = (*states.StageState)(nil)
)

// App — composition root. New собирает только то, что нужно сплешу;
// остальной граф достраивают шаги загрузчика на сплеше, а стейты,
// которым он нужен, собираются только после её завершения
// (переход TransitionToMainMenu отдаёт сплеш)
type App struct {
	config *Config
	state  gameState

	// Контекст приложения: отмена (например, SIGINT) завершает игровой цикл
	ctx context.Context

	session *session_entities.GameSessionEntity

	// Источники ввода кадра по порядку: тач, меню, затем хоткеи
	// (добавляются в конце загрузки)
	frameInputs []frameInput
	loader      *bootLoader

	// Старт: собирается в New
	audioContext       *audio.Context
	fileRepository     interfaces.IFileRepository
	storageRepository  interfaces.IStorageRepository
	settingsRepository interfaces.ISettingsRepository
	controlsRepository interfaces.IControlsRepository
	textFace           text.Face
	hudTextFace        text.Face
	touchControls      *touch_controls.TouchControlsAdapter
	menuInput          interfaces.IMenuInputAdapter
	effectsRenderer    *effects.EffectsRendererAdapter
	windowAdapter      interfaces.IWindowAdapter
	// Площадка запуска: десктоп или браузер (страница, портал);
	// один адаптер реализует все три контракта
	platformAdapter interfaces.IPlatformAdapter
	rewardAdapter   interfaces.IRewardAdapter
	purchaseAdapter interfaces.IPurchaseAdapter

	// Купленное игроком: общее для всех режимов
	inventory           *types.InventoryEntity
	inventoryRepository interfaces.IInventoryRepository
	inventoryUseCases   interfaces.IInventoryUseCases

	// Пользовательские настройки (графика, полный экран, громкость,
	// язык, режим) и раскладка управления
	settings *types.SettingsEntity
	controls *types.ControlsEntity

	// Тексты интерфейса на языке из настроек или языке площадки
	texts                interfaces.ITextsAdapter
	localizationUseCases interfaces.ILocalizationUseCases

	// Загружается на сплеше: долгоживущая инфраструктура
	tilesetRegistry   interfaces.ITilesetRepositoryRegistry
	spriteCache       *stage.SpriteCache
	mapsRepository    interfaces.IMapsDataRepository
	scriptsRepository interfaces.IScriptsRepository
	soundsRepository  interfaces.ISoundsRepository
	scriptEngine      interfaces.IAIScriptEngine
	soundAdapter      *stage.SoundAdapter

	// Экраны настроек общие для главного меню и паузы, хоткеи
	// работают на всех экранах после загрузки
	settingsUseCases interfaces.ISettingsUseCases
	settingsOverlay  *states.SettingsOverlay
	helpOverlay      *states.HelpOverlay

	// Магазин главного меню
	shopUseCases   interfaces.IShopUseCases
	shopOverlay    *states.ShopOverlay
	hotkeysHandler *states.HotkeysHandler

	// Кампания; уровни для превью экрана выбора не мутируются —
	// для игры уровень каждый раз читается заново
	campaign            *types.CampaignEntity
	campaignLevels      map[int]*types.LevelEntity
	levelSelectRenderer *level_select.LevelSelectRendererAdapter
	mainMenuRenderer    *main_menu.MainMenuRendererAdapter

	// Stateless-сервисы
	randomService            interfaces.IRandomService
	boundaryCollisionService interfaces.IBoundaryCollisionService
	entitiesCollisionService interfaces.IEntitiesCollisionService
	wallCollisionService     interfaces.IWallCollisionService
	bulletCollisionService   interfaces.IBulletCollisionService
	spawnCollisionService    interfaces.ISpawnCollisionService

	// Use Cases; прогресс и выбор уровня — свои у каждого режима
	specsUseCases       interfaces.ISpecsUseCases
	waveUseCases        interfaces.IWaveUseCases
	progressionUseCases [types.MaxPlayers]interfaces.IProgressionUseCases
	levelSelectUseCases [types.MaxPlayers]interfaces.ILevelSelectUseCases
}

// New собирает минимум для сплеша: хранилище, настройки и раскладку,
// тексты на языке игрока, шрифты, эффекты финального экрана и ввод
// меню; ресурсы игры грузятся на сплеше по шагу за кадр
func New(cfg *Config) *App {
	fileRepository := raw.NewFileRepository(assetsFS())
	storageRepository := newStorageRepository(cfg)
	settingsRepository := processed.NewSettingsRepository(storageRepository)
	userSettings := loadSettings(settingsRepository)
	controlsRepository := processed.NewControlsRepository(storageRepository)
	userControls := loadControls(controlsRepository)
	inventoryRepository := processed.NewInventoryRepository(storageRepository)
	inventory := loadInventory(inventoryRepository)

	// Шейдеры нужны финальному экрану с первого кадра: ошибка
	// в исходнике останавливает запуск до открытия окна
	effectsRenderer, err := effects.NewEffectsRendererAdapter(
		processed.NewShadersRepository(fileRepository),
	)
	mustLoad("creating effects renderer", err)

	fontsRepository := processed.NewFontsRepository(fileRepository)
	textFace, err := buildTextFace(fontsRepository, cfg.GetTitleFontSize())
	mustLoad("creating text face", err)
	hudTextFace, err := buildTextFace(fontsRepository, hudFontSize)
	mustLoad("creating HUD text face", err)

	touchControls := touch_controls.NewTouchControlsAdapter(
		cfg.ScreenWidth()/screenScaleFactor,
		cfg.ScreenHeight()/screenScaleFactor,
		userSettings,
	)
	menuInput := menu_input.NewMenuInputAdapter(touchControls)
	platformAdapter := platform.NewPlatformAdapter()

	// Язык определяется при запуске, до первого кадра: язык из
	// настроек, иначе язык площадки
	textsAdapter, err := texts.NewTextsAdapter(
		processed.NewTextsRepository(fileRepository),
		cfg.Languages,
	)
	mustLoad("loading texts", err)
	localizationUseCases := use_cases.NewLocalizationUseCases(
		textsAdapter,
		cfg.Languages,
		platformAdapter.GetLanguage(),
	)
	mustLoad("applying language", localizationUseCases.Apply(userSettings))

	app := &App{
		config:             cfg,
		session:            session_entities.NewGameSessionEntity(),
		frameInputs:        []frameInput{touchControls, menuInput},
		audioContext:       audio.NewContext(audioSampleRate),
		fileRepository:     fileRepository,
		storageRepository:  storageRepository,
		settingsRepository: settingsRepository,
		controlsRepository: controlsRepository,
		textFace:           textFace,
		hudTextFace:        hudTextFace,
		touchControls:      touchControls,
		menuInput:          menuInput,
		effectsRenderer:    effectsRenderer,
		windowAdapter:      window.NewWindowAdapter(),
		platformAdapter:    platformAdapter,
		rewardAdapter:      platformAdapter,
		purchaseAdapter:    platformAdapter,
		settings:           userSettings,
		controls:           userControls,

		texts:                textsAdapter,
		localizationUseCases: localizationUseCases,

		inventory:           inventory,
		inventoryRepository: inventoryRepository,
		inventoryUseCases: use_cases.NewInventoryUseCases(
			inventory, inventoryRepository,
		),
	}

	app.loader = newBootLoader(
		app.loadSprites,
		app.preloadSprites,
		app.loadCampaign,
		app.loadProgress,
		app.loadScripts,
		app.loadSounds,
		app.assembleGame,
	)
	app.state = states.NewSplashState(
		app.loader,
		splash.NewSplashRendererAdapter(app.splashLogo()),
		menuInput,
	)

	return app
}

// mustLoad останавливает запуск при ошибке ресурса: проблемы данных
// видны сразу, а не посреди игры
func mustLoad(what string, err error) {
	if err != nil {
		fmt.Printf("Error %s: %v\n", what, err)
		panic(err)
	}
}

// splashLogo — логотип дизайнера из assets/images, пока его нет —
// текстовая заглушка
func (app *App) splashLogo() splash.Logo {
	img, err := processed.NewImagesRepository(app.fileRepository).
		GetImage(splashLogoName)
	if err != nil {
		log.Printf("splash logo: %v; drawing text placeholder", err)
		return splash.NewTextLogo(app.textFace, splashLogoText)
	}
	return splash.NewImageLogo(img)
}

// loadSprites — тайлсеты и fail-fast проверка спрайтов: все
// идентификаторы, на которые ссылается код, существуют в тайлсетах
func (app *App) loadSprites() {
	tilesetRegistry, err := processed.NewTilesetRepositoryRegistry(
		app.fileRepository,
	)
	mustLoad("creating tileset registry", err)

	spriteManifest := processed.RequiredSprites().
		Merge(use_cases.RequiredSprites()).
		Merge(stage.RequiredSprites()).
		Merge(stageFactoryRequiredSprites())
	mustLoad(
		"validating sprites",
		use_cases.NewSpriteValidationUseCases(tilesetRegistry).
			Validate(spriteManifest),
	)
	app.tilesetRegistry = tilesetRegistry
}

// preloadSprites — кэш GPU-спрайтов общий для всех уровней; прогрев
// делает стоимость кадра детерминированной
func (app *App) preloadSprites() {
	app.spriteCache = stage.NewSpriteCache(
		use_cases.NewSpriteUseCases(app.tilesetRegistry),
	)
	app.spriteCache.Preload(types.AllTilesetTypes())
}

// loadCampaign — карты и fail-fast проверка кампании: все уровни
// существуют и разбираются
func (app *App) loadCampaign() {
	app.mapsRepository = processed.NewMapsDataRepository(
		app.fileRepository,
		app.tilesetRegistry,
		app.config.LevelDefaults,
	)
	campaign, err := processed.NewCampaignRepository(app.fileRepository).
		GetCampaign(app.config.Campaign)
	mustLoad("loading campaign", err)
	campaignLevels, err := loadCampaignLevels(
		campaign,
		app.mapsRepository,
		int(app.config.GetTileBaseSize()),
	)
	mustLoad("loading campaign levels", err)
	mustLoad(
		"validating shop products",
		validateProductPacks(app.config.Products, campaign),
	)

	app.campaign = campaign
	app.campaignLevels = campaignLevels
}

// loadProgress — прогресс кампании каждого режима в своём ключе
// хранилища
func (app *App) loadProgress() {
	for mode, key := range progressKeys {
		repository := processed.NewProgressRepository(
			app.storageRepository, key,
		)
		progression := use_cases.NewProgressionUseCases(
			app.campaign,
			loadProgress(repository),
			app.inventory,
			repository,
		)
		app.progressionUseCases[mode] = progression
		app.levelSelectUseCases[mode] = use_cases.NewLevelSelectUseCases(
			progression,
		)
	}
}

func (app *App) loadScripts() {
	app.scriptsRepository = processed.NewScriptsRepository(app.fileRepository)
	app.scriptEngine = scripting.NewLuaEngine(services.NewNavigationService())
}

// loadSounds — звуковой адаптер общий для всех уровней: PCM
// декодируется один раз, плееры живут на единственном audio.Context
func (app *App) loadSounds() {
	app.soundsRepository = processed.NewSoundsRepository(app.fileRepository)
	soundAdapter, err := stage.NewSoundAdapter(
		app.soundsRepository,
		app.audioContext,
		app.settings.GetVolume(),
	)
	mustLoad("creating sound adapter", err)
	app.soundAdapter = soundAdapter
}

// assembleGame достраивает граф: сервисы, use cases, экраны настроек
// и раскладки, хоткеи и рендеры меню. FULLSCREEN в настройках —
// только там, где окно можно развернуть при старте (в браузере —
// только жестом)
func (app *App) assembleGame() {
	cfg := app.config
	mapSizePx := types.Size{
		Width:  cfg.MapBlocksCount.Width * int(cfg.TileBaseSize),
		Height: cfg.MapBlocksCount.Height * int(cfg.TileBaseSize),
	}
	entitiesCollisionService := collision_services.NewEntitiesCollisionService()
	app.boundaryCollisionService = collision_services.NewBoundaryCollisionService(
		mapSizePx,
	)
	app.entitiesCollisionService = entitiesCollisionService
	app.wallCollisionService = collision_services.NewWallCollisionService(
		entitiesCollisionService,
	)
	app.bulletCollisionService = collision_services.NewBulletCollisionService(
		int(cfg.GetTileBaseSize()),
		entitiesCollisionService,
	)
	app.spawnCollisionService = collision_services.NewSpawnCollisionService(
		entitiesCollisionService,
	)
	app.randomService = services.NewRandomService()
	app.specsUseCases = use_cases.NewSpecsUseCases()
	app.waveUseCases = use_cases.NewWaveUseCases()

	app.settingsUseCases = use_cases.NewSettingsUseCases(
		app.settingsRepository,
		app.texts,
		runtime.GOOS != "js",
		cfg.Languages.Languages,
	)
	controlsUseCases := use_cases.NewControlsUseCases(app.controlsRepository)
	controlsOverlay := states.NewControlsOverlay(
		controlsUseCases,
		settings.NewControlsRendererAdapter(
			app.texts,
			app.textFace,
			int(cfg.GetTitleFontSize()),
			int(cfg.GetRegularFontSize()),
		),
		app.menuInput,
		bindings.NewCaptureAdapter(),
		app.controls,
	)
	app.settingsOverlay = states.NewSettingsOverlay(
		app.settingsUseCases,
		app.localizationUseCases,
		settings.NewSettingsRendererAdapter(
			app.texts,
			app.textFace,
			int(cfg.GetTitleFontSize()),
			int(cfg.GetRegularFontSize()),
		),
		app.menuInput,
		app.soundAdapter,
		app.windowAdapter,
		controlsOverlay,
		app.settings,
	)
	app.helpOverlay = states.NewHelpOverlay(
		app.settingsUseCases,
		controlsUseCases,
		settings.NewHelpRendererAdapter(
			app.texts,
			app.textFace,
			int(cfg.GetTitleFontSize()),
			int(cfg.GetRegularFontSize()),
		),
		app.menuInput,
		app.settings,
		app.controls,
	)
	app.hotkeysHandler = states.NewHotkeysHandler(
		app.settingsUseCases,
		bindings.NewHotkeysAdapter(app.controls),
		app.windowAdapter,
		app.settingsOverlay,
		app.settings,
	)
	app.frameInputs = append(app.frameInputs, app.hotkeysHandler)

	app.levelSelectRenderer = level_select.NewLevelSelectRendererAdapter(
		level_select.LevelSelectRendererDependencies{
			Texts:           app.texts,
			FontFace:        app.textFace,
			TitleFontSize:   int(cfg.GetTitleFontSize()),
			RegularFontSize: int(cfg.GetRegularFontSize()),
			HQPosition: types.Position{
				X: float64(cfg.GetHQPosition()[0]),
				Y: float64(cfg.GetHQPosition()[1]),
			},
			EnemySpawners: cfg.GetEnemySpawners(),
			CellSizePx:    int(cfg.GetBaseSizePx()),
		},
	)
	app.mainMenuRenderer = main_menu.NewMainMenuRendererAdapter(
		app.texts,
		app.textFace,
		int(cfg.GetTitleFontSize()),
		int(cfg.GetRegularFontSize()),
	)
	app.assembleShop()
}

// assembleShop — магазин главного меню и зачисление покупок,
// сделанных до запуска (в том числе не списанных площадкой)
func (app *App) assembleShop() {
	cfg := app.config
	app.shopUseCases = use_cases.NewShopUseCases(
		cfg.Products,
		app.campaign,
		app.inventory,
		app.progressionUseCases[:],
		app.purchaseAdapter,
		app.inventoryRepository,
	)
	if err := app.shopUseCases.Sync(); err != nil {
		log.Printf("sync purchases: %v", err)
	}
	app.shopOverlay = states.NewShopOverlay(
		app.shopUseCases,
		app.inventoryUseCases,
		shop.NewShopRendererAdapter(
			app.texts,
			app.textFace,
			int(cfg.GetTitleFontSize()),
			int(cfg.GetRegularFontSize()),
		),
		app.menuInput,
		app.soundAdapter,
	)
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

// newStorageRepository — пользовательское хранилище прогресса,
// настроек и раскладки: файлы в каталоге конфигурации ОС на десктопе,
// хранилище площадки (мост window.tnk9xPlatform) в браузере
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

// loadInventory читает купленное; повреждённое сохранение
// не мешает игре, покупки площадки зачислятся при синхронизации
func loadInventory(
	repository interfaces.IInventoryRepository,
) *types.InventoryEntity {
	inventory, err := repository.GetInventory()
	if err != nil {
		log.Printf("load inventory: %v", err)
	}
	return inventory
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

// mode — индекс текущего режима: 0 — один игрок, 1 — двое
func (app *App) mode() int {
	return int(app.settings.GetPlayers()) - 1
}

// newMainMenuState собирает главное меню над живой сценой с названием
// из блоков; в браузере ebiten.Termination заморозил бы canvas,
// поэтому выход есть только на десктопе, а магазин — только там,
// где площадка продаёт товары
func (app *App) newMainMenuState() (*states.MainMenuState, error) {
	scene, err := app.newDemoScene()
	if err != nil {
		return nil, err
	}
	// Площадка могла ответить о покупках позже запуска игры
	if err := app.shopUseCases.Sync(); err != nil {
		log.Printf("sync purchases: %v", err)
	}

	return states.NewMainMenuState(states.MainMenuStateDependencies{
		SettingsUseCases: app.settingsUseCases,
		Renderer:         app.mainMenuRenderer,
		MenuInput:        app.menuInput,
		Scene:            scene,
		SettingsOverlay:  app.settingsOverlay,
		ShopOverlay:      app.shopOverlay,
		Settings:         app.settings,
		QuitAvailable:    runtime.GOOS != "js",
		ShopAvailable:    app.shopUseCases.IsAvailable(),
	}), nil
}

// newLevelSelectState собирает экран выбора уровня кампании текущего
// режима; курсор встаёт на lastLevel или на последний открытый уровень
func (app *App) newLevelSelectState(lastLevel int) *states.LevelSelectState {
	return states.NewLevelSelectState(states.LevelSelectStateDependencies{
		LevelSelectUseCases: app.levelSelectUseCases[app.mode()],
		Renderer:            app.levelSelectRenderer,
		MenuInput:           app.menuInput,
		Settings:            app.settings,
		Levels:              app.campaignLevels,
		LastLevel:           lastLevel,
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

	// Площадка опрашивается первой: пока она приостановила игру
	// (реклама, скрытая вкладка), кадр пропускается целиком — ни
	// ввода, ни обновления состояния. Начало приостановки отдаётся
	// состоянию: уровень уходит на паузу
	app.platformAdapter.Update()
	if app.platformAdapter.IsJustSuspended() {
		if state, ok := app.state.(suspendable); ok {
			state.Suspend()
		}
	}
	if app.platformAdapter.IsSuspended() {
		app.platformAdapter.SetGameplayActive(false)
		app.windowAdapter.SetCursorVisible(true)
		return nil
	}

	// Ввод опрашивается один раз на кадр до обновления состояния:
	// тач — первым, меню собирает в том числе его события; затем
	// глобальные хоткеи
	for _, input := range app.frameInputs {
		input.Update()
	}

	transition := app.state.Update()
	err := app.applyTransition(transition)
	gameplayActive := app.isGameplayActive()
	app.platformAdapter.SetGameplayActive(gameplayActive)
	// Мышь управляет только меню: в бою курсор спрятан
	app.windowAdapter.SetCursorVisible(!gameplayActive)

	return err
}

// isGameplayActive — идёт ли в текущем состоянии активный геймплей
func (app *App) isGameplayActive() bool {
	state, ok := app.state.(gameplayReporter)
	return ok && state.IsGameplayActive()
}

// applyTransition применяет запрошенный стейтом переход:
// записывает параметры в сессию и собирает новое состояние
func (app *App) applyTransition(transition types.StateTransition) error {
	// Логическая пауза: площадка может показать рекламу до первого
	// кадра нового состояния; купленное «без рекламы» её отключает
	if transition.Intermission {
		app.platformAdapter.SetGameplayActive(false)
		if !app.inventoryUseCases.HasNoAds() {
			app.platformAdapter.RequestIntermission()
		}
	}
	// Победа: после нескольких за сессию площадка может попросить
	// оценить игру
	if transition.Victory {
		app.session.AddVictory()
		after := app.config.RatingAfterVictories
		if after > 0 && app.session.GetVictories() >= after {
			app.platformAdapter.RequestRating()
		}
	}

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
	case types.TransitionToMainMenu:
		mainMenu, err := app.newMainMenuState()
		if err != nil {
			return err
		}
		app.state = mainMenu
		// Главное меню — первый экран, где игрок может играть:
		// загрузка позади (повторные вызовы площадка игнорирует)
		app.platformAdapter.Ready()
	case types.TransitionToQuit:
		// Пункт QUIT главного меню; в js-сборке выхода нет,
		// поэтому переход возможен только на десктопе
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

// Close освобождает ресурсы; игру могут закрыть ещё на сплеше,
// до загрузки скриптов и звуков
func (app *App) Close() {
	if app.scriptEngine != nil {
		app.scriptEngine.Close()
		app.scriptEngine = nil
	}
	if app.soundAdapter != nil {
		app.soundAdapter.StopAll()
	}
}
