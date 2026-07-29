package states

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/tnk9x/internal/adapters/stage"
	"github.com/shpaker/tnk9x/internal/adapters/stage/input_adapters"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/repositories/processed"
	"github.com/shpaker/tnk9x/internal/services"
	collision_services "github.com/shpaker/tnk9x/internal/services/collision_services"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
	state_use_cases "github.com/shpaker/tnk9x/internal/use_cases/state_use_cases"
	tank_use_cases "github.com/shpaker/tnk9x/internal/use_cases/tank_use_cases"

	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

type StageStateBuilder struct {
	mapsRepository    interfaces.IMapsDataRepository
	scriptsRepository interfaces.IScriptsRepository
	gameRepository    interfaces.IGameRepositoriesRegistry
	tilesetRegistry   interfaces.ITilesetRepositoryRegistry
	fileRepository    interfaces.IFileRepository

	config      interfaces.IConfigProvider
	levelNumber int

	textFace text.Face

	mapEntity *types.MapEntity

	boundaryCollisionService interfaces.IBoundaryCollisionService
	wallCollisionService     interfaces.IWallCollisionService
	tankBrakingService       interfaces.ITankBrakingService

	scriptEngine interfaces.IAIScriptEngine

	session      *session_entities.GameSessionEntity
	audioContext *audio.Context
}

func NewStageStateBuilder(
	mapsRepository interfaces.IMapsDataRepository,
	scriptsRepository interfaces.IScriptsRepository,
	levelNumber int,
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	config interfaces.IConfigProvider,
	gameRepository interfaces.IGameRepositoriesRegistry,
	boundaryCollisionService interfaces.IBoundaryCollisionService,
	wallCollisionService interfaces.IWallCollisionService,
	tankBrakingService interfaces.ITankBrakingService,
	textFace text.Face,
	scriptEngine interfaces.IAIScriptEngine,
	session *session_entities.GameSessionEntity,
	fileRepository interfaces.IFileRepository,
	audioContext *audio.Context,
) *StageStateBuilder {
	return &StageStateBuilder{
		mapsRepository:           mapsRepository,
		scriptsRepository:        scriptsRepository,
		gameRepository:           gameRepository,
		tilesetRegistry:          tilesetRegistry,
		fileRepository:           fileRepository,
		config:                   config,
		levelNumber:              levelNumber,
		boundaryCollisionService: boundaryCollisionService,
		wallCollisionService:     wallCollisionService,
		tankBrakingService:       tankBrakingService,
		textFace:                 textFace,
		scriptEngine:             scriptEngine,
		session:                  session,
		audioContext:             audioContext,
	}
}

