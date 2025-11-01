package use_cases

import (
	"errors"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// TankActionsUseCases отвечает за действия танка (движение и боевые действия)
type TankActionsUseCases struct {
	brakingService    interfaces.ITankBrakingService
	coordinateService interfaces.ICoordinateService
	bulletUseCases    interfaces.IBulletUseCases
	commonUseCases    interfaces.ITankCommonUseCases // Для управления анимацией через TankCommonUseCases
}

// NewTankActionsUseCases создает новый экземпляр TankActionsUseCases
func NewTankActionsUseCases(
	brakingService interfaces.ITankBrakingService,
	coordinateService interfaces.ICoordinateService,
	bulletUseCases interfaces.IBulletUseCases,
	commonUseCases interfaces.ITankCommonUseCases,
) *TankActionsUseCases {
	return &TankActionsUseCases{
		brakingService:    brakingService,
		coordinateService: coordinateService,
		bulletUseCases:    bulletUseCases,
		commonUseCases:    commonUseCases,
	}
}

// Update обновляет состояние танка (движение)
func (uc *TankActionsUseCases) Update(
	tank *types.TankEntity,
	dt float64,
) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}

	// Делегируем обновление в TankCommonUseCases, который управляет анимацией
	return uc.commonUseCases.Update(tank, dt)
}

// Rotate поворачивает танк в указанном направлении
func (uc *TankActionsUseCases) Rotate(
	tank *types.TankEntity,
	direction types.Direction,
) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}
	if direction == tank.Direction {
		return nil
	}

	// Если танк в состоянии Braking, запоминаем новое направление
	if tank.State == types.TankStateBraking {
		uc.brakingService.HandleRotateWhileBraking(tank, direction)
		return nil
	}

	if tank.State == types.TankStateStopped {
		tank.Direction = direction
		return nil
	}

	// Если танк в состоянии Moving, переводим в Braking и запоминаем новое направление
	directionCopy := direction
	tank.NextDirection = &directionCopy
	tank.State = types.TankStateBraking
	return nil
}

// Move запускает движение танка (устанавливает скорость)
func (uc *TankActionsUseCases) Move(tank *types.TankEntity) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}

	// Если танк в состоянии Braking, нужно доехать до кратного 4
	if tank.State == types.TankStateBraking {
		if tank.NextDirection == nil {
			tank.State = types.TankStateMoving
		}
		return nil
	}

	tank.Speed = 32.0
	tank.State = types.TankStateMoving
	return nil
}

// Stop останавливает танк
func (uc *TankActionsUseCases) Stop(tank *types.TankEntity, byCollision bool) {
	if !tank.IsActive() {
		return
	}
	tank.NextDirection = nil
	if byCollision {
		uc.handleStopByCollision(tank)
		return
	}

	// При отпускании клавиши - переходим в состояние Braking
	tank.State = types.TankStateBraking
	// Анимация продолжается при торможении (ничего не делаем, анимация уже идет)
}

// IsStopped возвращает true если танк остановлен
func (uc *TankActionsUseCases) IsStopped(tank *types.TankEntity) bool {
	return tank.Speed == 0
}

// Shoot создает пулю от танка
func (uc *TankActionsUseCases) Shoot(tank *types.TankEntity) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}
	return uc.bulletUseCases.ShootBullet(tank)
}

// ApplyDecision применяет решение AI к танку
func (uc *TankActionsUseCases) ApplyDecision(
	tank *types.TankEntity,
	decision types.EnemyAIDecision,
) {
	if uc.IsStopped(tank) {
		_ = uc.Rotate(tank, decision.Direction)
		_ = uc.Move(tank)
	}
}

// Приватные вспомогательные методы

// handleStopByCollision обрабатывает остановку танка при коллизии
func (uc *TankActionsUseCases) handleStopByCollision(tank *types.TankEntity) {
	tank.Speed = 0
	tank.Position.X = uc.coordinateService.RoundToNearestMultipleOf4(
		tank.Position.X,
	)
	tank.Position.Y = uc.coordinateService.RoundToNearestMultipleOf4(
		tank.Position.Y,
	)
	tank.State = types.TankStateStopped
	// Анимация будет синхронизирована автоматически в TankCommonUseCases.Update
}
