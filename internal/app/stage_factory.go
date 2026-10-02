package app

import (
	"fmt"

	"github.com/shpaker/tnk9x/internal/adapters/stage"
	"github.com/shpaker/tnk9x/internal/adapters/stage/input_adapters"
	"github.com/shpaker/tnk9x/internal/interfaces"
	game_repos "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/services"
	"github.com/shpaker/tnk9x/internal/states"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
	"github.com/shpaker/tnk9x/internal/use_cases"
	state_use_cases "github.com/shpaker/tnk9x/internal/use_cases/state_use_cases"
	tank_use_cases "github.com/shpaker/tnk9x/internal/use_cases/tank_use_cases"
)

// Раскладка экрана уровня как в NES: поле 208x208 со смещением (16,8),
// справа остаётся панель HUD шириной 32px
const (
	stageMapOffsetX = 16
	stageMapOffsetY = 8
)

// hqSizePx — размер штаба в пикселях логического экрана
const hqSizePx = 16

// demoSceneLevel — карта демо-сцены главного меню:
// levels/demo/title.bcmap, вне кампании
const demoSceneLevel = "demo/title"

// stageFactoryRequiredSprites перечисляет спрайты, запрашиваемые
// фабрикой уровня (см. createHQ)
func stageFactoryRequiredSprites() types.SpriteManifest {
	return types.SpriteManifest{
		Images: map[types.TilesetType][]string{
			types.TilesetTypeHQ: {"hq_intact"},
		},
	}
}

// stageGraph — граф уровня: общий для игры и демо-сцены главного меню
type stageGraph struct {
	mapUseCases           interfaces.IMapUseCases
	tankTilesUseCases     *use_cases.TilesUseCases
	renderUseCases        interfaces.IRenderUseCases
	tankCommonUseCases    interfaces.ITankCommonUseCases
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases
	tankActionsUseCases   interfaces.ITankActionsUseCases
	bulletUseCases        interfaces.IBulletUseCases
	soundUseCases         interfaces.ISoundUseCases
	visualEffectsUseCases interfaces.IVisualEffectsUseCases
	lightingUseCases      interfaces.ILightingUseCases
	stageUseCases         interfaces.IStageUseCases
	bonusesRepository     interfaces.IBonusesRepository
	enemyInputAdapter     interfaces.IAiInputAdapter
	renderer              *stage.StageRendererAdapter
}

// newStageState собирает уровень кампании: граф уровня со штабом
// и вводом обоих игроков
func (app *App) newStageState() (*states.StageState, error) {
	level, err := app.mapsRepository.GetLevel(
		app.session.Level,
		int(app.config.GetTileBaseSize()),
	)
	if err != nil {
		return nil, err
	}

	stageSession := app.session.StageSession()
	stageSession.SetUpLevel(level)
	graph, err := app.buildStageGraph(level, stageSession, true)
	if err != nil {
		return nil, err
	}

	// Каждый игрок управляется своей раскладкой клавиатуры, своим
	// геймпадом и своими экранными контроллами: композит применяет
	// все источники каждый кадр
	playerInput := func(player types.PlayerTankNum) interfaces.IInputAdapter {
		return input_adapters.NewCompositeInputAdapter(
			input_adapters.NewStageKeyboardInputAdapter(
				graph.tankActionsUseCases, nil, graph.stageUseCases,
				app.controls, player,
			),
			input_adapters.NewStageGamepadInputAdapter(
				graph.tankActionsUseCases, nil, graph.stageUseCases,
				app.controls, player,
			),
			input_adapters.NewStageTouchInputAdapter(
				graph.tankActionsUseCases, nil, graph.stageUseCases,
				app.touchControls, player,
			),
		)
	}

	return states.NewStageState(states.StageStateDependencies{
		TankCommonUseCases:    graph.tankCommonUseCases,
		RenderUseCases:        graph.renderUseCases,
		TankLifecycleUseCases: graph.tankLifecycleUseCases,
		TilesUseCases:         graph.tankTilesUseCases,
		StageUseCases:         graph.stageUseCases,
		SoundUseCases:         graph.soundUseCases,
		LightingUseCases:      graph.lightingUseCases,
		VisualEffectsUseCases: graph.visualEffectsUseCases,
		ProgressionUseCases:   app.progressionUseCases[app.mode()],
		InputAdapters: [2]interfaces.IInputAdapter{
			playerInput(types.PlayerTankNumPlayer1),
			playerInput(types.PlayerTankNumPlayer2),
		},
		EnemyInputAdapter:  graph.enemyInputAdapter,
		Renderer:           graph.renderer,
		SoundPlayerAdapter: app.soundAdapter,
		MenuInput:          app.menuInput,
		RewardAdapter:      app.rewardAdapter,
		StageSession:       stageSession,
		BonusesRepository:  graph.bonusesRepository,
		SettingsOverlay:    app.settingsOverlay,
		Level:              level,
	}), nil
}