func (b *StageStateBuilder) Build() (*StageState, error) {
	if err := b.loadLevel(); err != nil {
		return nil, err
	}

	// Создаем звуковой адаптер
	soundsRepository := processed.NewSoundsRepository(b.fileRepository)
	soundAdapter, err := stage.NewSoundAdapter(
		soundsRepository,
		b.audioContext,
		b.config.GetVolume(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sound adapter: %w", err)
	}

	// Создаем SoundUseCases для использования в других use cases
	soundUseCases := use_cases.NewSoundUseCases(
		b.gameRepository.GetSoundEventsRepository(),
	)

	// Создаем SpecsUseCases для управления характеристиками танков
	specsUseCases := use_cases.NewSpecsUseCases()

	tilesUseCasesWithAnimations, err := b.buildTileServices()
	if err != nil {
		return nil, err
	}

	bulletUseCases, baseSizePx, err := b.buildBulletUseCases()
	if err != nil {
		return nil, err
	}

	entitiesCollisionService := collision_services.NewEntitiesCollisionService()
	spawnCollisionService := collision_services.NewSpawnCollisionService(
		entitiesCollisionService,
	)

	enemyRespawnDelay := b.config.GetEnemyRespawnDelayTicks()

	renderUseCases := use_cases.NewRenderUseCases(tilesUseCasesWithAnimations)

	tankCommonUseCases := tank_use_cases.NewTankCommonUseCases(
		b.tankBrakingService,
		renderUseCases,
		b.gameRepository.GetTanksRepository(),
		specsUseCases,
	)

	spawnLayout := types.SpawnLayout{
		EnemySpawners:  b.config.GetEnemySpawners(),
		Player1Spawner: b.config.GetPlayer1Spawn(),
		Player2Spawner: b.config.GetPlayer2Spawn(),
		BaseSize: types.Size{
			Width:  int(baseSizePx),
			Height: int(baseSizePx),
		},
	}

	tankLifecycleUseCases := tank_use_cases.NewTankLifecycleUseCases(
		tilesUseCasesWithAnimations,
		renderUseCases,
		tankCommonUseCases,
		b.gameRepository.GetTanksRepository(),
		spawnCollisionService,
		specsUseCases,
		spawnLayout,
	)

	mapUseCases := use_cases.NewMapUseCases(b.mapEntity)

	tankActionsUseCases := tank_use_cases.NewTankActionsUseCases(
		b.tankBrakingService,
		bulletUseCases,
		tankCommonUseCases,
		renderUseCases,
		mapUseCases,
		soundUseCases,
	)

	hqTileService := services.NewTileServiceWithSpecialRepos(
		b.tilesetRegistry,
		types.TilesetTypePlayer,
		types.TilesetType(""),
		types.TilesetTypeSpawner,
		types.TilesetTypeExplosion,
	)
	hqAnimationService := services.NewAnimationService()
	hqTilesUseCases := use_cases.NewTilesUseCasesWithAnimations(
		b.tilesetRegistry,
		types.TilesetTypeHQ,
		b.gameRepository.GetAnimationsRepository(),
		types.TilesetTypeSpawner,
		types.TilesetTypeExplosion,
		hqTileService,
		hqAnimationService,
	)

	hq := b.createHQ(hqTilesUseCases, baseSizePx)

	if b.mapEntity == nil {
		return nil, fmt.Errorf("map entity is nil")
	}
	sizePx := b.mapEntity.GetSizePx()
	tileBaseSize := int(b.config.GetTileBaseSize())
	mapBlocksWidth := sizePx.Width / tileBaseSize
	mapBlocksHeight := sizePx.Height / tileBaseSize

	b.scriptEngine.SetGlobalNumber(
		"MAP_X_BLOCKS_COUNT",
		float64(mapBlocksWidth),
	)
	b.scriptEngine.SetGlobalNumber(
		"MAP_Y_BLOCKS_COUNT",
		float64(mapBlocksHeight),
	)
	b.scriptEngine.SetGlobalNumber("MAP_WIDTH_PX", float64(sizePx.Width))
	b.scriptEngine.SetGlobalNumber("MAP_HEIGHT_PX", float64(sizePx.Height))
	b.scriptEngine.SetGlobalNumber("TANK_SIZE_PX", float64(baseSizePx))
	b.scriptEngine.SetGlobalNumber("BLOCK_SIZE_PX", float64(baseSizePx/2))

	enemyScript, err := b.scriptsRepository.GetScript("enemies")
	if err != nil {
		return nil, err
	}

	if err := b.scriptEngine.LoadScript(enemyScript); err != nil {
		return nil, err
	}

	updateInterval := 60
	if b.config.GetAIUpdateIntervalTicks() > 0 {
		updateInterval = b.config.GetAIUpdateIntervalTicks()
	}

	aiUseCases := use_cases.NewAIUseCases(b.scriptEngine)
	enemyInputAdapter, err := input_adapters.NewAiInputAdapter(
		tankActionsUseCases,
		nil,
		updateInterval,
		aiUseCases,
	)
	if err != nil {
		return nil, err
	}

	var stageSession *session_entities.StageSessionEntity
	if b.session != nil {
		stageSession = b.session.StageSession()
	}

	bonusesRepository := b.gameRepository.GetBonusesRepository()

	// Создаем BonusUseCases для использования в StageUseCases и buildCollisionServices
	bonusTilesUseCasesForBonus, err := b.buildBonusTilesUseCases()
	if err != nil {
		return nil, fmt.Errorf("failed to build bonus tiles use cases: %w", err)
	}
	bonusUseCases := use_cases.NewBonusUseCases(
		tankCommonUseCases,
		tankLifecycleUseCases,
		stageSession,
		bonusesRepository,
		b.config,
		bonusTilesUseCasesForBonus,
		renderUseCases,
		soundUseCases,
	)

	collisionUseCases, hqUseCases, err := b.buildCollisionServices(
		bulletUseCases,
		tankActionsUseCases,
		mapUseCases,
		tankCommonUseCases,
		tankLifecycleUseCases,
		hqTilesUseCases,
		hq,
		entitiesCollisionService,
		spawnCollisionService,
		bonusUseCases,
		soundUseCases,
	)
	if err != nil {
		return nil, err
	}

	stageUseCases := state_use_cases.NewStageUseCases(
		tankLifecycleUseCases,
		tankCommonUseCases,
		bulletUseCases,
		collisionUseCases,
		hqUseCases,
		stageSession,
		enemyRespawnDelay,
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

	renderAnimationService := services.NewAnimationService()
	mapTilesUseCases := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		types.TilesetTypeBlocks,
		services.NewTileService(b.tilesetRegistry, types.TilesetTypeBlocks),
		renderAnimationService,
	)
	tankTilesUseCases := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		types.TilesetTypePlayer,
		services.NewTileService(b.tilesetRegistry, types.TilesetTypePlayer),
		renderAnimationService,
	)
	bulletTilesUseCases := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		types.TilesetTypeBullet,
		services.NewTileService(b.tilesetRegistry, types.TilesetTypeBullet),
		renderAnimationService,
	)
	spawnerTilesUseCases := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		types.TilesetTypeSpawner,
		services.NewTileService(
			b.tilesetRegistry,
			types.TilesetTypeSpawner,
		),
		renderAnimationService,
	)
	explosionTilesUseCases := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		types.TilesetTypeExplosion,
		services.NewTileService(
			b.tilesetRegistry,
			types.TilesetTypeExplosion,
		),
		renderAnimationService,
	)
	hqTilesUseCasesForRenderer := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		types.TilesetTypeHQ,
		services.NewTileService(b.tilesetRegistry, types.TilesetTypeHQ),
		renderAnimationService,
	)
	bonusTilesUseCasesForRenderer := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		types.TilesetTypeBonuses,
		services.NewTileService(
			b.tilesetRegistry,
			types.TilesetTypeBonuses,
		),
		renderAnimationService,
	)

	mapBlocksCount := b.config.GetMapBlocksCount()
	rendererTileSize := int(b.config.GetTileBaseSize())
	mapWidthHeightForAdapter := mapBlocksCount.Width * rendererTileSize

	mapOffsetY := 16
	mapOffsetX := mapWidthHeightForAdapter / 2

	rendererAdapter := stage.NewStageRendererAdapter(
		mapUseCases,
		tankCommonUseCases,
		bulletUseCases,
		mapTilesUseCases,
		tankTilesUseCases,
		bulletTilesUseCases,
		spawnerTilesUseCases,
		explosionTilesUseCases,
		hqTilesUseCasesForRenderer,
		hqUseCases,
		b.gameRepository.GetBonusesRepository(),
		bonusTilesUseCasesForRenderer,
		b.textFace,
		rendererTileSize,
		mapOffsetX,
		mapOffsetY,
		mapWidthHeightForAdapter,
		int(b.config.GetTitleFontSize()),
		int(b.config.GetSubtitleFontSize()),
		int(b.config.GetRegularFontSize()),
	)

	stageState := b.buildStageState(
		hq,
		hqUseCases,
		tankActionsUseCases,
		tankCommonUseCases,
		renderUseCases,
		tankLifecycleUseCases,
		bulletUseCases,
		mapUseCases,
		collisionUseCases,
		tilesUseCasesWithAnimations,
		enemyInputAdapter,
	)

	// Звуковой адаптер уже создан выше, используем его

	stageState.inputAdapters = make([]interfaces.IInputAdapter, 2)
	stageState.inputAdapters[types.PlayerTankNumPlayer1] = inputAdapter1
	stageState.inputAdapters[types.PlayerTankNumPlayer2] = inputAdapter2
	stageState.RendererAdapter = rendererAdapter
	stageState.stageUseCases = stageUseCases
	stageState.bonusesRepository = b.gameRepository.GetBonusesRepository()
	stageState.soundUseCases = soundUseCases
	stageState.soundPlayerAdapter = soundAdapter
	stageState.debugEnabled = false // По умолчанию выключен, будет установлен из app.go

	return stageState, nil
}

