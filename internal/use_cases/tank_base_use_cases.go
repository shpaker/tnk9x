package use_cases

import (
	"errors"

	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
)

// TankUseCases предоставляет базовые операции для работы с танками
type TankUseCases struct {
	tanksRepo          game.ITanksRepository
	tilesetRepo        processed.ITilesetRepository
	spawnerTilesetRepo processed.ITilesetRepository
	animationUseCases  IAnimationUseCases
}

// NewTankUseCases создает новый экземпляр TankUseCases
func NewTankUseCases(
	tanksRepo game.ITanksRepository,
	tilesetRepo processed.ITilesetRepository,
	spawnerTilesetRepo processed.ITilesetRepository,
	animationUseCases IAnimationUseCases,
) *TankUseCases {
	return &TankUseCases{
		tanksRepo:          tanksRepo,
		tilesetRepo:        tilesetRepo,
		spawnerTilesetRepo: spawnerTilesetRepo,
		animationUseCases:  animationUseCases,
	}
}

// CreateTankWithSpawn создает танк с анимацией спавна
func (uc *TankUseCases) CreateTankWithSpawn(
	position types.Position,
	direction types.Direction,
) (*types.TankEntity, *types.TileAnimationEntity, *types.TileAnimationEntity, error) {
	// Создаем анимацию спавна
	spawnTilesUseCases := NewTilesUseCases(uc.spawnerTilesetRepo)
	spawnAnimation, err := spawnTilesUseCases.CreateAnimationTile("spawner")
	if err != nil {
		return nil, nil, nil, err
	}
	uc.animationUseCases.AddAnimation(spawnAnimation)

	// Создаем анимацию танка
	tankTilesUseCases := NewTilesUseCases(uc.tilesetRepo)
	tankAnimation, err := tankTilesUseCases.CreateAnimationTile("base_tank")
	if err != nil {
		return nil, nil, nil, err
	}
	uc.animationUseCases.AddAnimation(tankAnimation)

	// Создаем танк
	tank := &types.TankEntity{
		AnimationGetter: tankAnimation,
		SpawnPosition:   position,
		WorldPosition:   position,
		Speed:           0,
		Direction:       direction,
		IsSpawned:       false, // Танк не заспавнен
		SpawnedAt:       0,
		Altitude:        types.SURFACE,
	}

	// Добавляем танк в репозиторий
	uc.tanksRepo.AddTank(tank)

	return tank, spawnAnimation, tankAnimation, nil
}

// CreateTankAnimation создает анимацию танка
func (uc *TankUseCases) CreateTankAnimation() (*types.TileAnimationEntity, error) {
	tankTilesUseCases := NewTilesUseCases(uc.tilesetRepo)
	tankAnimation, err := tankTilesUseCases.CreateAnimationTile("base_tank")
	if err != nil {
		return nil, err
	}
	uc.animationUseCases.AddAnimation(tankAnimation)
	return tankAnimation, nil
}

// CreateSpawnAnimation создает анимацию спавна
func (uc *TankUseCases) CreateSpawnAnimation() (*types.TileAnimationEntity, error) {
	spawnTilesUseCases := NewTilesUseCases(uc.spawnerTilesetRepo)
	spawnAnimation, err := spawnTilesUseCases.CreateAnimationTile("spawner")
	if err != nil {
		return nil, err
	}
	uc.animationUseCases.AddAnimation(spawnAnimation)
	return spawnAnimation, nil
}

// GetTank возвращает танк по индексу
func (uc *TankUseCases) GetTank(index int) (*types.TankEntity, error) {
	tanks := uc.tanksRepo.GetAllTanks()
	if index < 0 || index >= len(tanks) {
		return nil, errors.New("tank index out of range")
	}
	return tanks[index], nil
}

// GetAllTanks возвращает всех танков
func (uc *TankUseCases) GetAllTanks() []*types.TankEntity {
	return uc.tanksRepo.GetAllTanks()
}

// RemoveTank удаляет танк из репозитория по указателю
func (uc *TankUseCases) RemoveTank(tank *types.TankEntity) error {
	return uc.tanksRepo.RemoveTank(tank)
}

// AddTank добавляет танк в репозиторий
func (uc *TankUseCases) AddTank(tank *types.TankEntity) {
	uc.tanksRepo.AddTank(tank)
}
