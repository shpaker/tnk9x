package states

import (
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"

	game "github.com/shpaker/gonflict/internal/adapters/game"
	"github.com/shpaker/gonflict/internal/adapters/game/input_adapters"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/services"
	collision_services "github.com/shpaker/gonflict/internal/services/collision_services"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
	tank_use_cases "github.com/shpaker/gonflict/internal/use_cases/tank_use_cases"
)

// GameStateBuilder создает все компоненты GameState
type GameStateBuilder struct {
	// Репозитории
	mapsRepository    interfaces.IMapsDataRepository
	scriptsRepository interfaces.IScriptsRepository
	gameRepository    interfaces.IGameRepositoriesRegistry
	tilesetRegistry   interfaces.ITilesetRepositoryRegistry

	// Конфигурация
	config      interfaces.IConfigProvider
	levelNumber int

	// Карта уровня
	mapEntity *types.MapEntity

	// Сервисы
	boundaryCollisionService interfaces.IBoundaryCollisionService
	wallCollisionService     interfaces.IWallCollisionService
	coordinateService        interfaces.ICoordinateService
	tankBrakingService       interfaces.ITankBrakingService

	// Временные адаптеры
	tempRendererAdapter *game.GameRendererAdapter
	tempInputAdapter    interfaces.IInputAdapter

	// Lua Engine для AI (существует весь срок жизни App)
	luaEngine interfaces.ILuaEngine

	// Сессия
	session *types.GameSessionEntity
}

// NewGameStateBuilder создает новый builder
func NewGameStateBuilder(
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
	tempRendererAdapter *game.GameRendererAdapter,
	tempInputAdapter interfaces.IInputAdapter,
	luaEngine interfaces.ILuaEngine,
	session *types.GameSessionEntity,
) *GameStateBuilder {
	return &GameStateBuilder{
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
		tempRendererAdapter:      tempRendererAdapter,
		tempInputAdapter:         tempInputAdapter,
		luaEngine:                luaEngine,
		session:                  session,
	}
}

// Build создает и возвращает новый экземпляр GameState
func (b *GameStateBuilder) Build() (*GameState, error) {
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
	tankRenderUseCases := tank_use_cases.NewTankRenderUseCases()
	tankCommonUseCases := tank_use_cases.NewTankCommonUseCases(
		bulletUseCases,
		tilesUseCasesWithAnimations,
		b.tankBrakingService,
		b.coordinateService,
		tankRenderUseCases,
		b.gameRepository.TanksRepository(),
	)
	entitiesCollisionService := collision_services.NewEntitiesCollisionService()

	enemyRespawnDelay := b.config.GetEnemyRespawnDelayTicks()

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
		mapUseCases,
	)

	// Создаем HQ TilesUseCases для работы с HQ tileset и анимациями взрыва
	hqTileService := services.NewTileServiceWithSpecialRepos(
		b.tilesetRegistry.Player(),
		b.tilesetRegistry.Spawner(),
		b.tilesetRegistry.Explosion(),
	)
	hqAnimationService := services.NewAnimationService()
	hqTilesUseCases := use_cases.NewTilesUseCasesWithAnimations(
		b.tilesetRegistry.HQ(),
		b.gameRepository.AnimationsRepository(),
		b.tilesetRegistry.Spawner(),
		b.tilesetRegistry.Explosion(), // Необходимо для CreateExplosionAnimation
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
	playerSpawners := b.config.GetPlayerSpawners()
	playerSpawner := types.Position{X: 12, Y: 24}
	if len(playerSpawners) > 0 {
		playerSpawner = playerSpawners[0]
	}
	baseSize := types.Size{Width: int(baseSizePx), Height: int(baseSizePx)}

	tankLifecycleUseCases.SetSpawnConfiguration(
		b.gameRepository.TanksRepository(),
		enemySpawners,
		playerSpawner,
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

	// Собираем финальный GameState
	gameState := b.buildGameState(
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

	return gameState, nil
}

// loadLevel загружает уровень и заполняет репозиторий блоков
func (b *GameStateBuilder) loadLevel() error {
	tileBaseSize := int(b.config.GetTileBaseSize())
	mapEntity, err := b.mapsRepository.GetLevel(b.levelNumber, tileBaseSize)
	if err != nil {
		return err
	}

	b.mapEntity = mapEntity
	return nil
}

// buildTileServices создает сервисы для тайлов и анимаций
func (b *GameStateBuilder) buildTileServices() (*use_cases.TilesUseCases, error) {
	tileService := services.NewTileServiceWithSpecialRepos(
		b.tilesetRegistry.Player(),
		b.tilesetRegistry.Spawner(),
		b.tilesetRegistry.Explosion(),
	)
	animationService := services.NewAnimationService()

	tilesUseCasesWithAnimations := use_cases.NewTilesUseCasesWithAnimations(
		b.tilesetRegistry.Player(),
		b.gameRepository.AnimationsRepository(),
		b.tilesetRegistry.Spawner(),
		b.tilesetRegistry.Explosion(),
		tileService,
		animationService,
	)

	return tilesUseCasesWithAnimations, nil
}

// buildBulletUseCases создает Use Cases для пуль
func (b *GameStateBuilder) buildBulletUseCases() (*use_cases.BulletUseCases, uint, error) {
	baseSizePx := b.config.GetBaseSizePx()

	bulletTileService := services.NewTileService(b.tilesetRegistry.Bullet())
	bulletAnimationService := services.NewAnimationService()
	bulletTilesUseCases := use_cases.NewTilesUseCases(
		b.tilesetRegistry.Bullet(),
		bulletTileService,
		bulletAnimationService,
	)
	bulletUseCases := use_cases.NewBulletUseCases(
		b.gameRepository.BulletsRepository(),
		bulletTilesUseCases,
		baseSizePx,
	)

	return bulletUseCases, baseSizePx, nil
}

// createHQ создает базу из конфига
func (b *GameStateBuilder) createHQ(
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
func (b *GameStateBuilder) buildCollisionServices(
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

// buildGameState собирает финальный GameState
func (b *GameStateBuilder) buildGameState(
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
) *GameState {
	return &GameState{
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
		InputAdapter:          b.tempInputAdapter,
		RendererAdapter:       b.tempRendererAdapter,
		EnemyInputAdapter:     enemyInputAdapter,
		StartTime:             time.Now(),
		Session:               b.session,
	}
}
