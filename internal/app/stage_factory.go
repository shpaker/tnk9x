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
	"github.com/shpaker/tnk9x/internal/use_cases"
	state_use_cases "github.com/shpaker/tnk9x/internal/use_cases/state_use_cases"
	tank_use_cases "github.com/shpaker/tnk9x/internal/use_cases/tank_use_cases"
)

// newStageState собирает граф зависимостей уровня в топологическом порядке:
// репозитории и сервисы приложения переиспользуются, игровое runtime-состояние
// (танки, пули, бонусы, анимации, звуковые события) создаётся заново
func (app *App) newStageState() (*states.StageState, error) {
	tileBaseSize := int(app.config.GetTileBaseSize())
	mapEntity, err := app.mapsRepository.GetLevel(
		app.session.Level,
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

	soundAdapter, err := stage.NewSoundAdapter(
		app.soundsRepository,
		app.audioContext,
		app.config.GetVolume(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sound adapter: %w", err)
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
	)

	renderUseCases := use_cases.NewRenderUseCases(tankTilesUseCases)

	tankCommonUseCases := tank_use_cases.NewTankCommonUseCases(
		app.tankBrakingService,
		renderUseCases,
		gameRepositories.GetTanksRepository(),
		app.specsUseCases,
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

	mapUseCases := use_cases.NewMapUseCases(mapEntity)

	tankActionsUseCases := tank_use_cases.NewTankActionsUseCases(
		app.tankBrakingService,
		bulletUseCases,
		tankCommonUseCases,
		renderUseCases,
		mapUseCases,
		soundUseCases,
	)

	hqTilesUseCases := app.buildHQTilesUseCases(gameRepositories)
	hq := app.createHQ(hqTilesUseCases, baseSizePx)
	var hqUseCases interfaces.IHQUseCases
	if hq != nil {
		hqUseCases = use_cases.NewHQUseCases(hqTilesUseCases, hq)
	}

	if err := app.loadEnemyAIScript(mapEntity, baseSizePx); err != nil {
		return nil, err
	}

	updateInterval := 60
	if app.config.GetAIUpdateIntervalTicks() > 0 {
		updateInterval = app.config.GetAIUpdateIntervalTicks()
	}

	aiUseCases := use_cases.NewAIUseCases(app.scriptEngine)
	enemyInputAdapter, err := input_adapters.NewAiInputAdapter(
		tankActionsUseCases,
		nil,
		updateInterval,
		aiUseCases,
	)
	if err != nil {
		return nil, err
	}

	bonusesRepository := gameRepositories.GetBonusesRepository()

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
		app.config.GetEnemyRespawnDelayTicks(),
		bonusesRepository,
		mapUseCases,
		bonusUseCases,
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
		ebiten.KeyP,
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
		ebiten.KeyP,
	)

	rendererAdapter := app.buildStageRenderer(
		mapUseCases,
		tankCommonUseCases,
		bulletUseCases,
		hqUseCases,
		bonusesRepository,
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
		SoundPlayerAdapter: soundAdapter,
		StageSession:       stageSession,
		BonusesRepository:  bonusesRepository,
	}), nil
}

// buildTankTilesUseCases — тайлы танков с анимациями спавна и взрыва;
// использует пер-стейдж репозиторий анимаций
func (app *App) buildTankTilesUseCases(
	gameRepositories interfaces.IGameRepositoriesRegistry,
) *use_cases.TilesUseCases {
	tileService := services.NewTileServiceWithSpecialRepos(
		app.tilesetRegistry,
		types.TilesetTypePlayer,
		types.TilesetTypeEnemy,
		types.TilesetTypeSpawner,
		types.TilesetTypeExplosion,
	)

	return use_cases.NewTilesUseCasesWithAnimations(
		app.tilesetRegistry,
		types.TilesetTypePlayer,
		gameRepositories.GetAnimationsRepository(),
		types.TilesetTypeSpawner,
		types.TilesetTypeExplosion,
		tileService,
		services.NewAnimationService(),
	)
}

// buildHQTilesUseCases — тайлы штаба с анимацией взрыва
func (app *App) buildHQTilesUseCases(
	gameRepositories interfaces.IGameRepositoriesRegistry,
) *use_cases.TilesUseCases {
	tileService := services.NewTileServiceWithSpecialRepos(
		app.tilesetRegistry,
		types.TilesetTypePlayer,
		types.TilesetType(""),
		types.TilesetTypeSpawner,
		types.TilesetTypeExplosion,
	)

	return use_cases.NewTilesUseCasesWithAnimations(
		app.tilesetRegistry,
		types.TilesetTypeHQ,
		gameRepositories.GetAnimationsRepository(),
		types.TilesetTypeSpawner,
		types.TilesetTypeExplosion,
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
) *types.HQEntity {
	hqPos := app.config.GetHQPosition()
	if len(hqPos) != 2 {
		return nil
	}

	hqPosition := types.Position{
		X: float64(hqPos[0]) * float64(baseSizePx),
		Y: float64(hqPos[1]) * float64(baseSizePx),
	}

	var imageGetter types.IImageProvider
	if tilesUseCases != nil {
		hqImageGetter, err := tilesUseCases.CreateStaticTile("hq_intact")
		if err == nil {
			imageGetter = hqImageGetter
		}
	}

	return &types.HQEntity{
		Position: hqPosition,
		Size:     types.Size{Width: 16, Height: 16},
		Altitude: types.SURFACE,
		Image:    imageGetter,
		State:    types.HQStateIntact,
	}
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
	bonusesRepository interfaces.IBonusesRepository,
) *stage.StageRendererAdapter {
	renderAnimationService := services.NewAnimationService()

	mapBlocksCount := app.config.GetMapBlocksCount()
	rendererTileSize := int(app.config.GetTileBaseSize())
	mapWidthHeightForAdapter := mapBlocksCount.Width * rendererTileSize

	mapOffsetY := 16
	mapOffsetX := mapWidthHeightForAdapter / 2

	return stage.NewStageRendererAdapter(
		mapUseCases,
		tankCommonUseCases,
		bulletUseCases,
		app.buildTilesUseCases(types.TilesetTypeBlocks, renderAnimationService),
		app.buildTilesUseCases(types.TilesetTypePlayer, renderAnimationService),
		app.buildTilesUseCases(types.TilesetTypeBullet, renderAnimationService),
		app.buildTilesUseCases(
			types.TilesetTypeSpawner,
			renderAnimationService,
		),
		app.buildTilesUseCases(
			types.TilesetTypeExplosion,
			renderAnimationService,
		),
		app.buildTilesUseCases(types.TilesetTypeHQ, renderAnimationService),
		hqUseCases,
		bonusesRepository,
		app.buildTilesUseCases(
			types.TilesetTypeBonuses,
			renderAnimationService,
		),
		app.textFace,
		rendererTileSize,
		mapOffsetX,
		mapOffsetY,
		mapWidthHeightForAdapter,
		int(app.config.GetTitleFontSize()),
		int(app.config.GetSubtitleFontSize()),
		int(app.config.GetRegularFontSize()),
	)
}