func (b *StageStateBuilder) loadLevel() error {
	tileBaseSize := int(b.config.GetTileBaseSize())
	mapEntity, err := b.mapsRepository.GetLevel(b.levelNumber, tileBaseSize)
	if err != nil {
		return err
	}

	b.mapEntity = mapEntity
	return nil
}

func (b *StageStateBuilder) buildTileServices() (*use_cases.TilesUseCases, error) {
	tileService := services.NewTileServiceWithSpecialRepos(
		b.tilesetRegistry,
		types.TilesetTypePlayer,
		types.TilesetTypeEnemy,
		types.TilesetTypeSpawner,
		types.TilesetTypeExplosion,
	)
	animationService := services.NewAnimationService()

	tilesUseCasesWithAnimations := use_cases.NewTilesUseCasesWithAnimations(
		b.tilesetRegistry,
		types.TilesetTypePlayer,
		b.gameRepository.GetAnimationsRepository(),
		types.TilesetTypeSpawner,
		types.TilesetTypeExplosion,
		tileService,
		animationService,
	)

	return tilesUseCasesWithAnimations, nil
}

func (b *StageStateBuilder) buildBulletUseCases() (*use_cases.BulletUseCases, uint, error) {
	baseSizePx := b.config.GetBaseSizePx()

	bulletTileService := services.NewTileService(
		b.tilesetRegistry,
		types.TilesetTypeBullet,
	)
	bulletAnimationService := services.NewAnimationService()
	bulletTilesUseCases := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		types.TilesetTypeBullet,
		bulletTileService,
		bulletAnimationService,
	)
	bulletUseCases := use_cases.NewBulletUseCases(
		b.gameRepository.GetBulletsRepository(),
		bulletTilesUseCases,
		baseSizePx,
	)

	return bulletUseCases, baseSizePx, nil
}