// newDemoScene собирает живую сцену главного меню: карта с названием
// игры из блоков, без игроков и без штаба, по танку каждого типа
// врагов под управлением настоящего ИИ. Своя сессия — прогресс
// и перенос кампании не затрагиваются
func (app *App) newDemoScene() (*states.DemoScene, error) {
	level, err := app.mapsRepository.GetSceneLevel(
		demoSceneLevel,
		int(app.config.GetTileBaseSize()),
	)
	if err != nil {
		return nil, err
	}

	stageSession := session_entities.NewStageSessionEntity()
	stageSession.SetUpLevel(level)
	stageSession.Reset()
	titleBlocks := append(types.MapBlocks{}, level.GetMap().GetBlocks()...)

	graph, err := app.buildStageGraph(level, stageSession, false)
	if err != nil {
		return nil, err
	}

	return states.NewDemoScene(states.DemoSceneDependencies{
		TankCommonUseCases:    graph.tankCommonUseCases,
		RenderUseCases:        graph.renderUseCases,
		TankLifecycleUseCases: graph.tankLifecycleUseCases,
		TilesUseCases:         graph.tankTilesUseCases,
		StageUseCases:         graph.stageUseCases,
		SoundUseCases:         graph.soundUseCases,
		LightingUseCases:      graph.lightingUseCases,
		VisualEffectsUseCases: graph.visualEffectsUseCases,
		MapUseCases:           graph.mapUseCases,
		BulletUseCases:        graph.bulletUseCases,
		EnemyInputAdapter:     graph.enemyInputAdapter,
		SoundPlayerAdapter:    app.soundAdapter,
		Renderer:              graph.renderer,
		TitleBlocks:           titleBlocks,
	}), nil
}

