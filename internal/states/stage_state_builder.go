package states

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	lua "github.com/yuin/gopher-lua"

	game "github.com/shpaker/gonflict/internal/adapters/stage"
	"github.com/shpaker/gonflict/internal/adapters/stage/input_adapters"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/services"
	collision_services "github.com/shpaker/gonflict/internal/services/collision_services"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
	stateusecases "github.com/shpaker/gonflict/internal/use_cases/state_use_cases"
	tank_use_cases "github.com/shpaker/gonflict/internal/use_cases/tank_use_cases"

	"github.com/shpaker/gonflict/internal/types/session_entities"
)

// StageStateBuilder создает все компоненты StageState
type StageStateBuilder struct {
	// Репозитории
	mapsRepository    interfaces.IMapsDataRepository
	scriptsRepository interfaces.IScriptsRepository
	gameRepository    interfaces.IGameRepositoriesRegistry
	tilesetRegistry   interfaces.ITilesetRepositoryRegistry

	// Конфигурация
	config      interfaces.IConfigProvider
	levelNumber int

	// Шрифты
	fontUseCases interfaces.IFontUseCases

	// Карта уровня
	mapEntity *types.MapEntity

	// Сервисы
	boundaryCollisionService interfaces.IBoundaryCollisionService
	wallCollisionService     interfaces.IWallCollisionService
	coordinateService        interfaces.ICoordinateService
	tankBrakingService       interfaces.ITankBrakingService

	// Lua Engine для AI (существует весь срок жизни App)
	luaEngine interfaces.ILuaEngine

	// Сессия
	session *session_entities.GameSessionEntity
}

// NewStageStateBuilder создает новый builder
func NewStageStateBuilder(
	mapsRepository interfaces.IMapsDataRepository,
	scriptsRepository interfaces.IScriptsRepository,
	levelNumber int,
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	config interfaces.IConfigProvider,
	gameRepository interfaces.IGameRepositoriesRegistry,
	boundaryCollisionService interfaces.IBoundaryCollisionService,
	wallCollisionService interfaces.IWallCollisionService,
	coordinateService interfaces.ICoordinateService,
	tankBrakingService interfaces.ITankBrakingService,
	fontUseCases interfaces.IFontUseCases,
	luaEngine interfaces.ILuaEngine,
	session *session_entities.GameSessionEntity,
) *StageStateBuilder {
	return &StageStateBuilder{
		mapsRepository:           mapsRepository,
		scriptsRepository:        scriptsRepository,
		gameRepository:           gameRepository,
		tilesetRegistry:          tilesetRegistry,
		config:                   config,
		levelNumber:              levelNumber,
		boundaryCollisionService: boundaryCollisionService,
		wallCollisionService:     wallCollisionService,
		coordinateService:        coordinateService,
		tankBrakingService:       tankBrakingService,
		fontUseCases:             fontUseCases,
		luaEngine:                luaEngine,
		session:                  session,
	}
}

