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
	playerUseCases         *use_cases.TankUseCases
	bulletUseCases         *use_cases.BulletUseCases
	mapUseCases            *use_cases.MapUseCases
	collisionUseCases      *use_cases.CollisionUseCases
	enemyUseCases          []*use_cases.TankUseCases
	enemyTanks             []*types.TankEntity        // Массив танков врагов для обратной совместимости
	enemyInputAdapters     []interfaces.IInputAdapter // AI input адаптеры для врагов
	tilesUseCasesWithAnims *use_cases.TilesUseCases   // Общий tilesUseCases для всех анимаций
	aiContext              *types.GameAiContext       // AI контекст для всех адаптеров
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

	// Создаем use case с внедренными сервисами
	playerUseCases := use_cases.NewTankUseCases(
		gameRepo.TanksRepository(),
		bulletUseCases,
		tilesUseCasesWithAnimations,
		playerSpawner,
		types.DirectionUp,
		tankBrakingService,
		coordinateService,
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
	enemyUseCasesList := make([]*use_cases.TankUseCases, 0, 3)
	enemyInputAdapters := make([]interfaces.IInputAdapter, 0, 3)
	enemyTanks := make([]*types.TankEntity, 0, 3)

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

		// Создаем TankUseCases для врага с внедренными сервисами
		enemyTankUseCases := use_cases.NewTankUseCases(
			gameRepo.TanksRepository(),
			bulletUseCases,
			tilesUseCasesWithAnimations,
			position,
			types.DirectionDown,
			tankBrakingService,
			coordinateService,
		)

		// Создаем танк врага
		err := enemyTankUseCases.StartSpawn()
		if err != nil {
			return nil, err
		}

		// Получаем танк врага для добавления в список
		enemyTank := enemyTankUseCases.GetTank()

		enemyTanks = append(enemyTanks, enemyTank)
		enemyUseCasesList = append(enemyUseCasesList, enemyTankUseCases)

		// Создаем AI input адаптер для этого врага с поддержкой Lua
		aiInputAdapter, err := input_adapters.NewAiInputAdapter(
			enemyTankUseCases,
			aiContext,
			updateInterval,
			enemyScript,
		)
		if err != nil {
			return nil, err
		}
		enemyInputAdapters = append(enemyInputAdapters, aiInputAdapter)
	}

	// Конвертируем enemyUseCasesList в []ITankUseCasesRef
	enemyUseCasesRefs := make(
		[]interfaces.ITankUseCasesRef,
		len(enemyUseCasesList),
	)
	for i, uc := range enemyUseCasesList {
		enemyUseCasesRefs[i] = uc
	}

	// Создаем временный BulletCollisionService с заглушкой для CheckColliders
	tempBulletCollisionService := services.NewBulletCollisionService(
		use_cases.TileMinSize,
		func(obj1 types.IMapObject, obj2 types.IMapObject) bool {
			return false // Временная заглушка
		},
	)

	// Создаем CollisionUseCases первый раз для получения CheckColliders
	collisionUseCases := use_cases.NewCollisionUseCasesWithEnemies(
		bulletUseCases,
		playerUseCases,
		mapUseCases,
		enemyTanks,
		enemyUseCasesRefs,
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
		playerUseCases,
		mapUseCases,
		enemyTanks,
		enemyUseCasesRefs,
		boundaryCollisionService,
		wallCollisionService,
		bulletCollisionService,
	)

	return &GameStateUseCasesFacade{
		playerUseCases:         playerUseCases,
		bulletUseCases:         bulletUseCases,
		mapUseCases:            mapUseCases,
		collisionUseCases:      collisionUseCases,
		enemyUseCases:          enemyUseCasesList,
		enemyTanks:             enemyTanks,
		enemyInputAdapters:     enemyInputAdapters,
		tilesUseCasesWithAnims: tilesUseCasesWithAnimations,
		aiContext:              aiContext,
	}, nil
}

// Update обновляет игровое состояние
func (g *GameStateUseCasesFacade) Update(dt float64) {
	// Обновляем игрока
	tank := g.playerUseCases.GetTank()
	if tank != nil {
		g.playerUseCases.Update(dt)
	}

	// Обновляем контекст AI с данными об игроке, врагах и пулях
	if g.aiContext != nil {
		bullets := g.bulletUseCases.GetBullets()
		g.aiContext.Player = tank
		g.aiContext.Enemies = g.enemyTanks
		g.aiContext.Bullets = bullets
	}

	// Обновляем AI input адаптеры врагов (они сами управляют движением)
	for _, adapter := range g.enemyInputAdapters {
		if adapter != nil {
			adapter.Update(dt)
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
	g.playerUseCases.IsSpawnFinished(currentTime)
	g.playerUseCases.IsExplosionFinished()
}

// UpdateEnemiesSpawn обновляет процесс спавна врагов
func (g *GameStateUseCasesFacade) UpdateEnemiesSpawn(currentTime float64) {
	for _, enemyUseCases := range g.enemyUseCases {
		enemyUseCases.IsSpawnFinished(currentTime)
		enemyUseCases.IsExplosionFinished()
	}
}

// UpdateEnemiesAnimations обновляет анимации врагов
// Устаревший метод, используйте UpdateAnimations()
func (g *GameStateUseCasesFacade) UpdateEnemiesAnimations() {
	g.UpdateAnimations()
}

// StartTankSpawn запускает спавн танка игрока
func (g *GameStateUseCasesFacade) StartTankSpawn(spawnStartTime float64) {
	// Создаем танк игрока через TankUseCases
	err := g.playerUseCases.StartSpawn()
	if err != nil {
		panic(err)
	}

	// Устанавливаем время спавна
	tank := g.playerUseCases.GetTank()
	tank.SpawnedAt = spawnStartTime
}

// Getter методы для доступа к use cases
func (g *GameStateUseCasesFacade) TankUseCases() *use_cases.TankUseCases {
	return g.playerUseCases
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
	return g.enemyTanks
}

func (g *GameStateUseCasesFacade) GetEnemyUseCases() []interfaces.ITankUseCasesRef {
	result := make([]interfaces.ITankUseCasesRef, len(g.enemyUseCases))
	for i, uc := range g.enemyUseCases {
		result[i] = uc
	}
	return result
}
