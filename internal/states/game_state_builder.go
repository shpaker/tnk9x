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
	session *types.SessionEntity
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
	session *types.SessionEntity,
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
	tankRenderUseCases := use_cases.NewTankRenderUseCases()
	tankLifecycleUseCases := use_cases.NewTankLifecycleUseCases(
		tilesUseCasesWithAnimations,
		tankRenderUseCases,
	)
	tankCommonUseCases := use_cases.NewTankCommonUseCases(
		bulletUseCases,
		tilesUseCasesWithAnimations,
		b.tankBrakingService,
		b.coordinateService,
		tankRenderUseCases,
	)
	// Создаем Use Cases для карты
	mapUseCases := use_cases.NewMapUseCases(
		b.mapEntity,
	)

	tankActionsUseCases := use_cases.NewTankActionsUseCases(
		b.tankBrakingService,
		b.coordinateService,
		bulletUseCases,
		tankCommonUseCases,
		mapUseCases,
	)

	// Создаем компоненты игрока
	playerTank, err := b.buildPlayerComponents(
		tilesUseCasesWithAnimations,
		baseSizePx,
	)
	if err != nil {
		return nil, err
	}

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

	// Создаем AI контекст
	aiContext := b.createAIContext(mapUseCases)

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

	// Создаем врагов
	enemyInputAdapters, enemyTanksEntities, err := b.buildEnemyComponents(
		tankLifecycleUseCases,
		tankActionsUseCases,
		aiContext,
		b.luaEngine,
		baseSizePx,
	)
	if err != nil {
		return nil, err
	}

	// Создаем сервисы коллизий
	collisionUseCases, hqUseCases, err := b.buildCollisionServices(
		bulletUseCases,
		playerTank,
		tankActionsUseCases,
		mapUseCases,
		enemyTanksEntities,
		tankCommonUseCases,
		tankLifecycleUseCases,
		tilesUseCasesWithAnimations,
		hqTilesUseCases,
		hq,
	)
	if err != nil {
		return nil, err
	}

	// Подготовим массив врагов
	enemyTanksArray := b.buildEnemyTanksArray(enemyTanksEntities)

	// Собираем финальный GameState
	gameState := b.buildGameState(
		playerTank,
		enemyTanksArray,
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
		aiContext,
		enemyInputAdapters,
	)

	// Запускаем спавн танка на старте
	gameState.StartTankSpawn()

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

