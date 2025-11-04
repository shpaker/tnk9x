package services

import (
	"github.com/shpaker/gonflict/internal/types"
)

// TankBrakingService предоставляет логику торможения танка
type TankBrakingService struct{}

// NewTankBrakingService создает новый сервис торможения
func NewTankBrakingService() *TankBrakingService {
	return &TankBrakingService{}
}

// brakingMovementContext содержит контекст для движения танка к цели в состоянии Braking
type brakingMovementContext struct {
	currentCoord      *float64
	targetMultipleOf4 float64
	isMovingForward   bool
}

// HandleBrakingState обрабатывает движение танка в состоянии Braking
// Танк должен доехать до координаты кратной 4
func (s *TankBrakingService) HandleBrakingState(
	tank *types.TankEntity,
	dt float64,
) error {
	ctx := s.getBrakingMovementContext(tank)

	// Проверка: если координата больше кратного 4 на 0.5, возвращаем на 0.5 назад
	if s.checkAndHandleHalfStepBack(tank, ctx) {
		return nil
	}

	// Двигаемся к целевому кратному 4
	s.moveTowardsTarget(tank, ctx, dt)

	return nil
}

// getBrakingMovementContext определяет текущую координату, цель и направление движения
func (s *TankBrakingService) getBrakingMovementContext(
	tank *types.TankEntity,
) brakingMovementContext {
	var currentCoord *float64
	var targetMultipleOf4 float64
	var isMovingForward bool

	// Определяем текущую координату и целевое кратное 4
	// Целевое кратное 4 должно быть ближайшим в направлении движения
	switch tank.Direction {
	case types.DirectionUp:
		currentCoord = &tank.Position.Y
		targetMultipleOf4 = s.getNearestMultipleOf4InDirection(
			*currentCoord,
			false,
		)
		isMovingForward = false
	case types.DirectionDown:
		currentCoord = &tank.Position.Y
		targetMultipleOf4 = s.getNearestMultipleOf4InDirection(
			*currentCoord,
			true,
		)
		isMovingForward = true
	case types.DirectionLeft:
		currentCoord = &tank.Position.X
		targetMultipleOf4 = s.getNearestMultipleOf4InDirection(
			*currentCoord,
			false,
		)
		isMovingForward = false
	case types.DirectionRight:
		currentCoord = &tank.Position.X
		targetMultipleOf4 = s.getNearestMultipleOf4InDirection(
			*currentCoord,
			true,
		)
		isMovingForward = true
	}

	return brakingMovementContext{
		currentCoord:      currentCoord,
		targetMultipleOf4: targetMultipleOf4,
		isMovingForward:   isMovingForward,
	}
}

// checkAndHandleHalfStepBack проверяет и обрабатывает случай возврата на 0.5 назад
// Возвращает true, если обработка выполнена и танк остановлен
func (s *TankBrakingService) checkAndHandleHalfStepBack(
	tank *types.TankEntity,
	ctx brakingMovementContext,
) bool {
	diff := *ctx.currentCoord - ctx.targetMultipleOf4
	if diff > 0 && diff <= 0.5 {
		*ctx.currentCoord = ctx.targetMultipleOf4 - 0.5
		// Устанавливаем состояние остановки - это часть логики торможения
		tank.State = types.TankStateStopped
		tank.NextDirection = nil
		return true
	}
	return false
}

// moveTowardsTarget двигает танк к целевому кратному 4
func (s *TankBrakingService) moveTowardsTarget(
	tank *types.TankEntity,
	ctx brakingMovementContext,
	dt float64,
) {
	delta := tank.Speed * dt

	if ctx.isMovingForward {
		s.moveForwardToTarget(tank, ctx, delta)
	} else {
		s.moveBackwardToTarget(tank, ctx, delta)
	}
}

// moveForwardToTarget двигает танк вперед к цели
func (s *TankBrakingService) moveForwardToTarget(
	tank *types.TankEntity,
	ctx brakingMovementContext,
	delta float64,
) {
	if *ctx.currentCoord < ctx.targetMultipleOf4 {
		// Двигаемся вперед к цели
		if *ctx.currentCoord+delta >= ctx.targetMultipleOf4 {
			*ctx.currentCoord = ctx.targetMultipleOf4
			s.completeBraking(tank)
		} else {
			*ctx.currentCoord += delta
		}
	} else if *ctx.currentCoord == ctx.targetMultipleOf4 {
		// Достигли цели
		s.completeBraking(tank)
	} else {
		// Переехали цель - останавливаемся на текущем кратном 4
		*ctx.currentCoord = ctx.targetMultipleOf4
		s.completeBraking(tank)
	}
}

// moveBackwardToTarget двигает танк назад к цели
func (s *TankBrakingService) moveBackwardToTarget(
	tank *types.TankEntity,
	ctx brakingMovementContext,
	delta float64,
) {
	if *ctx.currentCoord > ctx.targetMultipleOf4 {
		// Двигаемся назад к цели
		if *ctx.currentCoord-delta <= ctx.targetMultipleOf4 {
			*ctx.currentCoord = ctx.targetMultipleOf4
			s.completeBraking(tank)
		} else {
			*ctx.currentCoord -= delta
		}
	} else if *ctx.currentCoord == ctx.targetMultipleOf4 {
		// Достигли цели
		s.completeBraking(tank)
	} else {
		// Переехали цель - останавливаемся на текущем кратном 4
		*ctx.currentCoord = ctx.targetMultipleOf4
		s.completeBraking(tank)
	}
}

// completeBraking завершает процесс торможения и обнуляет скорость
func (s *TankBrakingService) completeBraking(tank *types.TankEntity) {
	tank.Speed = 0
	s.finishBraking(tank)
}

// finishBraking завершает состояние Braking
func (s *TankBrakingService) finishBraking(tank *types.TankEntity) {
	// Если есть новое направление, меняем направление и начинаем движение
	if tank.NextDirection != nil {
		tank.Direction = *tank.NextDirection
		tank.NextDirection = nil
		tank.State = types.TankStateMoving
		tank.Speed = 32.0
	} else {
		// Иначе просто останавливаемся
		tank.State = types.TankStateStopped
		tank.NextDirection = nil
	}
}

// HandleRotateWhileBraking обрабатывает поворот танка в состоянии Braking
func (s *TankBrakingService) HandleRotateWhileBraking(
	tank *types.TankEntity,
	direction types.Direction,
) {
	if direction == tank.Direction {
		tank.NextDirection = nil
	} else {
		directionCopy := direction
		tank.NextDirection = &directionCopy
	}
}

// getNearestMultipleOf4InDirection возвращает ближайшее кратное 4 в направлении движения
// forward: true - больше текущего значения, false - меньше
func (s *TankBrakingService) getNearestMultipleOf4InDirection(
	value float64,
	forward bool,
) float64 {
	// Находим нижнее кратное 4
	lower := float64(int(value) / 4 * 4)
	upper := lower + 4

	if forward {
		// Движемся вперед - берем ближайшее кратное 4, которое >= текущего значения
		if value <= lower {
			return lower
		}
		return upper
	} else {
		// Движемся назад - берем ближайшее кратное 4, которое <= текущего значения
		if value >= upper {
			return upper
		}
		return lower
	}
}
