package use_cases

import (
	"errors"

	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/utils"
)

// TankUseCases реализация интерфейса TankUseCases
type TankUseCases struct {
	tilesetRepo       processed.ITilesetRepository
	animationUseCases IAnimationUseCases
	tank              types.TankEntity
	tankAnimation     *types.TileAnimationEntity
}

// NewTankUseCases создает новый экземпляр TankUseCases
func NewTankUseCases(
	tilesetRepo processed.ITilesetRepository,
	animationUseCases IAnimationUseCases,
) *TankUseCases {
	uc := &TankUseCases{
		tilesetRepo:       tilesetRepo,
		animationUseCases: animationUseCases,
	}

	// Создаем танк при инициализации
	tank, tankAnimation, err := uc.makeTank()
	if err != nil {
		panic(err)
	}
	uc.tank = tank
	uc.tankAnimation = tankAnimation

	return uc
}

// makeTank создает танк с начальными параметрами
func (uc *TankUseCases) makeTank() (types.TankEntity, *types.TileAnimationEntity, error) {
	// Получаем данные анимации для танка
	animationFrames, err := uc.tilesetRepo.GetAnimationData("base_tank")
	if err != nil {
		return types.TankEntity{}, nil, err
	}

	// Создаем TileAnimationEntity для танка
	tankAnimation := types.NewTileAnimationEntity(animationFrames)

	// Добавляем анимацию танка через AnimationUseCases
	uc.animationUseCases.AddAnimation(tankAnimation)

	// Анимация танка начинается остановленной
	// uc.animationUseCases.StartAnimation(tankAnimation) - убираем автоматический запуск

	// Создаем игрока с начальными параметрами
	spawnPosition := types.Position{X: 4 * TankSpriteSize, Y: 12 * TankSpriteSize}

	player := types.TankEntity{
		AnimationGetter: tankAnimation,
		SpawnPosition:   spawnPosition,
		WorldPosition: types.Position{
			X: spawnPosition.X,
			Y: spawnPosition.Y,
		},
		Speed:     0,
		Direction: types.DirectionUp,
	}

	return player, tankAnimation, nil
}

// GetTank возвращает данные танка
func (uc *TankUseCases) GetTank() (*types.TankEntity, error) {
	return &uc.tank, nil
}

// GetDirection возвращает текущее направление танка
func (uc *TankUseCases) GetDirection() types.Direction {
	return uc.tank.Direction
}

// RotateTank поворачивает танк в указанном направлении
func (uc *TankUseCases) RotateTank(direction types.Direction) error {
	if uc.tank.Speed != 0 {
		return errors.New("cannot rotate while moving")
	}

	uc.tank.Speed = 32.0
	if uc.tank.Direction == direction {
		return errors.New("already facing this direction")
	}

	uc.tank.Direction = direction

	// Запускаем анимацию при повороте
	uc.animationUseCases.StartAnimation(uc.tankAnimation)

	return nil
}

// StopTank останавливает танк
func (uc *TankUseCases) StopTank(byCollision bool) error {
	uc.tank.Speed = 0

	// Останавливаем анимацию танка
	uc.animationUseCases.StopAnimation(uc.tankAnimation)

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

// MoveTank перемещает танк в указанном направлении
func (uc *TankUseCases) MoveTank(
	direction types.Direction,
	dt float64,
) error {
	delta := uc.tank.Speed * dt

	// Если танк движется, запускаем анимацию
	if delta > 0 {
		uc.animationUseCases.StartAnimation(uc.tankAnimation)
	}

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
