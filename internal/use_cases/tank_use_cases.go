package use_cases

import (
	"errors"
	"log"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// TankUseCases предоставляет базовые операции для работы с танками
type TankUseCases struct {
	tanksRepo         interfaces.ITanksRepository
	bulletUseCases    interfaces.IBulletUseCases     // Use Cases пуль
	tilesUseCases     *TilesUseCases                 // Для всех анимаций (спавн, взрыв, танк)
	spawnAt           types.Position                 // Координаты спавна игрока
	tank              types.TankEntity               // Танк
	animationGetter   types.IImageIDGetter           // Анимация танка
	brakingService    interfaces.ITankBrakingService // Сервис торможения танка
	coordinateService interfaces.ICoordinateService  // Сервис для работы с координатами
}

// ============================================================================
// КОНСТРУКТОР
// ============================================================================

// NewTankUseCases создает новый экземпляр TankUseCases
func NewTankUseCases(
	tanksRepo interfaces.ITanksRepository,
	bulletUseCases interfaces.IBulletUseCases,
	tilesUseCases *TilesUseCases,
	spawnAt types.Position,
	direction types.Direction,
	brakingService interfaces.ITankBrakingService,
	coordinateService interfaces.ICoordinateService,
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
		tanksRepo:         tanksRepo,
		bulletUseCases:    bulletUseCases,
		tilesUseCases:     tilesUseCases,
		spawnAt:           spawnAt,
		tank:              *tank,
		animationGetter:   nil,
		brakingService:    brakingService,
		coordinateService: coordinateService,
	}
	uc.tanksRepo.AddTank(&uc.tank)
	return uc
}

// ============================================================================
// УПРАВЛЕНИЕ СЕРВИСАМИ
// ============================================================================

// SetBrakingService устанавливает сервис торможения
func (uc *TankUseCases) SetBrakingService(
	brakingService interfaces.ITankBrakingService,
) {
	uc.brakingService = brakingService
}

// ============================================================================
// УПРАВЛЕНИЕ СОСТОЯНИЕМ ТАНКА
// ============================================================================

// StartSpawn создает танк и запускает процесс спавна с анимацией
// Использует танк, переданный в конструктор
func (uc *TankUseCases) StartSpawn() error {
	// Создаем анимацию спавна
	spawnAnimation, err := uc.tilesUseCases.CreateSpawnAnimation()
	if err != nil {
		return err
	}

	// Устанавливаем анимацию спавна танку
	uc.animationGetter = spawnAnimation
	uc.tank.State = types.TankStateSpawning
	uc.tank.Altitude = types.SURFACE

	// Запускаем анимацию спавна
	uc.tilesUseCases.StartAnimation(spawnAnimation)

	return nil
}

// StartExplosion устанавливает и запускает анимацию взрыва для танка
func (uc *TankUseCases) StartExplosion() error {
	explosionAnim, err := uc.tilesUseCases.CreateExplosionAnimation()
	if err != nil {
		return err
	}

	// Устанавливаем анимацию взрыва танку
	uc.animationGetter = explosionAnim
	uc.tank.State = types.TankStateExploding
	uc.tank.Altitude = types.AIR

	// Запускаем анимацию
	uc.tilesUseCases.StartAnimation(explosionAnim)

	return nil
}

