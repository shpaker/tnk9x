package states

import (
	"github.com/shpaker/gonflict/internal/adapters/input_adapters"
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
	enemyTanks             []*types.TankEntity            // Массив танков врагов для обратной совместимости
	enemyInputAdapters     []input_adapters.IInputAdapter // AI input адаптеры для врагов
	tilesUseCasesWithAnims *use_cases.TilesUseCases       // Общий tilesUseCases для всех анимаций
	aiContext              *types.GameAiContext           // AI контекст для всех адаптеров
}

// NewGameStateUseCasesFacade создает фасад для оркестрации use cases игрового состояния
func NewGameStateUseCasesFacade(
	mapsRepo processed.IMapsDataRepository,
	scriptsRepo processed.IScriptsRepository,
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
	tilesUseCasesWithAnimations := use_cases.NewTilesUseCasesWithAnimations(
		playerTilesetRepo,
		gameRepo.AnimationsRepository(),
		spawnerTilesetRepo,
		explosionTilesetRepo,
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

	bulletTilesUseCases := use_cases.NewTilesUseCases(bulletTilesetRepo)
	bulletUseCases := use_cases.NewBulletUseCases(
		gameRepo.BulletsRepository(),
		bulletTilesUseCases,
	)
	playerUseCases := use_cases.NewTankUseCases(
		gameRepo.TanksRepository(),
		bulletUseCases,
		tilesUseCasesWithAnimations,
		playerSpawner,
		types.DirectionUp,
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
	enemyInputAdapters := make([]input_adapters.IInputAdapter, 0, 3)
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
		enemyTankUseCases := use_cases.NewTankUseCases(
			gameRepo.TanksRepository(),
			bulletUseCases,
			tilesUseCasesWithAnimations,
			position,
			types.DirectionDown,
		)

		// Создаем танк врага
		err := enemyTankUseCases.StartSpawn()
		if err != nil {
			return nil, err
		}

		// Получаем танк врага
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
		enemyInputAdapters:     enemyInputAdapters,
		tilesUseCasesWithAnims: tilesUseCasesWithAnimations,
		aiContext:              aiContext,
	}, nil
}

// Update обновляет игровое состояние
func (g *GameStateUseCasesFacade) Update() {
	// Обновляем игрока
	tank := g.playerUseCases.GetTank()
	if tank != nil {
		g.playerUseCases.Update(use_cases.DT)
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
			adapter.Update()
		}
	}

	// Обновляем пули
	g.bulletUseCases.UpdateBullets(use_cases.DT)

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

func (g *GameStateUseCasesFacade) GetEnemyUseCases() []use_cases.ITankUseCasesRef {
	result := make([]use_cases.ITankUseCasesRef, len(g.enemyUseCases))
	for i, uc := range g.enemyUseCases {
		result[i] = uc
	}
	return result
}