// buildPlayerComponents создает танк игрока
func (b *GameStateBuilder) buildPlayerComponents(
	tilesUseCasesWithAnimations *use_cases.TilesUseCases,
	baseSizePx uint,
) (*types.TankEntity, error) {
	// Берем позицию спавна игрока из конфига
	var playerSpawner types.Position
	playerSpawners := b.config.GetPlayerSpawners()
	if len(playerSpawners) > 0 && len(playerSpawners[0]) == 2 {
		playerSpawner = types.Position{
			X: float64(playerSpawners[0][0]) * float64(baseSizePx),
			Y: float64(playerSpawners[0][1]) * float64(baseSizePx),
		}
	} else {
		// Значение по умолчанию
		playerSpawner = types.Position{X: float64(12 * baseSizePx), Y: float64(24 * baseSizePx)}
	}

	// Создаем танк игрока
	playerTank := &types.TankEntity{
		Position:  playerSpawner,
		Speed:     0,
		Direction: types.DirectionUp,
		State:     types.TankStateSpawning,
		Altitude:  types.SURFACE,
		Size:      types.Size{Width: int(baseSizePx), Height: int(baseSizePx)},
	}
	b.gameRepository.TanksRepository().AddTank(playerTank)

	return playerTank, nil
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

// createAIContext создает AI контекст
func (b *GameStateBuilder) createAIContext(
	mapUseCases *use_cases.MapUseCases,
) *types.GameAiContext {
	blocks := mapUseCases.GetBlocks()
	return &types.GameAiContext{
		Player:  nil, // Будет обновляться в Update
		Enemies: nil, // Будет обновляться в Update
		Bullets: nil, // Будет обновляться в Update
		Blocks:  blocks,
	}
}

// buildEnemyComponents создает все компоненты врагов
func (b *GameStateBuilder) buildEnemyComponents(
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	tankActionsUseCases interfaces.ITankActionsUseCases,
	aiContext *types.GameAiContext,
	luaEngine interfaces.ILuaEngine,
	baseSizePx uint,
) ([]interfaces.IInputAdapter, []*types.TankEntity, error) {
	updateInterval := 60
	if b.config.GetAIUpdateIntervalTicks() > 0 {
		updateInterval = b.config.GetAIUpdateIntervalTicks()
	}

	enemyInputAdapters := make([]interfaces.IInputAdapter, 0, 3)
	enemyTanksEntities := make([]*types.TankEntity, 0, 3)

	for i, spawner := range b.config.GetEnemySpawners() {
		if i >= 3 { // Максимум 3 врага
			break
		}
		if len(spawner) != 2 {
			continue
		}

		// Конвертируем координаты в пиксели
		position := types.Position{
			X: float64(spawner[0]) * float64(baseSizePx),
			Y: float64(spawner[1]) * float64(baseSizePx),
		}

		// Создаем танк врага
		enemyTank := &types.TankEntity{
			Position:  position,
			Speed:     0,
			Direction: types.DirectionDown,
			State:     types.TankStateSpawning,
			Altitude:  types.SURFACE,
			Size: types.Size{
				Width:  int(baseSizePx),
				Height: int(baseSizePx),
			},
		}
		b.gameRepository.TanksRepository().AddTank(enemyTank)

		// Запускаем спавн танка врага через общий lifecycle
		if err := tankLifecycleUseCases.Spawn(enemyTank); err != nil {
			return nil, nil, err
		}

		enemyTanksEntities = append(enemyTanksEntities, enemyTank)

		// Используем общий Lua engine для всех врагов
		// Создаем тип конвертер для AI
		typeConverter := services.NewAITypeConverter(luaEngine)

		// Создаем AI Use Cases
		aiUseCases := use_cases.NewAIUseCases(luaEngine, typeConverter)

		// Создаем AI input адаптер для этого врага
		aiInputAdapter, err := input_adapters.NewAiInputAdapter(
			tankActionsUseCases,
			enemyTank,
			aiContext,
			updateInterval,
			aiUseCases,
		)
		if err != nil {
			return nil, nil, err
		}
		enemyInputAdapters = append(enemyInputAdapters, aiInputAdapter)
	}

	return enemyInputAdapters, enemyTanksEntities, nil
}

// buildCollisionServices создает сервисы коллизий
func (b *GameStateBuilder) buildCollisionServices(
	bulletUseCases *use_cases.BulletUseCases,
	playerTank *types.TankEntity,
	playerTankActions interfaces.ITankActionsUseCases,
	mapUseCases *use_cases.MapUseCases,
	enemyTanksEntities []*types.TankEntity,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	tilesUseCasesWithAnimations *use_cases.TilesUseCases,
	hqTilesUseCases *use_cases.TilesUseCases,
	hq *types.HQEntity,
) (*use_cases.CollisionUseCases, interfaces.IHQUseCases, error) {
	// Создаем сервис проверки коллизий между сущностями
	entitiesCollisionService := collision_services.NewEntitiesCollisionService()

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

// buildEnemyTanksArray создает массив врагов
func (b *GameStateBuilder) buildEnemyTanksArray(
	enemyTanksEntities []*types.TankEntity,
) [3]*types.TankEntity {
	enemyTanksArray := [3]*types.TankEntity{}
	for i := range enemyTanksEntities {
		if i < 3 {
			enemyTanksArray[i] = enemyTanksEntities[i]
		}
	}
	return enemyTanksArray
}

// buildGameState собирает финальный GameState
func (b *GameStateBuilder) buildGameState(
	playerTank *types.TankEntity,
	enemyTanksArray [3]*types.TankEntity,
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
	aiContext *types.GameAiContext,
	enemyInputAdapters []interfaces.IInputAdapter,
) *GameState {
	return &GameState{
		PlayerTank:            playerTank,
		EnemyTanks:            enemyTanksArray,
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
		AIContext:             aiContext,
		InputAdapter:          b.tempInputAdapter,
		RendererAdapter:       b.tempRendererAdapter,
		EnemyInputAdapters:    enemyInputAdapters,
		StartTime:             time.Now(),
		Session:               b.session,
	}
}
