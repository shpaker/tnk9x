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
	playerUseCases         *use_cases.TankUseCases
	bulletUseCases         *use_cases.BulletUseCases
	mapUseCases            *use_cases.MapUseCases
	collisionUseCases      *use_cases.CollisionUseCases
	enemyUseCases          []*use_cases.TankUseCases
	enemyTanks             []*types.TankEntity      // Массив танков врагов для обратной совместимости
	enemyControllers       []*adapters.AIController // AI контроллеры для врагов
	tilesUseCasesWithAnims *use_cases.TilesUseCases // Общий tilesUseCases для всех анимаций
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
	tilesUseCasesWithAnimations := use_cases.NewTilesUseCasesWithAnimations(playerTilesetRepo, gameRepo.AnimationsRepository(), spawnerTilesetRepo, explosionTilesetRepo)

	// Берем первую позицию спавна игрока из конфига
	var playerSpawner types.Position
	if len(gameConfig.PlayerSpawners) > 0 && len(gameConfig.PlayerSpawners[0]) == 2 {
		playerSpawner = types.Position{
			X: float64(gameConfig.PlayerSpawners[0][0]) * use_cases.TankSpriteSize,
			Y: float64(gameConfig.PlayerSpawners[0][1]) * use_cases.TankSpriteSize,
		}
	} else {
		// Значение по умолчанию
		playerSpawner = types.Position{X: 12 * use_cases.TankSpriteSize, Y: 24 * use_cases.TankSpriteSize}
	}

	playerUseCases := use_cases.NewTankUseCases(gameRepo.TanksRepository(), tilesUseCasesWithAnimations, playerSpawner)
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
	enemyUseCasesList := make([]*use_cases.TankUseCases, 0, 3)
	enemyControllers := make([]*adapters.AIController, 0, 3)
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

		// Создаем отдельный TankUseCases для этого врага
		enemyTankUseCases := use_cases.NewTankUseCases(gameRepo.TanksRepository(), tilesUseCasesWithAnimations, position)

		// Создаем танк врага
		enemyTank, _, err := enemyTankUseCases.StartTankSpawn(position)
		if err != nil {
			return nil, err
		}

		// Устанавливаем направление врага вниз
		enemyTank.Direction = types.DirectionDown

		// Сохраняем танк в TankUseCases
		enemyTankUseCases.SetEnemyTank(enemyTank)
		enemyTanks = append(enemyTanks, enemyTank)
		enemyUseCasesList = append(enemyUseCasesList, enemyTankUseCases)

		// Создаем AI контроллер для этого врага
		aiController := adapters.NewAIController(enemyTankUseCases, bulletUseCases, aiUseCases, enemyTank)
		enemyControllers = append(enemyControllers, aiController)
	}

	// Создаем CollisionUseCases
	collisionUseCases := use_cases.NewCollisionUseCasesWithEnemies(
		bulletUseCases,
		playerUseCases,
		mapUseCases,
		enemyTanks,
	)

	return &GameStateUseCasesFacade{
		playerUseCases:         playerUseCases,
		bulletUseCases:         bulletUseCases,
		mapUseCases:            mapUseCases,
		collisionUseCases:      collisionUseCases,
		enemyUseCases:          enemyUseCasesList,
		enemyTanks:             enemyTanks,
		enemyControllers:       enemyControllers,
		tilesUseCasesWithAnims: tilesUseCasesWithAnimations,
	}, nil
}

// Update обновляет игровое состояние
func (g *GameStateUseCasesFacade) Update() {
	// Двигаем игрока
	tank, _ := g.playerUseCases.GetPlayerTank()
	if tank != nil {
		g.playerUseCases.MoveTank(tank, tank.Direction, use_cases.DT)
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
	g.playerUseCases.UpdateAnimations()
}

// UpdateTankSpawn обновляет процесс спавна танка
func (g *GameStateUseCasesFacade) UpdateTankSpawn(currentTime float64) {
	g.playerUseCases.UpdatePlayerSpawn(currentTime)
}

// UpdateEnemiesSpawn обновляет процесс спавна врагов
func (g *GameStateUseCasesFacade) UpdateEnemiesSpawn(currentTime float64) {
	for _, enemyUseCases := range g.enemyUseCases {
		enemyUseCases.UpdateEnemySpawn(currentTime)
	}
}

// UpdateEnemiesAnimations обновляет анимации врагов
// Теперь анимации управляются через TilesUseCases
func (g *GameStateUseCasesFacade) UpdateEnemiesAnimations() {
	if g.tilesUseCasesWithAnims != nil {
		g.tilesUseCasesWithAnims.UpdateAnimations()
	}
}

// StartTankSpawn запускает спавн танка игрока
func (g *GameStateUseCasesFacade) StartTankSpawn(spawnStartTime float64) {
	// Получаем позицию спавна из конфига
	spawnPosition := g.playerUseCases.GetPlayerSpawner()

	// Создаем танк игрока через TankUseCases
	player, _, err := g.playerUseCases.StartTankSpawn(
		spawnPosition,
	)
	if err != nil {
		panic(err)
	}

	// Сохраняем ссылку на танк
	g.playerUseCases.SetPlayerTank(player)
	player.SpawnedAt = spawnStartTime
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

func (g *GameStateUseCasesFacade) AnimationUseCases() *use_cases.TilesUseCases {
	// Возвращаем tilesUseCases через публичный метод
	// TODO: добавить геттер в TankUseCases для tilesUseCases
	return nil // Временно возвращаем nil, так как tilesUseCases теперь private
}

func (g *GameStateUseCasesFacade) GetEnemyTanks() []*types.TankEntity {
	return g.enemyTanks
}

func (g *GameStateUseCasesFacade) TankUseCasesRef() *use_cases.TankUseCases {
	return g.playerUseCases
}
