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
	tilesUseCases *TilesUseCases    // Для всех анимаций (спавн, взрыв, танк)
	playerSpawner types.Position    // Координаты спавна игрока
	playerTank    *types.TankEntity // Указатель на танк игрока
	enemyTank     *types.TankEntity // Указатель на танк врага (если используется для врага)
	direction     types.Direction   // Направление танка по умолчанию
}

// NewTankUseCases создает новый экземпляр TankUseCases
func NewTankUseCases(
	tanksRepo game.ITanksRepository,
	tilesUseCases *TilesUseCases,
	playerSpawner types.Position,
) *TankUseCases {
	return &TankUseCases{
		tanksRepo:     tanksRepo,
		tilesUseCases: tilesUseCases,
		playerSpawner: playerSpawner,
		playerTank:    nil,
		enemyTank:     nil,
		direction:     types.DirectionUp, // Направление по умолчанию
	}
}

// StartTankSpawn создает танк и запускает процесс спавна с анимацией
// Возвращает созданный танк и анимацию спавна
func (uc *TankUseCases) StartTankSpawn(
	position types.Position,
) (*types.TankEntity, *types.TileAnimationEntity, error) {
	// Создаем анимацию спавна
	spawnAnimation, err := uc.tilesUseCases.CreateSpawnAnimation()
	if err != nil {
		return nil, nil, err
	}

	// Создаем анимацию танка
	// Создаем временный TilesUseCases для работы с анимацией танка
	tankTilesUseCases := NewTilesUseCases(uc.tilesUseCases.tilesRepository)
	tankAnimation, err := tankTilesUseCases.CreateAnimationTile("base_tank")
	if err != nil {
		return nil, nil, err
	}
	uc.tilesUseCases.AddAnimation(tankAnimation)

	// Создаем танк
	tank := &types.TankEntity{
		AnimationGetter: spawnAnimation, // Сначала устанавливаем анимацию спавна
		Position:        position,
		Speed:           0,
		Direction:       uc.direction,
		State:           types.TankStateSpawning,
		SpawnedAt:       0,
		Altitude:        types.SURFACE,
	}

	// Добавляем танк в репозиторий
	uc.tanksRepo.AddTank(tank)

	// Запускаем анимацию спавна
	uc.tilesUseCases.StartAnimation(spawnAnimation)

	return tank, spawnAnimation, nil
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

// SetExplosionAnimation устанавливает и запускает анимацию взрыва для танка
func (uc *TankUseCases) SetExplosionAnimation(tank *types.TankEntity) error {
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

// GetPlayerTank возвращает танк игрока
func (uc *TankUseCases) GetPlayerTank() (*types.TankEntity, error) {
	if uc.playerTank == nil {
		return nil, errors.New("tank not created yet")
	}
	return uc.playerTank, nil
}

// SetPlayerTank устанавливает танк игрока
func (uc *TankUseCases) SetPlayerTank(tank *types.TankEntity) {
	uc.playerTank = tank
}

// GetPlayerSpawner возвращает координаты спавна игрока
func (uc *TankUseCases) GetPlayerSpawner() types.Position {
	return uc.playerSpawner
}

// GetEnemyTank возвращает танк врага
func (uc *TankUseCases) GetEnemyTank() *types.TankEntity {
	return uc.enemyTank
}

// SetEnemyTank устанавливает танк врага
func (uc *TankUseCases) SetEnemyTank(tank *types.TankEntity) {
	uc.enemyTank = tank
}

// UpdateAnimations обновляет все анимации
func (uc *TankUseCases) UpdateAnimations() {
	uc.tilesUseCases.UpdateAnimations()
}

// UpdatePlayerSpawn обновляет процесс спавна игрока
func (uc *TankUseCases) UpdatePlayerSpawn(currentTime float64) {
	uc.UpdateTankSpawn(uc.playerTank, currentTime)
}

// UpdateEnemySpawn обновляет процесс спавна врага
func (uc *TankUseCases) UpdateEnemySpawn(currentTime float64) {
	uc.UpdateTankSpawn(uc.enemyTank, currentTime)
}

// UpdateTankSpawn обновляет процесс спавна танка (для игрока или врага)
func (uc *TankUseCases) UpdateTankSpawn(tank *types.TankEntity, currentTime float64) {
	if tank == nil {
		return
	}

	// Если танк еще не заспавнен, проверяем анимацию спавна
	if tank.State == types.TankStateSpawning {
		if anim, ok := tank.AnimationGetter.(*types.TileAnimationEntity); ok {
			if anim.IsFinished() {
				// Завершаем спавн - устанавливаем анимацию танка
				tankTilesUseCases := NewTilesUseCases(uc.tilesUseCases.tilesRepository)
				tankAnimation, err := tankTilesUseCases.CreateAnimationTile("base_tank")
				if err == nil {
					tank.AnimationGetter = tankAnimation
					uc.tilesUseCases.AddAnimation(tankAnimation)
				}

				tank.State = types.TankStateStopped
				tank.SpawnedAt = currentTime
			}
		}
	}
}
