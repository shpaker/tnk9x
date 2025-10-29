package services

import (
	"log"

	"github.com/shpaker/gonflict/internal/types"
)

// TankBrakingService предоставляет логику торможения танка
type TankBrakingService struct {
	tank *types.TankEntity
}

// NewTankBrakingService создает новый сервис торможения
func NewTankBrakingService(
	tank *types.TankEntity,
) *TankBrakingService {
	return &TankBrakingService{
		tank: tank,
	}
}

// brakingMovementContext содержит контекст для движения танка к цели в состоянии Braking
type brakingMovementContext struct {
	currentCoord      *float64
	targetMultipleOf4 float64
	isMovingForward   bool
}

// HandleBrakingState обрабатывает движение танка в состоянии Braking
// Танк должен доехать до координаты кратной 4
func (s *TankBrakingService) HandleBrakingState(dt float64) error {
	ctx := s.getBrakingMovementContext()

	// Проверка: если координата больше кратного 4 на 0.5, возвращаем на 0.5 назад
	if s.checkAndHandleHalfStepBack(ctx) {
		log.Printf(
			"DEBUG: Tank braking - half step back (%.2f, %.2f) target=%.2f",
			s.tank.Position.X,
			s.tank.Position.Y,
			ctx.targetMultipleOf4,
		)
		return nil
	}

	// Двигаемся к целевому кратному 4
	s.moveTowardsTarget(ctx, dt)

	log.Printf(
		"DEBUG: Tank braking position (%.2f, %.2f) target=%.2f currentCoord=%.2f state=%d direction=%d speed=%.2f",
		s.tank.Position.X,
		s.tank.Position.Y,
		ctx.targetMultipleOf4,
		*ctx.currentCoord,
		s.tank.State,
		s.tank.Direction,
		s.tank.Speed,
	)

	return nil
}

// getBrakingMovementContext определяет текущую координату, цель и направление движения
func (s *TankBrakingService) getBrakingMovementContext() brakingMovementContext {
	var currentCoord *float64
	var targetMultipleOf4 float64
	var isMovingForward bool

	// Определяем текущую координату и целевое кратное 4
	// Целевое кратное 4 должно быть ближайшим в направлении движения
	switch s.tank.Direction {
	case types.DirectionUp:
		currentCoord = &s.tank.Position.Y
		targetMultipleOf4 = s.getNearestMultipleOf4InDirection(
			*currentCoord,
			false,
		)
		isMovingForward = false
	case types.DirectionDown:
		currentCoord = &s.tank.Position.Y
		targetMultipleOf4 = s.getNearestMultipleOf4InDirection(
			*currentCoord,
			true,
		)
		isMovingForward = true
	case types.DirectionLeft:
		currentCoord = &s.tank.Position.X
		targetMultipleOf4 = s.getNearestMultipleOf4InDirection(
			*currentCoord,
			false,
		)
		isMovingForward = false
	case types.DirectionRight:
		currentCoord = &s.tank.Position.X
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
	ctx brakingMovementContext,
) bool {
	diff := *ctx.currentCoord - ctx.targetMultipleOf4
	if diff > 0 && diff <= 0.5 {
		*ctx.currentCoord = ctx.targetMultipleOf4 - 0.5
		s.tank.Speed = 0
		s.tank.State = types.TankStateStopped
		s.tank.NextDirection = nil
		return true
	}
	return false
}

// moveTowardsTarget двигает танк к целевому кратному 4
func (s *TankBrakingService) moveTowardsTarget(
	ctx brakingMovementContext,
	dt float64,
) {
	delta := s.tank.Speed * dt

	if ctx.isMovingForward {
		s.moveForwardToTarget(ctx, delta)
	} else {
		s.moveBackwardToTarget(ctx, delta)
	}
}

// moveForwardToTarget двигает танк вперед к цели
func (s *TankBrakingService) moveForwardToTarget(
	ctx brakingMovementContext,
	delta float64,
) {
	if *ctx.currentCoord < ctx.targetMultipleOf4 {
		// Двигаемся вперед к цели
		if *ctx.currentCoord+delta >= ctx.targetMultipleOf4 {
			*ctx.currentCoord = ctx.targetMultipleOf4
			s.completeBraking()
		} else {
			*ctx.currentCoord += delta
		}
	} else if *ctx.currentCoord == ctx.targetMultipleOf4 {
		// Достигли цели
		s.completeBraking()
	} else {
		// Переехали цель - останавливаемся на текущем кратном 4
		*ctx.currentCoord = ctx.targetMultipleOf4
		s.completeBraking()
	}
}

// moveBackwardToTarget двигает танк назад к цели
func (s *TankBrakingService) moveBackwardToTarget(
	ctx brakingMovementContext,
	delta float64,
) {
	if *ctx.currentCoord > ctx.targetMultipleOf4 {
		// Двигаемся назад к цели
		if *ctx.currentCoord-delta <= ctx.targetMultipleOf4 {
			*ctx.currentCoord = ctx.targetMultipleOf4
			s.completeBraking()
		} else {
			*ctx.currentCoord -= delta
		}
	} else if *ctx.currentCoord == ctx.targetMultipleOf4 {
		// Достигли цели
		s.completeBraking()
	} else {
		// Переехали цель - останавливаемся на текущем кратном 4
		*ctx.currentCoord = ctx.targetMultipleOf4
		s.completeBraking()
	}
}

// completeBraking завершает процесс торможения и обнуляет скорость
func (s *TankBrakingService) completeBraking() {
	s.tank.Speed = 0
	s.finishBraking()
}

// finishBraking завершает состояние Braking
func (s *TankBrakingService) finishBraking() {
	// Если есть новое направление, меняем направление и начинаем движение
	if s.tank.NextDirection != nil {
		oldDir := s.tank.Direction
		s.tank.Direction = *s.tank.NextDirection
		s.tank.NextDirection = nil
		s.tank.State = types.TankStateMoving
		s.tank.Speed = 32.0
		log.Printf(
			"DEBUG: Tank finished braking, rotated %d->%d position (%.2f, %.2f) state=Moving",
			oldDir,
			s.tank.Direction,
			s.tank.Position.X,
			s.tank.Position.Y,
		)
	} else {
		// Иначе просто останавливаемся
		s.tank.State = types.TankStateStopped
		s.tank.NextDirection = nil
		log.Printf("DEBUG: Tank finished braking, stopped at position (%.2f, %.2f) state=Stopped",
			s.tank.Position.X, s.tank.Position.Y)
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