func (b *StageStateBuilder) buildBonusTilesUseCases() (*use_cases.TilesUseCases, error) {
	bonusTileService := services.NewTileService(
		b.tilesetRegistry,
		types.TilesetTypeBonuses,
	)
	bonusAnimationService := services.NewAnimationService()
	bonusTilesUseCases := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		types.TilesetTypeBonuses,
		bonusTileService,
		bonusAnimationService,
	)

	return bonusTilesUseCases, nil
}

func (b *StageStateBuilder) createHQ(
	tilesUseCases *use_cases.TilesUseCases,
	baseSizePx uint,
) *types.HQEntity {
	hqPos := b.config.GetHQPosition()
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

func (b *StageStateBuilder) buildCollisionServices(
	bulletUseCases *use_cases.BulletUseCases,
	playerTankActions interfaces.ITankActionsUseCases,
	mapUseCases *use_cases.MapUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	hqTilesUseCases *use_cases.TilesUseCases,
	hq *types.HQEntity,
	entitiesCollisionService interfaces.IEntitiesCollisionService,
	spawnCollisionService interfaces.ISpawnCollisionService,
	bonusUseCases *use_cases.BonusUseCases,
	soundUseCases *use_cases.SoundUseCases,
) (*use_cases.CollisionUseCases, interfaces.IHQUseCases, error) {
	bulletCollisionService := collision_services.NewBulletCollisionService(
		int(b.config.GetTileBaseSize()),
		entitiesCollisionService,
	)

	var hqUseCases interfaces.IHQUseCases
	if hq != nil {
		hqUseCases = use_cases.NewHQUseCases(
			hqTilesUseCases,
			hq,
		)
	}

	bonusesRepository := b.gameRepository.GetBonusesRepository()
	// BonusUseCases уже создан выше и передан как параметр

	collisionUseCases := use_cases.NewCollisionUseCases(
		bulletUseCases,
		playerTankActions,
		mapUseCases,
		tankCommonUseCases,
		tankLifecycleUseCases,
		b.boundaryCollisionService,
		b.wallCollisionService,
		bulletCollisionService,
		entitiesCollisionService,
		spawnCollisionService,
		hqUseCases,
		bonusUseCases,
		bonusesRepository,
		soundUseCases,
	)

	return collisionUseCases, hqUseCases, nil
}

func (b *StageStateBuilder) buildStageState(
	hq *types.HQEntity,
	hqUseCases interfaces.IHQUseCases,
	tankActionsUseCases interfaces.ITankActionsUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	renderUseCases interfaces.IRenderUseCases,
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	bulletUseCases *use_cases.BulletUseCases,
	mapUseCases *use_cases.MapUseCases,
	collisionUseCases *use_cases.CollisionUseCases,
	tilesUseCasesWithAnimations *use_cases.TilesUseCases,
	enemyInputAdapter interfaces.IAiInputAdapter,
) *StageState {
	return &StageState{
		HQEntity:              hq,
		HQUseCases:            hqUseCases,
		TankActionsUseCases:   tankActionsUseCases,
		TankCommonUseCases:    tankCommonUseCases,
		RenderUseCases:        renderUseCases,
		TankLifecycleUseCases: tankLifecycleUseCases,
		BulletUseCases:        bulletUseCases,
		MapUseCases:           mapUseCases,
		CollisionUseCases:     collisionUseCases,
		TilesUseCases:         tilesUseCasesWithAnimations,
		inputAdapters:         make([]interfaces.IInputAdapter, 2),
		RendererAdapter:       nil,
		EnemyInputAdapter:     enemyInputAdapter,
		StartTime:             time.Now(),
		stageUseCases:         nil,
		session:               b.session,
	}
}