// buildStageGraph собирает граф уровня в топологическом порядке:
// репозитории и сервисы приложения переиспользуются, игровое
// runtime-состояние (танки, пули, бонусы, анимации, звуковые события)
// создаётся заново. withHQ=false — без штаба (демо-сцена)
func (app *App) buildStageGraph(
	level *types.LevelEntity,
	stageSession *session_entities.StageSessionEntity,
	withHQ bool,
) (*stageGraph, error) {
	mapEntity := level.GetMap()
	gameRepositories := game_repos.NewGameRepositoriesRegistry()

	// анимации блоков карты (вода) продвигаются общим UpdateAnimations
	for _, block := range mapEntity.GetBlocks() {
		if anim, ok := block.Image.(*image_providers.AnimationProvider); ok {
			gameRepositories.GetAnimationsRepository().AddAnimation(anim)
		}
	}

	soundUseCases := use_cases.NewSoundUseCases(
		gameRepositories.GetSoundEventsRepository(),
	)

	tankTilesUseCases := app.buildTankTilesUseCases(gameRepositories)

	baseSizePx := app.config.GetBaseSizePx()
	bulletUseCases := use_cases.NewBulletUseCases(
		gameRepositories.GetBulletsRepository(),
		app.buildTilesUseCases(
			types.TilesetTypeBullet,
			services.NewAnimationService(),
		),
		baseSizePx,
		app.config.ShotCooldown,
	)

	renderUseCases := use_cases.NewRenderUseCases(tankTilesUseCases)

	mapUseCases := use_cases.NewMapUseCases(mapEntity)

	tankCommonUseCases := tank_use_cases.NewTankCommonUseCases(
		app.tankBrakingService,
		renderUseCases,
		gameRepositories.GetTanksRepository(),
		app.specsUseCases,
		mapUseCases,
		stageSession,
	)

	visualEffectsUseCases := use_cases.NewVisualEffectsUseCases(
		gameRepositories.GetVisualEffectsRepository(),
		app.tilesetRegistry,
		tankCommonUseCases,
		bulletUseCases,
	)

	spawnLayout := types.SpawnLayout{
		EnemySpawners:  app.config.GetEnemySpawners(),
		Player1Spawner: app.config.GetPlayer1Spawn(),
		Player2Spawner: app.config.GetPlayer2Spawn(),
		BaseSize: types.Size{
			Width:  int(baseSizePx),
			Height: int(baseSizePx),
		},
	}

	tankLifecycleUseCases := tank_use_cases.NewTankLifecycleUseCases(
		tankTilesUseCases,
		renderUseCases,
		tankCommonUseCases,
		gameRepositories.GetTanksRepository(),
		app.spawnCollisionService,
		app.specsUseCases,
		visualEffectsUseCases,
		spawnLayout,
	)

	tankActionsUseCases := tank_use_cases.NewTankActionsUseCases(
		app.tankBrakingService,
		bulletUseCases,
		tankCommonUseCases,
		renderUseCases,
		mapUseCases,
		soundUseCases,
		visualEffectsUseCases,
	)

	hqTilesUseCases := app.buildHQTilesUseCases(gameRepositories)
	var hq *types.HQEntity
	if withHQ {
		created, err := app.createHQ(hqTilesUseCases, baseSizePx)
		if err != nil {
			return nil, err
		}
		hq = created
	}
	hqUseCases := use_cases.NewHQUseCases(
		hqTilesUseCases,
		visualEffectsUseCases,
		hq,
	)

	if err := app.loadEnemyAIScript(mapEntity, baseSizePx); err != nil {
		return nil, err
	}

	aiUseCases := use_cases.NewAIUseCases(
		stageSession,
		mapEntity,
		hq,
		gameRepositories.GetTanksRepository(),
		gameRepositories.GetBulletsRepository(),
		gameRepositories.GetBonusesRepository(),
		app.config.EnemyBonusPickup,
		app.scriptEngine,
	)
	enemyInputAdapter := input_adapters.NewAiInputAdapter(
		tankActionsUseCases,
		aiUseCases,
	)

	bonusesRepository := gameRepositories.GetBonusesRepository()

	bonusUseCases := use_cases.NewBonusUseCases(
		tankCommonUseCases,
		tankLifecycleUseCases,
		hqUseCases,
		stageSession,
		mapEntity,
		bonusesRepository,
		app.config,
		app.buildTilesUseCases(
			types.TilesetTypeBonuses,
			services.NewAnimationService(),
		),
		renderUseCases,
		soundUseCases,
	)

	collisionUseCases := use_cases.NewCollisionUseCases(
		bulletUseCases,
		tankActionsUseCases,
		mapUseCases,
		tankCommonUseCases,
		tankLifecycleUseCases,
		app.boundaryCollisionService,
		app.wallCollisionService,
		app.bulletCollisionService,
		app.entitiesCollisionService,
		app.spawnCollisionService,
		hqUseCases,
		bonusUseCases,
		bonusesRepository,
		soundUseCases,
		visualEffectsUseCases,
		app.config.EnemyBonusPickup,
	)

	enemySpawnSelectionUseCases := use_cases.NewEnemySpawnSelectionUseCases(
		tankCommonUseCases,
		app.spawnCollisionService,
		app.randomService,
		spawnLayout,
		app.config.SpawnPlayerSafeRadius,
	)

	stageUseCases := state_use_cases.NewStageUseCases(
		tankLifecycleUseCases,
		app.waveUseCases,
		enemySpawnSelectionUseCases,
		tankCommonUseCases,
		bulletUseCases,
		collisionUseCases,
		hqUseCases,
		stageSession,
		bonusesRepository,
		mapUseCases,
		bonusUseCases,
	)

	lightingUseCases := use_cases.NewLightingUseCases(
		tankCommonUseCases,
		bulletUseCases,
		hqUseCases,
		bonusUseCases,
		visualEffectsUseCases,
	)

	return &stageGraph{
		mapUseCases:           mapUseCases,
		tankTilesUseCases:     tankTilesUseCases,
		renderUseCases:        renderUseCases,
		tankCommonUseCases:    tankCommonUseCases,
		tankLifecycleUseCases: tankLifecycleUseCases,
		tankActionsUseCases:   tankActionsUseCases,
		bulletUseCases:        bulletUseCases,
		soundUseCases:         soundUseCases,
		visualEffectsUseCases: visualEffectsUseCases,
		lightingUseCases:      lightingUseCases,
		stageUseCases:         stageUseCases,
		bonusesRepository:     bonusesRepository,
		enemyInputAdapter:     enemyInputAdapter,
		renderer: app.buildStageRenderer(
			mapUseCases,
			tankCommonUseCases,
			bulletUseCases,
			hqUseCases,
			renderUseCases,
			bonusUseCases,
			lightingUseCases,
			visualEffectsUseCases,
		),
	}, nil
}

