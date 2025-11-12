package services

import (
	"github.com/shpaker/gonflict/internal/types"
)

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
) error {
	ctx := s.getBrakingMovementContext(tank)

	if s.checkAndHandleHalfStepBack(tank, ctx) {
		return nil
	}

	s.moveTowardsTarget(tank, ctx, dt)

	return nil
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

	return brakingMovementContext{
		currentCoord:      currentCoord,
		targetMultipleOf4: targetMultipleOf4,
		isMovingForward:   isMovingForward,
	}
}

func (s *TankBrakingService) checkAndHandleHalfStepBack(
	tank *types.TankEntity,
	ctx brakingMovementContext,
) bool {
	diff := *ctx.currentCoord - ctx.targetMultipleOf4
	if diff > 0 && diff <= 0.5 {
		*ctx.currentCoord = ctx.targetMultipleOf4 - 0.5

		tank.State = types.TankStateStopped
		tank.NextDirection = nil
		return true
	}
	return false
}

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
	tank.Speed = 0
	s.finishBraking(tank)
}

func (s *TankBrakingService) finishBraking(tank *types.TankEntity) {
	if tank.NextDirection != nil {
		tank.Direction = *tank.NextDirection
		tank.NextDirection = nil
		tank.State = types.TankStateMoving
		tank.Speed = 32.0
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