// Build создает и возвращает новый экземпляр StageState
func (b *StageStateBuilder) Build() (*StageState, error) {
	// Загружаем уровень
	if err := b.loadLevel(); err != nil {
		return nil, err
	}

	// Создаем сервисы для тайлов
	tilesUseCasesWithAnimations, err := b.buildTileServices()
	if err != nil {
		return nil, err
	}

	// Создаем Use Cases для пуль
	bulletUseCases, baseSizePx, err := b.buildBulletUseCases()
	if err != nil {
		return nil, err
	}

	// Создаем общие use cases для всех танков
	entitiesCollisionService := collision_services.NewEntitiesCollisionService()

	enemyRespawnDelay := b.config.GetEnemyRespawnDelayTicks()

	// Создаем финальный tankRenderUseCases
	tankRenderUseCases := tank_use_cases.NewTankRenderUseCases(
		tilesUseCasesWithAnimations,
	)

	// Создаем финальный tankCommonUseCases с финальным tankRenderUseCases
	tankCommonUseCases := tank_use_cases.NewTankCommonUseCases(
		bulletUseCases,
		b.tankBrakingService,
		b.coordinateService,
		tankRenderUseCases,
		b.gameRepository.GetTanksRepository(),
	)

	// Создаем финальный tankLifecycleUseCases с финальными зависимостями
	tankLifecycleUseCases := tank_use_cases.NewTankLifecycleUseCases(
		tilesUseCasesWithAnimations,
		tankRenderUseCases,
		tankCommonUseCases,
		enemyRespawnDelay,
	)

	// Создаем Use Cases для карты
	mapUseCases := use_cases.NewMapUseCases(
		b.mapEntity,
	)

	tankActionsUseCases := tank_use_cases.NewTankActionsUseCases(
		b.tankBrakingService,
		b.coordinateService,
		bulletUseCases,
		tankCommonUseCases,
		tankRenderUseCases,
		mapUseCases,
	)

	// Создаем HQ TilesUseCases для работы с HQ tileset и анимациями взрыва
	hqTileService := services.NewTileServiceWithSpecialRepos(
		b.tilesetRegistry,
		processed.TilesetTypePlayer,
		processed.TilesetType(""),
		processed.TilesetTypeSpawner,
		processed.TilesetTypeExplosion,
	)
	hqAnimationService := services.NewAnimationService()
	hqTilesUseCases := use_cases.NewTilesUseCasesWithAnimations(
		b.tilesetRegistry,
		processed.TilesetTypeHQ,
		b.gameRepository.GetAnimationsRepository(),
		processed.TilesetTypeSpawner,
		processed.TilesetTypeExplosion,
		hqTileService,
		hqAnimationService,
	)

	// Создаем базу
	hq := b.createHQ(hqTilesUseCases, baseSizePx)

	// Получаем размеры карты (в блоках) из MapEntity
	if b.mapEntity == nil {
		return nil, fmt.Errorf("map entity is nil")
	}
	sizePx := b.mapEntity.GetSizePx()
	tileBaseSize := int(b.config.GetTileBaseSize())
	mapBlocksWidth := sizePx.Width / tileBaseSize
	mapBlocksHeight := sizePx.Height / tileBaseSize

	// Устанавливаем глобальные переменные размера карты и базового размера в Lua
	b.luaEngine.SetGlobal("MAP_X_BLOCKS_COUNT", lua.LNumber(mapBlocksWidth))
	b.luaEngine.SetGlobal("MAP_Y_BLOCKS_COUNT", lua.LNumber(mapBlocksHeight))
	b.luaEngine.SetGlobal("TANK_SIZE_PX", lua.LNumber(baseSizePx))
	b.luaEngine.SetGlobal("BLOCK_SIZE_PX", lua.LNumber(baseSizePx/2))

	// Загружаем скрипт AI для врагов
	enemyScript, err := b.scriptsRepository.GetScript("enemies")
	if err != nil {
		return nil, err
	}

	// Выполняем скрипт AI в общем Lua engine (существует весь срок жизни App)
	if err := b.luaEngine.Execute(enemyScript); err != nil {
		return nil, err
	}

	updateInterval := 60
	if b.config.GetAIUpdateIntervalTicks() > 0 {
		updateInterval = b.config.GetAIUpdateIntervalTicks()
	}

	typeConverter := services.NewAITypeConverter(b.luaEngine)
	aiUseCases := use_cases.NewAIUseCases(b.luaEngine, typeConverter)
	enemyInputAdapter, err := input_adapters.NewAiInputAdapter(
		tankActionsUseCases,
		nil,
		updateInterval,
		aiUseCases,
	)
	if err != nil {
		return nil, err
	}

	enemySpawners := b.config.GetEnemySpawners()
	player1Spawner := b.config.GetPlayer1Spawn()
	player2Spawner := b.config.GetPlayer2Spawn()
	baseSize := types.Size{Width: int(baseSizePx), Height: int(baseSizePx)}

	tankLifecycleUseCases.SetSpawnConfiguration(
		b.gameRepository.GetTanksRepository(),
		enemySpawners,
		player1Spawner,
		player2Spawner,
		baseSize,
	)

	// Создаем сервисы коллизий
	collisionUseCases, hqUseCases, err := b.buildCollisionServices(
		bulletUseCases,
		tankActionsUseCases,
		mapUseCases,
		tankCommonUseCases,
		tankLifecycleUseCases,
		hqTilesUseCases,
		hq,
		entitiesCollisionService,
	)
	if err != nil {
		return nil, err
	}

	tankLifecycleUseCases.SetCollisionUseCases(collisionUseCases)

	var stageSession *session_entities.StageSessionEntity
	if b.session != nil {
		stageSession = b.session.StageSession()
	}

	stageUseCasesValue := stateusecases.NewStageUseCases(
		tankLifecycleUseCases,
		tankCommonUseCases,
		bulletUseCases,
		collisionUseCases,
		hqUseCases,
		stageSession,
		enemyRespawnDelay,
	)
	stageUseCases := &stageUseCasesValue

	// Создаем адаптер ввода игрока 1
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

	// Создаем адаптер ввода игрока 2 (стрелки)
	inputAdapter2 := input_adapters.NewStageKeyboardInputAdapter(
		tankActionsUseCases,
		nil,
		stageUseCases,
		ebiten.KeyArrowUp,
		ebiten.KeyArrowDown,
		ebiten.KeyArrowLeft,
		ebiten.KeyArrowRight,
		ebiten.KeyEnter,
		ebiten.KeyP, // Пауза общая для обоих игроков
	)

	// Создаем TilesUseCases для рендера
	renderAnimationService := services.NewAnimationService()
	mapTilesUseCases := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		processed.TilesetTypeBlocks,
		services.NewTileService(b.tilesetRegistry, processed.TilesetTypeBlocks),
		renderAnimationService,
	)
	tankTilesUseCases := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		processed.TilesetTypePlayer,
		services.NewTileService(b.tilesetRegistry, processed.TilesetTypePlayer),
		renderAnimationService,
	)
	bulletTilesUseCases := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		processed.TilesetTypeBullet,
		services.NewTileService(b.tilesetRegistry, processed.TilesetTypeBullet),
		renderAnimationService,
	)
	spawnerTilesUseCases := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		processed.TilesetTypeSpawner,
		services.NewTileService(
			b.tilesetRegistry,
			processed.TilesetTypeSpawner,
		),
		renderAnimationService,
	)
	explosionTilesUseCases := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		processed.TilesetTypeExplosion,
		services.NewTileService(
			b.tilesetRegistry,
			processed.TilesetTypeExplosion,
		),
		renderAnimationService,
	)
	hqTilesUseCasesForRenderer := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		processed.TilesetTypeHQ,
		services.NewTileService(b.tilesetRegistry, processed.TilesetTypeHQ),
		renderAnimationService,
	)

	mapOffsets := b.config.GetMapOffsets()
	mapOffsetX := int(mapOffsets[0])
	mapOffsetY := int(mapOffsets[1])
	mapBlocksCount := b.config.GetMapBlocksCount()
	mapWidthHeightForAdapter := mapBlocksCount.Width * int(
		b.config.GetTileBaseSize(),
	)

	rendererAdapter := game.NewStageRendererAdapter(
		mapUseCases,
		tankCommonUseCases,
		tankRenderUseCases,
		bulletUseCases,
		mapTilesUseCases,
		tankTilesUseCases,
		bulletTilesUseCases,
		spawnerTilesUseCases,
		explosionTilesUseCases,
		hqTilesUseCasesForRenderer,
		hqUseCases,
		b.fontUseCases,
		int(b.config.GetTileBaseSize()),
		mapOffsetX,
		mapOffsetY,
		mapWidthHeightForAdapter,
		int(b.config.GetTitleFontSize()),
		int(b.config.GetSubtitleFontSize()),
		int(b.config.GetRegularFontSize()),
	)

	// Собираем финальный StageState
	stageState := b.buildStageState(
		hq,
		hqUseCases,
		tankActionsUseCases,
		tankCommonUseCases,
		tankRenderUseCases,
		tankLifecycleUseCases,
		bulletUseCases,
		mapUseCases,
		collisionUseCases,
		tilesUseCasesWithAnimations,
		enemyInputAdapter,
	)

	// Инициализируем массив адаптеров
	stageState.inputAdapters = make([]interfaces.IInputAdapter, 2)
	stageState.inputAdapters[types.PlayerTankNumPlayer1] = inputAdapter1
	stageState.inputAdapters[types.PlayerTankNumPlayer2] = inputAdapter2 // Может быть nil, если игра на одного игрока
	stageState.RendererAdapter = rendererAdapter
	stageState.stageUseCases = stageUseCases

	return stageState, nil
}

