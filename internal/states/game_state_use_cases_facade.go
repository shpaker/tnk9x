package states

import (
	"github.com/shpaker/gonflict/internal/adapters"
	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// GameStateUseCasesFacade - фасад для оркестрации use cases игрового состояния
type GameStateUseCasesFacade struct {
	tankUseCases      *use_cases.TankUseCases
	playerUseCases    *use_cases.PlayerUseCases
	bulletUseCases    *use_cases.BulletUseCases
	mapUseCases       *use_cases.MapUseCases
	collisionUseCases *use_cases.CollisionUseCases
	animationUseCases *use_cases.AnimationUseCases
	enemyUseCasesList []*use_cases.EnemyUseCases // Массив врагов (до 3 штук)
	enemyControllers  []*adapters.AIController   // AI контроллеры для врагов
}

// NewGameStateUseCasesFacade создает фасад для оркестрации use cases игрового состояния
func NewGameStateUseCasesFacade(
	mapsRepo processed.IMapsDataRepository,
	levelNumber int,
	mapTilesetRepo processed.ITilesetRepository,
	playerTilesetRepo processed.ITilesetRepository,
	bulletTilesetRepo processed.ITilesetRepository,
	spawnerTilesetRepo processed.ITilesetRepository,
	explosionTilesetRepo processed.ITilesetRepository,
	gameConfig *GameConfig,
) (*GameStateUseCasesFacade, error) {
	// Загружаем уровень
	level, err := mapsRepo.GetLevel(levelNumber)
	if err != nil {
		return nil, err
	}

	// Создаем реестр игровых репозиториев
	gameRepo := game.NewGameRepositoriesRegistry()

	// Заполняем репозиторий блоков данными уровня
	for _, block := range level {
		gameRepo.BlocksRepository().AddBlock(block)
	}

	// Создаем Use Cases
	animationUseCases := use_cases.NewAnimationUseCases(gameRepo.AnimationsRepository())
	tankUseCases := use_cases.NewTankUseCases(gameRepo.TanksRepository(), playerTilesetRepo, spawnerTilesetRepo, animationUseCases)
	playerUseCases := use_cases.NewPlayerUseCases(tankUseCases, animationUseCases, gameConfig.PlayerSpawners)
	bulletTilesUseCases := use_cases.NewTilesUseCases(bulletTilesetRepo)
	bulletUseCases := use_cases.NewBulletUseCases(gameRepo.BulletsRepository(), bulletTilesUseCases)
	mapUseCases := use_cases.NewMapUseCases(gameRepo.BlocksRepository())

	// Создаем AI
	ai, err := adapters.NewEnemyAILua("assets/scripts/enemies.lua")
	if err != nil {
		return nil, err
	}

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

	// Создаем AIUseCases
	aiUseCases := use_cases.NewAIUseCases(ai, aiContext, updateInterval)

	// Создаем до 3 врагов
	enemyUseCasesList := make([]*use_cases.EnemyUseCases, 0, 3)
	enemyControllers := make([]*adapters.AIController, 0, 3)
	for i, spawner := range gameConfig.EnemySpawners {
		if i >= 3 { // Максимум 3 врага
			break
		}
		if len(spawner) != 2 {
			continue
		}

		// Создаем отдельного врага
		enemy := use_cases.NewEnemyUseCases(tankUseCases, explosionTilesetRepo, animationUseCases)

		// Конвертируем координаты в пиксели
		position := types.Position{
			X: float64(spawner[0]) * use_cases.TankSpriteSize,
			Y: float64(spawner[1]) * use_cases.TankSpriteSize,
		}

		// Инициализируем врага
		if err := enemy.StartEnemySpawn(position); err != nil {
			return nil, err
		}

		// Создаем AI контроллер для этого врага
		enemyTank := enemy.GetEnemy()
		aiController := adapters.NewAIController(tankUseCases, bulletUseCases, aiUseCases, enemyTank)
		enemyControllers = append(enemyControllers, aiController)

		enemyUseCasesList = append(enemyUseCasesList, enemy)
	}

	// Создаем CollisionUseCases
	collisionUseCases := use_cases.NewCollisionUseCasesWithEnemies(
		bulletUseCases,
		playerUseCases,
		tankUseCases,
		mapUseCases,
		enemyUseCasesList,
	)

	return &GameStateUseCasesFacade{
		tankUseCases:      tankUseCases,
		playerUseCases:    playerUseCases,
		bulletUseCases:    bulletUseCases,
		mapUseCases:       mapUseCases,
		collisionUseCases: collisionUseCases,
		animationUseCases: animationUseCases,
		enemyUseCasesList: enemyUseCasesList,
		enemyControllers:  enemyControllers,
	}, nil
}

// Update обновляет игровое состояние
func (g *GameStateUseCasesFacade) Update() {
	// Двигаем игрока
	tank, _ := g.playerUseCases.GetTank()
	if tank != nil {
		g.tankUseCases.MoveTank(tank, tank.Direction, use_cases.DT)
	}

	// Обновляем AI контроллеры врагов (они сами управляют движением)
	for _, aiController := range g.enemyControllers {
		if aiController != nil {
			aiController.Update()
		}
	}

	// Обновляем пули
	g.bulletUseCases.UpdateBullets(use_cases.DT)

	// Проверяем коллизии ПОСЛЕ движения всех объектов
	g.collisionUseCases.UpdateCollisions()

	// Обновляем анимации
	g.animationUseCases.UpdateAnimations()
}

// UpdateTankSpawn обновляет процесс спавна танка
func (g *GameStateUseCasesFacade) UpdateTankSpawn(currentTime float64) {
	g.playerUseCases.UpdateSpawn(currentTime)
}

// UpdateEnemiesSpawn обновляет процесс спавна врагов
func (g *GameStateUseCasesFacade) UpdateEnemiesSpawn(currentTime float64) {
	for _, enemyUseCases := range g.enemyUseCasesList {
		enemyUseCases.UpdateEnemiesSpawn(currentTime)
	}
}

// UpdateEnemiesAnimations обновляет анимации врагов
func (g *GameStateUseCasesFacade) UpdateEnemiesAnimations() {
	for _, enemyUseCases := range g.enemyUseCasesList {
		enemyUseCases.UpdateEnemiesAnimations()
	}
}

// StartTankSpawn запускает спавн танка
func (g *GameStateUseCasesFacade) StartTankSpawn(spawnStartTime float64) {
	g.playerUseCases.StartSpawn(spawnStartTime)
}

// Getter методы для доступа к use cases
func (g *GameStateUseCasesFacade) TankUseCases() *use_cases.PlayerUseCases {
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

func (g *GameStateUseCasesFacade) AnimationUseCases() *use_cases.AnimationUseCases {
	return g.animationUseCases
}

func (g *GameStateUseCasesFacade) EnemyUseCases() *use_cases.EnemyUseCases {
	// Возвращаем первый враг для обратной совместимости
	// TODO: обновить интерфейс для работы с массивом
	if len(g.enemyUseCasesList) > 0 {
		return g.enemyUseCasesList[0]
	}
	return nil
}

func (g *GameStateUseCasesFacade) GetEnemyUseCasesList() []*use_cases.EnemyUseCases {
	return g.enemyUseCasesList
}

func (g *GameStateUseCasesFacade) TankUseCasesRef() *use_cases.TankUseCases {
	return g.tankUseCases
}
