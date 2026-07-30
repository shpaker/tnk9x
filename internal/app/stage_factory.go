package app

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/adapters/stage"
	"github.com/shpaker/tnk9x/internal/adapters/stage/input_adapters"
	"github.com/shpaker/tnk9x/internal/interfaces"
	game_repos "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/services"
	"github.com/shpaker/tnk9x/internal/states"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
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

// stageFactoryRequiredSprites перечисляет спрайты, запрашиваемые
// фабрикой уровня (см. createHQ)
func stageFactoryRequiredSprites() types.SpriteManifest {
	return types.SpriteManifest{
		Images: map[types.TilesetType][]string{
			types.TilesetTypeHQ: {"hq_intact"},
		},
	}
}

// newStageState собирает граф зависимостей уровня в топологическом порядке:
// репозитории и сервисы приложения переиспользуются, игровое runtime-состояние
// (танки, пули, бонусы, анимации, звуковые события) создаётся заново
func (app *App) newStageState() (*states.StageState, error) {
	tileBaseSize := int(app.config.GetTileBaseSize())
	mapEntity, err := app.mapsRepository.GetLevel(
		int(app.session.RunSession().GetStage()),
		tileBaseSize,
	)
	if err != nil {
		return nil, err
	}
	if mapEntity == nil {
		return nil, fmt.Errorf("map entity is nil")
	}

	stageSession := app.session.StageSession()
	gameRepositories := game_repos.NewGameRepositoriesRegistry()

	// Состав волны и интервал спавна — по таблицам и формуле NES
	stageNumber := app.session.RunSession().GetStage()
	wave, err := app.wavesRepository.GetWave(int(stageNumber))
	if err != nil {
		return nil, err
	}
	stageSession.SetEnemyQueue(wave.Tiers)

	enemySpawnDelay := use_cases.NewWaveUseCases().SpawnDelayTicks(
		stageNumber,
		stageSession.GetPlayerCount(),
	)

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
	// Тайлы пуль с анимациями: взрыв пули тикается общим репозиторием
	bulletTilesUseCases := use_cases.NewTilesUseCasesWithAnimations(
		app.tilesetRegistry,
		types.TilesetTypeBullet,
		gameRepositories.GetAnimationsRepository(),
		services.NewTileService(app.tilesetRegistry, types.TilesetTypeBullet),
		services.NewAnimationService(),
	)
	bulletUseCases := use_cases.NewBulletUseCases(
		gameRepositories.GetBulletsRepository(),
		gameRepositories.GetEffectsRepository(),
		bulletTilesUseCases,
		baseSizePx,
	)

	renderUseCases := use_cases.NewRenderUseCases(tankTilesUseCases)

	mapUseCases := use_cases.NewMapUseCases(mapEntity)

	tankCommonUseCases := tank_use_cases.NewTankCommonUseCases(
		app.tankBrakingService,
		renderUseCases,
		gameRepositories.GetTanksRepository(),
		app.specsUseCases,
		mapUseCases,
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
		spawnLayout,
	)

	tankActionsUseCases := tank_use_cases.NewTankActionsUseCases(
		app.tankBrakingService,
		bulletUseCases,
		tankCommonUseCases,
		renderUseCases,
		mapUseCases,
		soundUseCases,
	)

	hqTilesUseCases := app.buildHQTilesUseCases(gameRepositories)
	hq, err := app.createHQ(hqTilesUseCases, baseSizePx)
	if err != nil {
		return nil, err
	}
	hqUseCases := use_cases.NewHQUseCases(hqTilesUseCases, hq)

	if err := app.loadEnemyAIScript(mapEntity, baseSizePx); err != nil {
		return nil, err
	}

	updateInterval := 60
	if app.config.GetAIUpdateIntervalTicks() > 0 {
		updateInterval = app.config.GetAIUpdateIntervalTicks()
	}

	// Центр штаба — цель врагов в фазе атаки
	hqPos := app.config.GetHQPosition()
	hqCenter := types.Position{
		X: float64(hqPos[0])*float64(baseSizePx) + float64(baseSizePx)/2,
		Y: float64(hqPos[1])*float64(baseSizePx) + float64(baseSizePx)/2,
	}

	aiUseCases := use_cases.NewAIUseCases(app.scriptEngine)
	enemyInputAdapter, err := input_adapters.NewAiInputAdapter(
		tankActionsUseCases,
		nil,
		updateInterval,
		aiUseCases,
		tankCommonUseCases,
		stageSession,
		hqCenter,
	)
	if err != nil {
		return nil, err
	}

	bonusesRepository := gameRepositories.GetBonusesRepository()

	fortressUseCases := use_cases.NewFortressUseCases(
		mapUseCases,
		stageSession,
		app.fortressRingPositions(),
		tileBaseSize,
	)

	bonusUseCases := use_cases.NewBonusUseCases(
		tankCommonUseCases,
		tankLifecycleUseCases,
		stageSession,
		bonusesRepository,
		app.config,
		app.buildTilesUseCases(
			types.TilesetTypeBonuses,
			services.NewAnimationService(),
		),
		renderUseCases,
		soundUseCases,
		mapUseCases,
		app.spawnCollisionService,
		fortressUseCases,
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
	)

	stageUseCases := state_use_cases.NewStageUseCases(
		tankLifecycleUseCases,
		tankCommonUseCases,
		bulletUseCases,
		collisionUseCases,
		hqUseCases,
		stageSession,
		enemySpawnDelay,
		bonusesRepository,
		fortressUseCases,
		soundUseCases,
	)

	inputAdapter1 := input_adapters.NewStageKeyboardInputAdapter(
		tankActionsUseCases,
		nil,
		stageUseCases,
		ebiten.KeyW,
		ebiten.KeyS,
		ebiten.KeyA,
		ebiten.KeyD,
		ebiten.KeySpace,
	)

	inputAdapter2 := input_adapters.NewStageKeyboardInputAdapter(
		tankActionsUseCases,
		nil,
		stageUseCases,
		ebiten.KeyArrowUp,
		ebiten.KeyArrowDown,
		ebiten.KeyArrowLeft,
		ebiten.KeyArrowRight,
		ebiten.KeyEnter,
	)

	rendererAdapter := app.buildStageRenderer(
		mapUseCases,
		tankCommonUseCases,
		bulletUseCases,
		hqUseCases,
		renderUseCases,
		bonusUseCases,
	)

	return states.NewStageState(states.StageStateDependencies{
		TankCommonUseCases:    tankCommonUseCases,
		RenderUseCases:        renderUseCases,
		TankLifecycleUseCases: tankLifecycleUseCases,
		TilesUseCases:         tankTilesUseCases,
		StageUseCases:         stageUseCases,
		SoundUseCases:         soundUseCases,
		InputAdapters: [2]interfaces.IInputAdapter{
			inputAdapter1,
			inputAdapter2,
		},
		EnemyInputAdapter:  enemyInputAdapter,
		Renderer:           rendererAdapter,
		SoundPlayerAdapter: app.soundAdapter,
		StageSession:       stageSession,
		BonusesRepository:  bonusesRepository,
	}), nil
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

// fortressRingPositions возвращает px-координаты 8px-тайлов кольца
// вокруг штаба: ряд сверху и колонки по бокам (низ — край карты)
func (app *App) fortressRingPositions() []types.Position {
	hqPos := app.config.GetHQPosition()
	baseSizePx := int(app.config.GetBaseSizePx())
	tileBaseSize := int(app.config.GetTileBaseSize())
	mapBlocks := app.config.GetMapBlocksCount()

	if tileBaseSize == 0 || len(hqPos) != 2 {
		return nil
	}

	// Координаты штаба в 8px-тайлах и его размер в тайлах
	hqTileX := hqPos[0] * baseSizePx / tileBaseSize
	hqTileY := hqPos[1] * baseSizePx / tileBaseSize
	hqTiles := baseSizePx / tileBaseSize

	var ring []types.Position
	addTile := func(x, y int) {
		if x < 0 || y < 0 || x >= mapBlocks.Width || y >= mapBlocks.Height {
			return
		}
		ring = append(ring, types.Position{
			X: float64(x * tileBaseSize),
			Y: float64(y * tileBaseSize),
		})
	}

	for x := hqTileX - 1; x <= hqTileX+hqTiles; x++ {
		addTile(x, hqTileY-1)
	}
	for y := hqTileY; y < hqTileY+hqTiles; y++ {
		addTile(hqTileX-1, y)
		addTile(hqTileX+hqTiles, y)
	}

	return ring
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
) *stage.StageRendererAdapter {
	mapBlocksCount := app.config.GetMapBlocksCount()
	rendererTileSize := int(app.config.GetTileBaseSize())
	mapWidthHeightForAdapter := mapBlocksCount.Width * rendererTileSize

	return stage.NewStageRendererAdapter(stage.StageRendererDependencies{
		MapUseCases:        mapUseCases,
		TankCommonUseCases: tankCommonUseCases,
		BulletUseCases:     bulletUseCases,
		HQUseCases:         hqUseCases,
		HUDUseCases:        use_cases.NewHUDUseCases(),
		RenderUseCases:     renderUseCases,
		BonusUseCases:      bonusUseCases,
		SpriteCache:        app.spriteCache,
		FontFace:           app.textFace,
		HUDFontFace:        app.hudTextFace,
		MapOffsetX:         stageMapOffsetX,
		MapOffsetY:         stageMapOffsetY,
		MapWidthHeight:     mapWidthHeightForAdapter,
		TitleFontSize:      int(app.config.GetTitleFontSize()),
		SubtitleFontSize:   int(app.config.GetSubtitleFontSize()),
		RegularFontSize:    int(app.config.GetRegularFontSize()),
	})
}