// Update обновляет состояние танка (движение)
func (uc *TankUseCases) Update(
	dt float64,
) error {
	if !uc.tank.IsActive() {
		return errors.New("tank is not active")
	}

	// Обрабатываем состояние Braking отдельно
	if uc.tank.State == types.TankStateBraking {
		if uc.brakingService == nil {
			return errors.New("brakingService is not initialized")
		}
		return uc.brakingService.HandleBrakingState(&uc.tank, dt)
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

	log.Printf(
		"DEBUG: Tank position (%.2f, %.2f) state=%d direction=%d speed=%.2f",
		uc.tank.Position.X,
		uc.tank.Position.Y,
		uc.tank.State,
		uc.tank.Direction,
		uc.tank.Speed,
	)

	return nil
}

// ============================================================================
// УПРАВЛЕНИЕ ДВИЖЕНИЕМ
// ============================================================================

// Rotate поворачивает танк в указанном направлении
func (uc *TankUseCases) Rotate(
	direction types.Direction,
) error {
	if !uc.tank.IsActive() {
		return errors.New("tank is not active")
	}
	if direction == uc.tank.Direction {
		return nil
	}

	// Если танк в состоянии Braking, запоминаем новое направление
	// После доезжания до кратного 4 танк начнет движение в новом направлении
	if uc.tank.State == types.TankStateBraking {
		// Если направление совпадает с текущим, не нужно его менять
		if direction == uc.tank.Direction {
			uc.tank.NextDirection = nil
		} else {
			directionCopy := direction
			uc.tank.NextDirection = &directionCopy
		}
		return nil
	}

	if uc.tank.State == types.TankStateStopped {
		uc.tank.Direction = direction
		return nil
	}

	// Если танк в состоянии Moving, переводим в Braking и запоминаем новое направление
	// Танк сначала доедет до кратного 4, потом повернется
	directionCopy := direction
	uc.tank.NextDirection = &directionCopy
	uc.tank.State = types.TankStateBraking
	return nil
}

// Move запускает движение танка (устанавливает скорость)
func (uc *TankUseCases) Move() error {
	if !uc.tank.IsActive() {
		return errors.New("tank is not active")
	}

	// Если танк в состоянии Braking, нужно доехать до кратного 4
	// Новое направление уже установлено через Rotate, так что просто продолжаем
	if uc.tank.State == types.TankStateBraking {
		// Если NextDirection уже установлен, ничего не делаем - просто продолжаем доезжать
		// Если NextDirection не установлен, значит направление не менялось, просто продолжаем движение
		if uc.tank.NextDirection == nil {
			// Направление не менялось - останавливаем процесс остановки и продолжаем движение
			uc.tank.State = types.TankStateMoving
		}
		return nil
	}

	uc.tank.Speed = 32.0
	uc.tank.State = types.TankStateMoving
	return nil
}

// Stop останавливает танк
func (uc *TankUseCases) Stop(
	byCollision bool,
) {
	if !uc.tank.IsActive() {
		return
	}
	uc.tank.NextDirection = nil
	if byCollision {
		// При коллизии сразу останавливаем и округляем
		uc.tank.Speed = 0
		uc.tank.Position.X = uc.coordinateService.RoundToNearestMultipleOf4(
			uc.tank.Position.X,
		)
		uc.tank.Position.Y = uc.coordinateService.RoundToNearestMultipleOf4(
			uc.tank.Position.Y,
		)
		uc.tank.State = types.TankStateStopped
		return
	}

	// При отпускании клавиши - переходим в состояние Braking
	// Танк будет доезжать до кратного 4
	uc.tank.State = types.TankStateBraking
}

// ============================================================================
// СТРЕЛЬБА
// ============================================================================

// Shoot создает пулю от танка
func (uc *TankUseCases) Shoot() error {
	// Проверяем, активен ли танк
	if !uc.tank.IsActive() {
		return errors.New("tank is not active")
	}

	// Используем BulletUseCases для создания пули
	return uc.bulletUseCases.ShootBullet(&uc.tank)
}

// ============================================================================
// ПРОВЕРКИ СОСТОЯНИЯ
// ============================================================================

// IsActive возвращает true если танк активен
func (uc *TankUseCases) IsActive() bool {
	return uc.tank.IsActive()
}

// IsStopped возвращает true если танк остановлен
func (uc *TankUseCases) IsStopped() bool {
	return uc.tank.Speed == 0
}

// ============================================================================
// ПОЛУЧЕНИЕ ДАННЫХ
// ============================================================================

// GetTank возвращает танк
func (uc *TankUseCases) GetTank() *types.TankEntity {
	return &uc.tank
}

// GetImageID возвращает ID изображения танка
func (uc *TankUseCases) GetImageID() (string, error) {
	if uc.animationGetter == nil {
		return "", errors.New("AnimationGetter is nil")
	}
	return uc.animationGetter.GetImageID()
}

// GetAnimationGetter возвращает AnimationGetter для доступа к offset
func (uc *TankUseCases) GetAnimationGetter() types.IImageIDGetter {
	return uc.animationGetter
}

// ============================================================================
// ПРОВЕРКИ АНИМАЦИЙ
// ============================================================================

// IsSpawnFinished проверяет и обновляет процесс спавна танка
func (uc *TankUseCases) IsSpawnFinished(currentTime float64) {
	// Если танк еще не заспавнен, проверяем анимацию спавна
	if uc.tank.State == types.TankStateSpawning {
		if anim, ok := uc.animationGetter.(*types.TileAnimationEntity); ok {
			if anim.IsFinished() {
				// Завершаем спавн - устанавливаем анимацию танка
				// Используем существующий tilesUseCases для создания анимации
				tankAnimation, err := uc.tilesUseCases.CreateAnimationTile(
					"base_tank",
				)
				if err == nil {
					uc.animationGetter = tankAnimation
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
		if anim, ok := uc.animationGetter.(*types.TileAnimationEntity); ok {
			if anim.IsFinished() {
				// Завершаем взрыв - устанавливаем состояние взорванного танка
				uc.tank.State = types.TankStateExploded
			}
		}
	}
}
