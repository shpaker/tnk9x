package services

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ITankBrakingService = (*TankBrakingService)(nil)

type TankBrakingService struct{}

func NewTankBrakingService() *TankBrakingService {
	return &TankBrakingService{}
}

type brakingMovementContext struct {
	currentCoord      *float64
	targetMultipleOf4 float64
	isMovingForward   bool
}

func (s *TankBrakingService) HandleBrakingState(
	tank *types.TankEntity,
	dt float64,
	onIce bool,
) error {
	if onIce && tank.SlideTarget == nil {
		s.latchSlideTarget(tank)
	}

	ctx := s.getBrakingMovementContext(tank)

	s.moveTowardsTarget(tank, ctx, dt)

	return nil
}

// latchSlideTarget фиксирует цель скольжения один раз при входе
// в торможение на льду: обычная точка остановки плюс ещё 4px
// по инерции. Пересчёт каждый тик сдвигал бы цель бесконечно.
func (s *TankBrakingService) latchSlideTarget(tank *types.TankEntity) {
	var target float64

	switch tank.Direction {
	case types.DirectionUp:
		target = s.getNearestMultipleOf4InDirection(tank.Position.Y, false) - 4
	case types.DirectionDown:
		target = s.getNearestMultipleOf4InDirection(tank.Position.Y, true) + 4
	case types.DirectionLeft:
		target = s.getNearestMultipleOf4InDirection(tank.Position.X, false) - 4
	case types.DirectionRight:
		target = s.getNearestMultipleOf4InDirection(tank.Position.X, true) + 4
	default:
		return
	}

	tank.SlideTarget = &target
}

func (s *TankBrakingService) getBrakingMovementContext(
	tank *types.TankEntity,
) brakingMovementContext {
	var currentCoord *float64
	var targetMultipleOf4 float64
	var isMovingForward bool

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

	if tank.SlideTarget != nil {
		targetMultipleOf4 = *tank.SlideTarget
	}

	return brakingMovementContext{
		currentCoord:      currentCoord,
		targetMultipleOf4: targetMultipleOf4,
		isMovingForward:   isMovingForward,
	}
}

func (s *TankBrakingService) moveTowardsTarget(
	tank *types.TankEntity,
	ctx brakingMovementContext,
	dt float64,
) {
	// Получаем скорость танка из спецификаций
	speed := float64(32.0) // Значение по умолчанию
	if tank.GetSpecs() != nil {
		speed = tank.GetSpecs().GetSpeed()
	}
	delta := speed * dt

	if ctx.isMovingForward {
		s.moveForwardToTarget(tank, ctx, delta)
	} else {
		s.moveBackwardToTarget(tank, ctx, delta)
	}
}

func (s *TankBrakingService) moveForwardToTarget(
	tank *types.TankEntity,
	ctx brakingMovementContext,
	delta float64,
) {
	if *ctx.currentCoord < ctx.targetMultipleOf4 {
		if *ctx.currentCoord+delta >= ctx.targetMultipleOf4 {
			*ctx.currentCoord = ctx.targetMultipleOf4
			s.completeBraking(tank)
		} else {
			*ctx.currentCoord += delta
		}
	} else if *ctx.currentCoord == ctx.targetMultipleOf4 {
		s.completeBraking(tank)
	} else {

		*ctx.currentCoord = ctx.targetMultipleOf4
		s.completeBraking(tank)
	}
}

func (s *TankBrakingService) moveBackwardToTarget(
	tank *types.TankEntity,
	ctx brakingMovementContext,
	delta float64,
) {
	if *ctx.currentCoord > ctx.targetMultipleOf4 {
		if *ctx.currentCoord-delta <= ctx.targetMultipleOf4 {
			*ctx.currentCoord = ctx.targetMultipleOf4
			s.completeBraking(tank)
		} else {
			*ctx.currentCoord -= delta
		}
	} else if *ctx.currentCoord == ctx.targetMultipleOf4 {
		s.completeBraking(tank)
	} else {

		*ctx.currentCoord = ctx.targetMultipleOf4
		s.completeBraking(tank)
	}
}

func (s *TankBrakingService) completeBraking(tank *types.TankEntity) {
	s.finishBraking(tank)
}

func (s *TankBrakingService) finishBraking(tank *types.TankEntity) {
	tank.SlideTarget = nil
	if tank.NextDirection != nil {
		tank.Direction = *tank.NextDirection
		tank.NextDirection = nil
		tank.State = types.TankStateMoving
	} else {
		tank.State = types.TankStateStopped
		tank.NextDirection = nil
	}
}

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

func (s *TankBrakingService) getNearestMultipleOf4InDirection(
	value float64,
	forward bool,
) float64 {
	lower := float64(int(value) / 4 * 4)
	upper := lower + 4

	if forward {

		if value <= lower {
			return lower
		}
		return upper
	} else {

		if value >= upper {
			return upper
		}
		return lower
	}
}