// buildTankTilesUseCases — тайлы танков с анимациями спавна и взрыва;
// использует пер-стейдж репозиторий анимаций
func (app *App) buildTankTilesUseCases(
	gameRepositories interfaces.IGameRepositoriesRegistry,
) *use_cases.TilesUseCases {
	tileService := services.NewTileServiceWithEnemyFallback(
		app.tilesetRegistry,
		types.TilesetTypePlayer,
		types.TilesetTypeEnemy,
	)

	return use_cases.NewTilesUseCasesWithAnimations(
		app.tilesetRegistry,
		types.TilesetTypePlayer,
		gameRepositories.GetAnimationsRepository(),
		tileService,
		services.NewAnimationService(),
	)
}

// buildHQTilesUseCases — тайлы штаба с анимацией взрыва
func (app *App) buildHQTilesUseCases(
	gameRepositories interfaces.IGameRepositoriesRegistry,
) *use_cases.TilesUseCases {
	tileService := services.NewTileService(
		app.tilesetRegistry,
		types.TilesetTypePlayer,
	)

	return use_cases.NewTilesUseCasesWithAnimations(
		app.tilesetRegistry,
		types.TilesetTypeHQ,
		gameRepositories.GetAnimationsRepository(),
		tileService,
		services.NewAnimationService(),
	)
}

// buildTilesUseCases — тайлы одного тайлсета без анимаций спавна/взрыва
func (app *App) buildTilesUseCases(
	tilesetType types.TilesetType,
	animationService interfaces.IAnimationService,
) *use_cases.TilesUseCases {
	return use_cases.NewTilesUseCases(
		app.tilesetRegistry,
		tilesetType,
		services.NewTileService(app.tilesetRegistry, tilesetType),
		animationService,
	)
}

