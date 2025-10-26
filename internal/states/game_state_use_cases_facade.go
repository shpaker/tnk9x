package states

import (
	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// GameStateUseCasesFacade - фасад для оркестрации use cases игрового состояния
type GameStateUseCasesFacade struct {
	tankUseCases      *use_cases.TankUseCases
	bulletUseCases    *use_cases.BulletUseCases
	mapUseCases       *use_cases.MapUseCases
	collisionUseCases *use_cases.CollisionUseCases
	animationUseCases *use_cases.AnimationUseCases
}

// NewGameStateUseCasesFacade создает фасад для оркестрации use cases игрового состояния
func NewGameStateUseCasesFacade(
	mapsRepo processed.IMapsDataRepository,
	levelNumber int,
	mapTilesetRepo processed.ITilesetRepository,
	playerTilesetRepo processed.ITilesetRepository,
	bulletTilesetRepo processed.ITilesetRepository,
	spawnerTilesetRepo processed.ITilesetRepository,
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

	// Заполняем репозиторий блоков данными уровня
	for _, block := range level {
		blocksRepo.AddBlock(block)
	}

	// Создаем Use Cases
	animationUseCases := use_cases.NewAnimationUseCases(animationsRepo)
	tankUseCases := use_cases.NewTankUseCases(playerTilesetRepo, spawnerTilesetRepo, animationUseCases)
	bulletTilesUseCases := use_cases.NewTilesUseCases(bulletTilesetRepo)
	bulletUseCases := use_cases.NewBulletUseCases(bulletsRepo, bulletTilesUseCases)
	mapUseCases := use_cases.NewMapUseCases(blocksRepo)
	collisionUseCases := use_cases.NewCollisionUseCases(
		bulletUseCases,
		tankUseCases,
		mapUseCases,
	)

	return &GameStateUseCasesFacade{
		tankUseCases:      tankUseCases,
		bulletUseCases:    bulletUseCases,
		mapUseCases:       mapUseCases,
		collisionUseCases: collisionUseCases,
		animationUseCases: animationUseCases,
	}, nil
}

// Update обновляет игровое состояние
func (g *GameStateUseCasesFacade) Update() {
	g.tankUseCases.MoveTank(g.tankUseCases.GetDirection(), use_cases.DT)
	g.animationUseCases.UpdateAnimations()
	g.bulletUseCases.UpdateBullets(use_cases.DT)
	g.collisionUseCases.UpdateCollisions()
}

// UpdateTankSpawn обновляет процесс спавна танка
func (g *GameStateUseCasesFacade) UpdateTankSpawn(currentTime float64) {
	g.tankUseCases.UpdateSpawn(currentTime)
}

// StartTankSpawn запускает спавн танка
func (g *GameStateUseCasesFacade) StartTankSpawn(spawnStartTime float64) {
	g.tankUseCases.StartSpawn(spawnStartTime)
}

// Getter методы для доступа к use cases
func (g *GameStateUseCasesFacade) TankUseCases() *use_cases.TankUseCases {
	return g.tankUseCases
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
