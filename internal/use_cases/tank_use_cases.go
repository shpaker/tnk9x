package use_cases

import (
	"errors"

	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/utils"
)

// TankUseCases предоставляет базовые операции для работы с танками
type TankUseCases struct {
	tanksRepo     game.ITanksRepository
	tilesUseCases *TilesUseCases   // Для всех анимаций (спавн, взрыв, танк)
	spawnAt       types.Position   // Координаты спавна игрока
	tank          types.TankEntity // Танк
}

// NewTankUseCases создает новый экземпляр TankUseCases
func NewTankUseCases(
	tanksRepo game.ITanksRepository,
	tilesUseCases *TilesUseCases,
	spawnAt types.Position,
	direction types.Direction,
) *TankUseCases {
	// Создаем танк
	tank := &types.TankEntity{
		Position:  spawnAt,
		Speed:     0,
		Direction: direction,
		State:     types.TankStateSpawning,
		Altitude:  types.SURFACE,
	}

	uc := &TankUseCases{
		tanksRepo:     tanksRepo,
		tilesUseCases: tilesUseCases,
		spawnAt:       spawnAt,
		tank:          *tank,
	}
	uc.tanksRepo.AddTank(&uc.tank)
	return uc
}

// StartSpawn создает танк и запускает процесс спавна с анимацией
// Использует танк, переданный в конструктор
func (uc *TankUseCases) StartSpawn() error {

	// Создаем анимацию спавна
	spawnAnimation, err := uc.tilesUseCases.CreateSpawnAnimation()
	if err != nil {
		return err
	}

	// Устанавливаем анимацию спавна танку
	uc.tank.AnimationGetter = spawnAnimation
	uc.tank.State = types.TankStateSpawning

	// Запускаем анимацию спавна
	uc.tilesUseCases.StartAnimation(spawnAnimation)

	return nil
}

// Rotate поворачивает танк в указанном направлении
func (uc *TankUseCases) Rotate(
	direction types.Direction,
) error {
	if !uc.tank.IsActive() {
		return errors.New("tank is not active")
	}

	if uc.tank.Speed != 0 {
		return errors.New("cannot rotate while moving")
	}

	uc.tank.Direction = direction
	return nil
}

// Move запускает движение танка (устанавливает скорость)
func (uc *TankUseCases) Move() error {
	if !uc.tank.IsActive() {
		return errors.New("tank is not active")
	}

	uc.tank.Speed = 32.0
	return nil
}

// StopTank останавливает танк
func (uc *TankUseCases) StopTank(
	byCollision bool,
) error {
	if !uc.tank.IsActive() {
		return errors.New("tank is not active")
	}

	uc.tank.Speed = 0

	if byCollision {
		// Округляем координаты до ближайшего кратного 4
		uc.tank.Position.X = utils.RoundToNearestMultipleOf4(uc.tank.Position.X)
		uc.tank.Position.Y = utils.RoundToNearestMultipleOf4(uc.tank.Position.Y)
		return nil
	}

	// Выравниваем позицию по сетке
	switch uc.tank.Direction {
	case types.DirectionUp:
		uc.tank.Position.Y = float64(utils.RoundToEven(uc.tank.Position.Y, false))
	case types.DirectionDown:
		uc.tank.Position.Y = float64(utils.RoundToEven(uc.tank.Position.Y, true))
	case types.DirectionLeft:
		uc.tank.Position.X = float64(utils.RoundToEven(uc.tank.Position.X, false))
	case types.DirectionRight:
		uc.tank.Position.X = float64(utils.RoundToEven(uc.tank.Position.X, true))
	}

	return nil
}

// Update обновляет состояние танка (движение)
func (uc *TankUseCases) Update(
	dt float64,
) error {
	if !uc.tank.IsActive() {
		return errors.New("tank is not active")
	}

	delta := uc.tank.Speed * dt

	switch uc.tank.Direction {
	case types.DirectionUp:
		uc.tank.Position.Y -= delta
	case types.DirectionDown:
		uc.tank.Position.Y += delta
	case types.DirectionLeft:
		uc.tank.Position.X -= delta
	case types.DirectionRight:
		uc.tank.Position.X += delta
	}

	return nil
}

// StartExplosion устанавливает и запускает анимацию взрыва для танка
func (uc *TankUseCases) StartExplosion(
	tank *types.TankEntity,
) error {
	if tank == nil {
		return errors.New("tank is nil")
	}

	explosionAnim, err := uc.tilesUseCases.CreateExplosionAnimation()
	if err != nil {
		return err
	}

	// Устанавливаем анимацию взрыва танку
	tank.AnimationGetter = explosionAnim
	tank.State = types.TankStateExploding

	// Запускаем анимацию
	uc.tilesUseCases.StartAnimation(explosionAnim)

	return nil
}

// GetTank возвращает танк
func (uc *TankUseCases) GetTank() *types.TankEntity {
	return &uc.tank
}

// IsActive возвращает true если танк активен
func (uc *TankUseCases) IsActive() bool {
	return uc.tank.IsActive()
}

// IsStopped возвращает true если танк остановлен
func (uc *TankUseCases) IsStopped() bool {
	return uc.tank.Speed == 0
}

// IsSpawnFinished проверяет и обновляет процесс спавна танка
func (uc *TankUseCases) IsSpawnFinished(currentTime float64) {
	// Если танк еще не заспавнен, проверяем анимацию спавна
	if uc.tank.State == types.TankStateSpawning {
		if anim, ok := uc.tank.AnimationGetter.(*types.TileAnimationEntity); ok {
			if anim.IsFinished() {
				// Завершаем спавн - устанавливаем анимацию танка
				tankTilesUseCases := NewTilesUseCases(uc.tilesUseCases.tilesRepository)
				tankAnimation, err := tankTilesUseCases.CreateAnimationTile("base_tank")
				if err == nil {
					uc.tank.AnimationGetter = tankAnimation
					uc.tilesUseCases.AddAnimation(tankAnimation)
				}

				uc.tank.State = types.TankStateStopped
				uc.tank.SpawnedAt = currentTime
			}
		}
	}
}

// IsExplosionFinished проверяет завершение анимации взрыва танка
func (uc *TankUseCases) IsExplosionFinished() {
	// Если танк взрывается, проверяем завершение анимации взрыва
	if uc.tank.State == types.TankStateExploding {
		if anim, ok := uc.tank.AnimationGetter.(*types.TileAnimationEntity); ok {
			if anim.IsFinished() {
				// Завершаем взрыв - устанавливаем состояние взорванного танка
				uc.tank.State = types.TankStateExploded
			}
		}
	}
}
