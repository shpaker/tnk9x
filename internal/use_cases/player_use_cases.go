package use_cases

import (
	"errors"

	"github.com/shpaker/gonflict/internal/repositories"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/utils"
)

// PlayerUseCases реализация интерфейса PlayerUseCases
type PlayerUseCases struct {
	tilesetRepo repositories.ITilesetRepository
	tank        types.TankEntity
}

// NewPlayerUseCases создает новый экземпляр PlayerUseCases
func NewPlayerUseCases(tilesetRepo repositories.ITilesetRepository) *PlayerUseCases {
	uc := &PlayerUseCases{
		tilesetRepo: tilesetRepo,
	}

	// Создаем игрока при инициализации
	tank, err := uc.makePlayer()
	if err != nil {
		panic(err)
	}
	uc.tank = tank

	return uc
}

// makePlayer создает игрока с начальными параметрами
func (uc *PlayerUseCases) makePlayer() (types.TankEntity, error) {
	tankSprite, err := uc.tilesetRepo.GetImage("base_tank_1")
	if err != nil {
		return types.TankEntity{}, err
	}

	// Создаем игрока с начальными параметрами
	spawnPosition := types.Position{X: 4 * TankSpriteSize, Y: 12 * TankSpriteSize}

	player := types.TankEntity{
		Image:         tankSprite,
		SpawnPosition: spawnPosition,
		WorldPosition: types.Position{
			X: spawnPosition.X,
			Y: spawnPosition.Y,
		},
		Speed:     0,
		Direction: types.DirectionUp,
	}

	return player, nil
}

// GetPlayer возвращает данные игрока
func (uc *PlayerUseCases) GetPlayer() (*types.TankEntity, error) {
	return &uc.tank, nil
}

// GetDirection возвращает текущее направление игрока
func (uc *PlayerUseCases) GetDirection() types.Direction {
	return uc.tank.Direction
}

// RotatePlayer поворачивает игрока в указанном направлении
func (uc *PlayerUseCases) RotatePlayer(direction types.Direction) error {
	if uc.tank.Speed != 0 {
		return errors.New("cannot rotate while moving")
	}

	uc.tank.Speed = 32.0
	if uc.tank.Direction == direction {
		return errors.New("already facing this direction")
	}

	uc.tank.Direction = direction
	return nil
}

// StopPlayer останавливает игрока
func (uc *PlayerUseCases) StopPlayer(byCollision bool) error {
	uc.tank.Speed = 0

	if byCollision {
		uc.tank.WorldPosition.X = float64(int(uc.tank.WorldPosition.X))
		uc.tank.WorldPosition.Y = float64(int(uc.tank.WorldPosition.Y))
		return nil
	}

	// Выравниваем позицию по сетке
	switch uc.tank.Direction {
	case types.DirectionUp:
		uc.tank.WorldPosition.Y = float64(utils.RoundToEven(uc.tank.WorldPosition.Y, false))
	case types.DirectionDown:
		uc.tank.WorldPosition.Y = float64(utils.RoundToEven(uc.tank.WorldPosition.Y, true))
	case types.DirectionLeft:
		uc.tank.WorldPosition.X = float64(utils.RoundToEven(uc.tank.WorldPosition.X, false))
	case types.DirectionRight:
		uc.tank.WorldPosition.X = float64(utils.RoundToEven(uc.tank.WorldPosition.X, true))
	}

	return nil
}

// MovePlayer перемещает игрока в указанном направлении
func (uc *PlayerUseCases) MovePlayer(direction types.Direction, dt float64) error {
	delta := uc.tank.Speed * dt

	switch uc.tank.Direction {
	case types.DirectionUp:
		uc.tank.WorldPosition.Y -= delta
	case types.DirectionDown:
		uc.tank.WorldPosition.Y += delta
	case types.DirectionLeft:
		uc.tank.WorldPosition.X -= delta
	case types.DirectionRight:
		uc.tank.WorldPosition.X += delta
	}

	return nil
}
