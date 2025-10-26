package states

import (
	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// GameStateUseCasesFacade - фасад для оркестрации use cases игрового состояния
type GameStateUseCasesFacade struct {
	playerUseCases    *use_cases.PlayerUseCases
	bulletUseCases    *use_cases.BulletUseCases
	mapUseCases       *use_cases.MapUseCases
	collisionUseCases *use_cases.CollisionUseCases
	animationUseCases *use_cases.AnimationUseCases
	enemyUseCasesList []*use_cases.EnemyUseCases // Массив врагов (до 3 штук)
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

	// Создаем репозитории
	blocksRepo := game.NewBlocksRepository()
	bulletsRepo := game.NewBulletsRepository()
	animationsRepo := game.NewAnimationsRepository()
	tanksRepo := game.NewTanksRepository()

	// Заполняем репозиторий блоков данными уровня
	for _, block := range level {
		blocksRepo.AddBlock(block)
	}

	// Создаем Use Cases
	animationUseCases := use_cases.NewAnimationUseCases(animationsRepo)
	tankUseCases := use_cases.NewTankUseCases(tanksRepo, playerTilesetRepo, spawnerTilesetRepo, animationUseCases)
	playerUseCases := use_cases.NewPlayerUseCases(tankUseCases, animationUseCases, gameConfig.PlayerSpawners)
	bulletTilesUseCases := use_cases.NewTilesUseCases(bulletTilesetRepo)
	bulletUseCases := use_cases.NewBulletUseCases(bulletsRepo, bulletTilesUseCases)
	mapUseCases := use_cases.NewMapUseCases(blocksRepo)

	// Создаем до 3 врагов
	enemyUseCasesList := make([]*use_cases.EnemyUseCases, 0, 3)
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
		if err := enemy.InitEnemy(position); err != nil {
			return nil, err
		}

		enemyUseCasesList = append(enemyUseCasesList, enemy)
	}

	// Создаем CollisionUseCases
	collisionUseCases := use_cases.NewCollisionUseCasesWithEnemies(
		bulletUseCases,
		playerUseCases,
		mapUseCases,
		enemyUseCasesList,
	)

	return &GameStateUseCasesFacade{
		playerUseCases:    playerUseCases,
		bulletUseCases:    bulletUseCases,
		mapUseCases:       mapUseCases,
		collisionUseCases: collisionUseCases,
		animationUseCases: animationUseCases,
		enemyUseCasesList: enemyUseCasesList,
	}, nil
}

// Update обновляет игровое состояние
func (g *GameStateUseCasesFacade) Update() {
	g.playerUseCases.MoveTank(g.playerUseCases.GetDirection(), use_cases.DT)
	g.animationUseCases.UpdateAnimations()
	g.bulletUseCases.UpdateBullets(use_cases.DT)
	g.collisionUseCases.UpdateCollisions()
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