// loadLevel загружает уровень и заполняет репозиторий блоков
func (b *StageStateBuilder) loadLevel() error {
	tileBaseSize := int(b.config.GetTileBaseSize())
	mapEntity, err := b.mapsRepository.GetLevel(b.levelNumber, tileBaseSize)
	if err != nil {
		return err
	}

	b.mapEntity = mapEntity
	return nil
}

// buildTileServices создает сервисы для тайлов и анимаций
func (b *StageStateBuilder) buildTileServices() (*use_cases.TilesUseCases, error) {
	tileService := services.NewTileServiceWithSpecialRepos(
		b.tilesetRegistry,
		processed.TilesetTypePlayer,
		processed.TilesetTypeEnemy,
		processed.TilesetTypeSpawner,
		processed.TilesetTypeExplosion,
	)
	animationService := services.NewAnimationService()

	tilesUseCasesWithAnimations := use_cases.NewTilesUseCasesWithAnimations(
		b.tilesetRegistry,
		processed.TilesetTypePlayer,
		b.gameRepository.GetAnimationsRepository(),
		processed.TilesetTypeSpawner,
		processed.TilesetTypeExplosion,
		tileService,
		animationService,
	)

	return tilesUseCasesWithAnimations, nil
}

// buildBulletUseCases создает Use Cases для пуль
func (b *StageStateBuilder) buildBulletUseCases() (*use_cases.BulletUseCases, uint, error) {
	baseSizePx := b.config.GetBaseSizePx()

	bulletTileService := services.NewTileService(
		b.tilesetRegistry,
		processed.TilesetTypeBullet,
	)
	bulletAnimationService := services.NewAnimationService()
	bulletTilesUseCases := use_cases.NewTilesUseCases(
		b.tilesetRegistry,
		processed.TilesetTypeBullet,
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

// createHQ создает базу из конфига
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

	// Создаем Image для базы
	var imageGetter types.IImageProvider
	if tilesUseCases != nil {
		// Создаем статический тайл для HQ
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

// buildCollisionServices создает сервисы коллизий
func (b *StageStateBuilder) buildCollisionServices(
	bulletUseCases *use_cases.BulletUseCases,
	playerTankActions interfaces.ITankActionsUseCases,
	mapUseCases *use_cases.MapUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	hqTilesUseCases *use_cases.TilesUseCases,
	hq *types.HQEntity,
	entitiesCollisionService interfaces.IEntitiesCollisionService,
) (*use_cases.CollisionUseCases, interfaces.IHQUseCases, error) {
	// Создаем BulletCollisionService с EntitiesCollisionService
	bulletCollisionService := collision_services.NewBulletCollisionService(
		int(b.config.GetTileBaseSize()),
		entitiesCollisionService,
	)

	// Создаем HQUseCases с hqTilesUseCases
	var hqUseCases interfaces.IHQUseCases
	if hq != nil {
		hqUseCases = use_cases.NewHQUseCases(
			hqTilesUseCases,
			hq,
		)
	}

	// Создаем CollisionUseCases с правильным BulletCollisionService
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
		hqUseCases,
	)

	return collisionUseCases, hqUseCases, nil
}

// buildStageState собирает финальный StageState
func (b *StageStateBuilder) buildStageState(
	hq *types.HQEntity,
	hqUseCases interfaces.IHQUseCases,
	tankActionsUseCases interfaces.ITankActionsUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	tankRenderUseCases interfaces.ITankRenderUseCases,
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
		TankRenderUseCases:    tankRenderUseCases,
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