func (app *App) createHQ(
	tilesUseCases *use_cases.TilesUseCases,
	baseSizePx uint,
) (*types.HQEntity, error) {
	hqPos := app.config.GetHQPosition()
	if len(hqPos) != 2 {
		return nil, fmt.Errorf("invalid hq_position in config: %v", hqPos)
	}

	hqPosition := types.Position{
		X: float64(hqPos[0]) * float64(baseSizePx),
		Y: float64(hqPos[1]) * float64(baseSizePx),
	}

	imageGetter, err := tilesUseCases.CreateStaticTile("hq_intact")
	if err != nil {
		return nil, fmt.Errorf("failed to create hq tile: %w", err)
	}

	return &types.HQEntity{
		Position: hqPosition,
		Size:     types.Size{Width: hqSizePx, Height: hqSizePx},
		Altitude: types.SURFACE,
		Image:    imageGetter,
		State:    types.HQStateIntact,
	}, nil
}

// loadEnemyAIScript задаёт скрипту глобальные параметры карты
// и загружает сценарий поведения врагов
func (app *App) loadEnemyAIScript(
	mapEntity *types.MapEntity,
	baseSizePx uint,
) error {
	sizePx := mapEntity.GetSizePx()
	tileBaseSize := int(app.config.GetTileBaseSize())
	mapBlocksWidth := sizePx.Width / tileBaseSize
	mapBlocksHeight := sizePx.Height / tileBaseSize

	app.scriptEngine.SetGlobalNumber(
		"MAP_X_BLOCKS_COUNT",
		float64(mapBlocksWidth),
	)
	app.scriptEngine.SetGlobalNumber(
		"MAP_Y_BLOCKS_COUNT",
		float64(mapBlocksHeight),
	)
	app.scriptEngine.SetGlobalNumber("MAP_WIDTH_PX", float64(sizePx.Width))
	app.scriptEngine.SetGlobalNumber("MAP_HEIGHT_PX", float64(sizePx.Height))
	app.scriptEngine.SetGlobalNumber("TANK_SIZE_PX", float64(baseSizePx))
	app.scriptEngine.SetGlobalNumber("BLOCK_SIZE_PX", float64(baseSizePx/2))

	enemyScript, err := app.scriptsRepository.GetScript("enemies")
	if err != nil {
		return err
	}

	return app.scriptEngine.LoadScript(enemyScript)
}

// buildStageRenderer собирает рендер-адаптер уровня с отдельным
// набором тайлов для отрисовки
func (app *App) buildStageRenderer(
	mapUseCases interfaces.IMapUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	bulletUseCases interfaces.IBulletUseCases,
	hqUseCases interfaces.IHQUseCases,
	renderUseCases interfaces.IRenderUseCases,
	bonusUseCases interfaces.IBonusUseCases,
	lightingUseCases interfaces.ILightingUseCases,
	visualEffectsUseCases interfaces.IVisualEffectsUseCases,
) *stage.StageRendererAdapter {
	mapBlocksCount := app.config.GetMapBlocksCount()
	rendererTileSize := int(app.config.GetTileBaseSize())
	mapWidthHeightForAdapter := mapBlocksCount.Width * rendererTileSize

	return stage.NewStageRendererAdapter(stage.StageRendererDependencies{
		MapUseCases:           mapUseCases,
		TankCommonUseCases:    tankCommonUseCases,
		BulletUseCases:        bulletUseCases,
		HQUseCases:            hqUseCases,
		HUDUseCases:           use_cases.NewHUDUseCases(),
		RenderUseCases:        renderUseCases,
		BonusUseCases:         bonusUseCases,
		LightingUseCases:      lightingUseCases,
		VisualEffectsUseCases: visualEffectsUseCases,
		SpriteCache:           app.spriteCache,
		Effects:               app.effectsRenderer,
		Settings:              app.settings,
		FontFace:              app.textFace,
		HUDFontFace:           app.hudTextFace,
		MapOffsetX:            stageMapOffsetX,
		MapOffsetY:            stageMapOffsetY,
		MapWidthHeight:        mapWidthHeightForAdapter,
		TitleFontSize:         int(app.config.GetTitleFontSize()),
		SubtitleFontSize:      int(app.config.GetSubtitleFontSize()),
		RegularFontSize:       int(app.config.GetRegularFontSize()),
	})
}
