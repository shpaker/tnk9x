package states

import (
	"github.com/shpaker/gonflict/internal/adapters/input_adapters"
	"github.com/shpaker/gonflict/internal/config"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/services"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// GameStateUseCasesFacade - фасад для оркестрации use cases игрового состояния
type GameStateUseCasesFacade struct {
	playerUseCases         interfaces.ITankCommonUseCases
	playerRenderUseCases   interfaces.ITankRenderUseCases    // Для графики игрока
	playerLifecycle        interfaces.ITankLifecycleUseCases // Для Spawn/Explode игрока
	playerTank             *types.TankEntity                 // Танк игрока для передачи в адаптеры
	bulletUseCases         *use_cases.BulletUseCases
	mapUseCases            *use_cases.MapUseCases
	collisionUseCases      *use_cases.CollisionUseCases
	enemyUseCases          []interfaces.ITankCommonUseCases
	enemyRenderUseCases    []interfaces.ITankRenderUseCases    // Для графики врагов
	enemyLifecycles        []interfaces.ITankLifecycleUseCases // Для Spawn/Explode врагов
	enemyTanksEntities     []*types.TankEntity                 // Массив танков врагов для обратной совместимости
	enemyInputAdapters     []interfaces.IInputAdapter          // AI input адаптеры для врагов
	tilesUseCasesWithAnims *use_cases.TilesUseCases            // Общий tilesUseCases для всех анимаций
	aiContext              *types.GameAiContext                // AI контекст для всех адаптеров
}

// NewGameStateUseCasesFacade создает фасад для оркестрации use cases игрового состояния
func NewGameStateUseCasesFacade(
	mapsRepo interfaces.IMapsDataRepository,
	scriptsRepo interfaces.IScriptsRepository,
	levelNumber int,
	mapTilesetRepo interfaces.ITilesetRepository,
	playerTilesetRepo interfaces.ITilesetRepository,
	bulletTilesetRepo interfaces.ITilesetRepository,
	spawnerTilesetRepo interfaces.ITilesetRepository,
	explosionTilesetRepo interfaces.ITilesetRepository,
	gameConfig *config.GameConfig,
	gameRepo interfaces.IGameRepositoriesRegistry,
	boundaryCollisionService interfaces.IBoundaryCollisionService,
	wallCollisionService interfaces.IWallCollisionService,
	coordinateService interfaces.ICoordinateService,
	tankBrakingService interfaces.ITankBrakingService,
) (*GameStateUseCasesFacade, error) {
	// Загружаем уровень
	level, err := mapsRepo.GetLevel(levelNumber)
	if err != nil {
		return nil, err
	}

	// Заполняем репозиторий блоков данными уровня
	for _, block := range level {
		gameRepo.BlocksRepository().AddBlock(block)
	}

	// Создаем сервисы для тайлов и анимаций
	tileService := services.NewTileServiceWithSpecialRepos(
		playerTilesetRepo,
		spawnerTilesetRepo,
		explosionTilesetRepo,
	)
	animationService := services.NewAnimationService()

	// Создаем Use Cases
	tilesUseCasesWithAnimations := use_cases.NewTilesUseCasesWithAnimations(
		playerTilesetRepo,
		gameRepo.AnimationsRepository(),
		spawnerTilesetRepo,
		explosionTilesetRepo,
		tileService,
		animationService,
	)

	// Берем первую позицию спавна игрока из конфига
	var playerSpawner types.Position
	if len(gameConfig.PlayerSpawners) > 0 &&
		len(gameConfig.PlayerSpawners[0]) == 2 {
		playerSpawner = types.Position{
			X: float64(
				gameConfig.PlayerSpawners[0][0],
			) * use_cases.TankSpriteSize,
			Y: float64(
				gameConfig.PlayerSpawners[0][1],
			) * use_cases.TankSpriteSize,
		}
	} else {
		// Значение по умолчанию
		playerSpawner = types.Position{X: 12 * use_cases.TankSpriteSize, Y: 24 * use_cases.TankSpriteSize}
	}

	// Создаем сервисы для пули
	bulletTileService := services.NewTileService(bulletTilesetRepo)
	bulletAnimationService := services.NewAnimationService()
	bulletTilesUseCases := use_cases.NewTilesUseCases(
		bulletTilesetRepo,
		bulletTileService,
		bulletAnimationService,
	)
	bulletUseCases := use_cases.NewBulletUseCases(
		gameRepo.BulletsRepository(),
		bulletTilesUseCases,
	)

	// Создаем танк игрока
	playerTank := &types.TankEntity{
		Position:  playerSpawner,
		Speed:     0,
		Direction: types.DirectionUp,
		State:     types.TankStateSpawning,
		Altitude:  types.SURFACE,
	}
	gameRepo.TanksRepository().AddTank(playerTank)

	// Создаем TankRenderUseCases для игрока
	playerRenderUseCases := use_cases.NewTankRenderUseCases()

	// Создаем TankLifecycleUseCases для игрока
	playerLifecycle := use_cases.NewTankLifecycleUseCases(
		tilesUseCasesWithAnimations,
		playerRenderUseCases,
	)

	playerUseCases := use_cases.NewTankCommonUseCases(
		bulletUseCases,
		tilesUseCasesWithAnimations,
		tankBrakingService,
		coordinateService,
	)

	// Создаем TankActionsUseCases для игрока (для Stop в коллизиях)
	playerTankActions := use_cases.NewTankActionsUseCases(
		tankBrakingService,
		coordinateService,
		bulletUseCases,
	)

	mapUseCases := use_cases.NewMapUseCases(gameRepo.BlocksRepository())

	// Создаем AI контекст
	blocks := mapUseCases.GetBlocks()
	aiContext := &types.GameAiContext{
		Player:  nil, // Будет обновляться в Update
		Enemies: nil, // Будет обновляться в Update
		Bullets: nil, // Будет обновляться в Update
		Blocks:  blocks,
	}

	// Получаем интервал обновления AI в тиках (по умолчанию 60 тиков)
	updateInterval := 60
	if gameConfig.AIUpdateIntervalTicks > 0 {
		updateInterval = gameConfig.AIUpdateIntervalTicks
	}

	// Загружаем скрипт AI для врагов один раз
	enemyScript, err := scriptsRepo.GetScript("enemies")
	if err != nil {
		return nil, err
	}

	// Создаем до 3 врагов
	enemyUseCasesList := make([]interfaces.ITankCommonUseCases, 0, 3)
	enemyRenderUseCasesList := make([]interfaces.ITankRenderUseCases, 0, 3)
	enemyLifecyclesList := make([]interfaces.ITankLifecycleUseCases, 0, 3)
	enemyInputAdapters := make([]interfaces.IInputAdapter, 0, 3)
	enemyTanksEntities := make([]*types.TankEntity, 0, 3)

	for i, spawner := range gameConfig.EnemySpawners {
		if i >= 3 { // Максимум 3 врага
			break
		}
		if len(spawner) != 2 {
			continue
		}

		// Конвертируем координаты в пиксели
		position := types.Position{
			X: float64(spawner[0]) * use_cases.TankSpriteSize,
			Y: float64(spawner[1]) * use_cases.TankSpriteSize,
		}

		// Создаем танк врага
		enemyTank := &types.TankEntity{
			Position:  position,
			Speed:     0,
			Direction: types.DirectionDown,
			State:     types.TankStateSpawning,
			Altitude:  types.SURFACE,
		}
		gameRepo.TanksRepository().AddTank(enemyTank)

		// Создаем TankRenderUseCases для врага
		enemyRenderUseCases := use_cases.NewTankRenderUseCases()

		// Создаем TankLifecycleUseCases для врага
		enemyLifecycle := use_cases.NewTankLifecycleUseCases(
			tilesUseCasesWithAnimations,
			enemyRenderUseCases,
		)

		enemyUseCases := use_cases.NewTankCommonUseCases(
			bulletUseCases,
			tilesUseCasesWithAnimations,
			tankBrakingService,
			coordinateService,
		)

		// Запускаем спавн танка врага через lifecycle
		err := enemyLifecycle.Spawn(enemyTank)
		if err != nil {
			return nil, err
		}

		enemyTanksEntities = append(enemyTanksEntities, enemyTank)
		enemyUseCasesList = append(enemyUseCasesList, enemyUseCases)
		enemyRenderUseCasesList = append(
			enemyRenderUseCasesList,
			enemyRenderUseCases,
		)
		enemyLifecyclesList = append(enemyLifecyclesList, enemyLifecycle)

		// Создаем TankActionsUseCases для AI инпут-адаптера врага
		enemyTankActions := use_cases.NewTankActionsUseCases(
			tankBrakingService,
			coordinateService,
			bulletUseCases,
		)

		// Создаем AI input адаптер для этого врага
		aiInputAdapter, err := input_adapters.NewAiInputAdapter(
			enemyTankActions,
			enemyTank,
			aiContext,
			updateInterval,
			enemyScript,
		)
		if err != nil {
			return nil, err
		}
		enemyInputAdapters = append(enemyInputAdapters, aiInputAdapter)
	}

	// Создаем временный BulletCollisionService с заглушкой для CheckColliders
	tempBulletCollisionService := services.NewBulletCollisionService(
		use_cases.TileMinSize,
		func(obj1 types.IMapObject, obj2 types.IMapObject) bool {
			return false // Временная заглушка
		},
	)

	// Преобразуем []*TankCommonUseCases в []*TankCommonUseCases для CollisionUseCases
	enemyUseCasesRefs := enemyUseCasesList

	// Создаем CollisionUseCases первый раз для получения CheckColliders
	collisionUseCases := use_cases.NewCollisionUseCasesWithEnemies(
		bulletUseCases,
		playerTank,
		playerTankActions,
		mapUseCases,
		enemyTanksEntities,
		enemyUseCasesRefs,
		enemyLifecyclesList,
		boundaryCollisionService,
		wallCollisionService,
		tempBulletCollisionService,
	)

	// Создаем правильный BulletCollisionService с реальным CheckColliders из CollisionUseCases
	bulletCollisionService := services.NewBulletCollisionService(
		use_cases.TileMinSize,
		func(obj1 types.IMapObject, obj2 types.IMapObject) bool {
			return collisionUseCases.CheckColliders(obj1, obj2)
		},
	)

	// Пересоздаем CollisionUseCases с правильным BulletCollisionService
	collisionUseCases = use_cases.NewCollisionUseCasesWithEnemies(
		bulletUseCases,
		playerTank,
		playerTankActions,
		mapUseCases,
		enemyTanksEntities,
		enemyUseCasesRefs,
		enemyLifecyclesList,
		boundaryCollisionService,
		wallCollisionService,
		bulletCollisionService,
	)

	return &GameStateUseCasesFacade{
		playerUseCases:         playerUseCases,
		playerRenderUseCases:   playerRenderUseCases,
		playerLifecycle:        playerLifecycle,
		playerTank:             playerTank,
		bulletUseCases:         bulletUseCases,
		mapUseCases:            mapUseCases,
		collisionUseCases:      collisionUseCases,
		enemyUseCases:          enemyUseCasesList,
		enemyRenderUseCases:    enemyRenderUseCasesList,
		enemyLifecycles:        enemyLifecyclesList,
		enemyTanksEntities:     enemyTanksEntities,
		enemyInputAdapters:     enemyInputAdapters,
		tilesUseCasesWithAnims: tilesUseCasesWithAnimations,
		aiContext:              aiContext,
	}, nil
}

// Update обновляет игровое состояние
func (g *GameStateUseCasesFacade) Update(dt float64) {
	// Обновляем игрока
	if g.playerUseCases != nil {
		playerTank := g.playerTank
		if playerTank != nil {
			if err := g.playerUseCases.Update(playerTank, dt); err != nil {
				// Ошибка обновления игнорируется
				_ = err
			}
		}
	}

	// Обновляем контекст AI с данными об игроке, врагах и пулях
	if g.aiContext != nil {
		bullets := g.bulletUseCases.GetBullets()
		g.aiContext.Player = g.playerTank
		g.aiContext.Enemies = g.enemyTanksEntities
		g.aiContext.Bullets = bullets
	}

	// Обновляем AI input адаптеры врагов (принимают решения о движении)
	for _, adapter := range g.enemyInputAdapters {
		if adapter != nil {
			adapter.Update(dt)
		}
	}

	// Обновляем движение врагов
	for i, enemyUseCases := range g.enemyUseCases {
		if enemyUseCases != nil {
			enemyTank := g.enemyTanksEntities[i]
			if enemyTank != nil {
				if err := enemyUseCases.Update(enemyTank, dt); err != nil {
					// Ошибка обновления игнорируется
					_ = err
				}
			}
		}
	}

	// Обновляем пули
	g.bulletUseCases.UpdateBullets(dt)

	// Проверяем коллизии ПОСЛЕ движения всех объектов
	g.collisionUseCases.UpdateCollisions()
}

// UpdateAnimations обновляет все анимации из репозитория
func (g *GameStateUseCasesFacade) UpdateAnimations() {
	if g.tilesUseCasesWithAnims != nil {
		g.tilesUseCasesWithAnims.UpdateAnimations()
	}
}

// UpdateTankSpawn обновляет процесс спавна танка
func (g *GameStateUseCasesFacade) UpdateTankSpawn(currentTime float64) {
	if g.playerLifecycle != nil && g.playerTank != nil {
		g.playerLifecycle.IsSpawnFinished(g.playerTank, currentTime)
		g.playerLifecycle.IsExplosionFinished(g.playerTank)
	}
}

// UpdateEnemiesSpawn обновляет процесс спавна врагов
func (g *GameStateUseCasesFacade) UpdateEnemiesSpawn(currentTime float64) {
	for i, enemyLifecycle := range g.enemyLifecycles {
		if enemyLifecycle != nil && i < len(g.enemyTanksEntities) {
			enemyTank := g.enemyTanksEntities[i]
			if enemyTank != nil {
				enemyLifecycle.IsSpawnFinished(enemyTank, currentTime)
				enemyLifecycle.IsExplosionFinished(enemyTank)
			}
		}
	}
}

// UpdateEnemiesAnimations обновляет анимации врагов
// Устаревший метод, используйте UpdateAnimations()
func (g *GameStateUseCasesFacade) UpdateEnemiesAnimations() {
	g.UpdateAnimations()
}

// StartTankSpawn запускает спавн танка игрока
func (g *GameStateUseCasesFacade) StartTankSpawn(spawnStartTime float64) {
	if g.playerLifecycle != nil && g.playerTank != nil {
		err := g.playerLifecycle.Spawn(g.playerTank)
		if err != nil {
			panic(err)
		}

		// Устанавливаем время спавна
		g.playerTank.SpawnedAt = spawnStartTime
	}
}

// Getter методы для доступа к use cases
func (g *GameStateUseCasesFacade) TankUseCases() interfaces.ITankCommonUseCases {
	return g.playerUseCases
}

func (g *GameStateUseCasesFacade) GetPlayerTank() *types.TankEntity {
	return g.playerTank
}

func (g *GameStateUseCasesFacade) BulletUseCases() *use_cases.BulletUseCases {
	return g.bulletUseCases
}

func (g *GameStateUseCasesFacade) MapUseCases() *use_cases.MapUseCases {
	return g.mapUseCases
}

func (g *GameStateUseCasesFacade) CollisionUseCases() *use_cases.CollisionUseCases {
	return g.collisionUseCases
}

func (g *GameStateUseCasesFacade) GetEnemyTanks() []*types.TankEntity {
	return g.enemyTanksEntities
}

func (g *GameStateUseCasesFacade) GetEnemyUseCases() []interfaces.ITankCommonUseCases {
	return g.enemyUseCases
}

func (g *GameStateUseCasesFacade) PlayerRenderUseCases() interfaces.ITankRenderUseCases {
	return g.playerRenderUseCases
}

func (g *GameStateUseCasesFacade) GetEnemyRenderUseCases() []interfaces.ITankRenderUseCases {
	return g.enemyRenderUseCases
}
