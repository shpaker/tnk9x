package use_cases

import (
	"errors"

	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/utils"
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
		Position:        position,
		Speed:           0,
		Direction:       direction,
		State:           types.TankStateSpawning, // Танк спавнится
		SpawnedAt:       0,
		Altitude:        types.SURFACE,
	}

	// Добавляем танк в репозиторий
	uc.tanksRepo.AddTank(tank)

	return tank, spawnAnimation, tankAnimation, nil
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

// RotateTank поворачивает танк в указанном направлении
func (uc *TankUseCases) RotateTank(tank *types.TankEntity, direction types.Direction) error {
	if tank == nil {
		return errors.New("tank is nil")
	}
	if tank.State == types.TankStateSpawning {
		return errors.New("tank is not spawned yet")
	}

	if tank.Speed != 0 {
		return errors.New("cannot rotate while moving")
	}

	tank.Speed = 32.0
	if tank.Direction == direction {
		return nil
	}

	tank.Direction = direction
	return nil
}

// StopTank останавливает танк
func (uc *TankUseCases) StopTank(tank *types.TankEntity, byCollision bool) error {
	if tank == nil {
		return errors.New("tank is nil")
	}
	if tank.State == types.TankStateSpawning {
		return errors.New("tank is not spawned yet")
	}

	tank.Speed = 0

	if byCollision {
		// Округляем координаты до ближайшего кратного 4
		tank.Position.X = utils.RoundToNearestMultipleOf4(tank.Position.X)
		tank.Position.Y = utils.RoundToNearestMultipleOf4(tank.Position.Y)
		return nil
	}

	// Выравниваем позицию по сетке
	switch tank.Direction {
	case types.DirectionUp:
		tank.Position.Y = float64(utils.RoundToEven(tank.Position.Y, false))
	case types.DirectionDown:
		tank.Position.Y = float64(utils.RoundToEven(tank.Position.Y, true))
	case types.DirectionLeft:
		tank.Position.X = float64(utils.RoundToEven(tank.Position.X, false))
	case types.DirectionRight:
		tank.Position.X = float64(utils.RoundToEven(tank.Position.X, true))
	}

	return nil
}

// MoveTank перемещает танк в указанном направлении
func (uc *TankUseCases) MoveTank(tank *types.TankEntity, direction types.Direction, dt float64) error {
	if tank == nil {
		return errors.New("tank is nil")
	}
	if tank.State == types.TankStateSpawning {
		return errors.New("tank is not spawned yet")
	}

	delta := tank.Speed * dt

	switch tank.Direction {
	case types.DirectionUp:
		tank.Position.Y -= delta
	case types.DirectionDown:
		tank.Position.Y += delta
	case types.DirectionLeft:
		tank.Position.X -= delta
	case types.DirectionRight:
		tank.Position.X += delta
	}

	return nil
}
