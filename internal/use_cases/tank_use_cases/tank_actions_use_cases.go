package tank_use_cases

import (
	"errors"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ITankActionsUseCases = (*TankActionsUseCases)(nil)

type TankActionsUseCases struct {
	brakingService interfaces.ITankBrakingService
	bulletUseCases interfaces.IBulletUseCases
	commonUseCases interfaces.ITankCommonUseCases
	renderUseCases interfaces.IRenderUseCases
	mapUseCases    interfaces.IMapUseCases
	soundUseCases  interfaces.ISoundUseCases
}

func NewTankActionsUseCases(
	brakingService interfaces.ITankBrakingService,
	bulletUseCases interfaces.IBulletUseCases,
	commonUseCases interfaces.ITankCommonUseCases,
	renderUseCases interfaces.IRenderUseCases,
	mapUseCases interfaces.IMapUseCases,
	soundUseCases interfaces.ISoundUseCases,
) *TankActionsUseCases {
	return &TankActionsUseCases{
		brakingService: brakingService,
		bulletUseCases: bulletUseCases,
		commonUseCases: commonUseCases,
		renderUseCases: renderUseCases,
		mapUseCases:    mapUseCases,
		soundUseCases:  soundUseCases,
	}
}

func (uc *TankActionsUseCases) Update(
	tank *types.TankEntity,
	dt float64,
) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}

	return uc.commonUseCases.Update(tank, dt)
}

func (uc *TankActionsUseCases) Rotate(
	tank *types.TankEntity,
	direction types.Direction,
) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}
	if tank.IsFrozen() {
		return nil
	}
	if direction == tank.Direction {
		return nil
	}

	if tank.State == types.TankStateBraking {
		uc.brakingService.HandleRotateWhileBraking(tank, direction)

		uc.renderUseCases.UpdateTankAnimation(tank)
		return nil
	}

	if tank.State == types.TankStateStopped {
		tank.Direction = direction
		uc.renderUseCases.UpdateTankAnimation(tank)
		return nil
	}

	directionCopy := direction
	tank.NextDirection = &directionCopy
	tank.State = types.TankStateBraking

	uc.renderUseCases.UpdateTankAnimation(tank)
	return nil
}

func (uc *TankActionsUseCases) Move(tank *types.TankEntity) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}
	if tank.IsFrozen() {
		return nil
	}

	if tank.State == types.TankStateBraking {
		if tank.NextDirection == nil {
			tank.SlideTarget = nil
			tank.State = types.TankStateMoving
		}
		return nil
	}

	tank.State = types.TankStateMoving
	return nil
}

func (uc *TankActionsUseCases) Stop(tank *types.TankEntity, byCollision bool) {
	if !tank.IsActive() {
		return
	}
	tank.NextDirection = nil
	if byCollision {
		uc.handleStopByCollision(tank)
		return
	}

	tank.State = types.TankStateBraking
}

func (uc *TankActionsUseCases) Shoot(tank *types.TankEntity) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}
	if tank.IsFrozen() {
		return nil
	}
	if !tank.IsEnemy() {
		uc.soundUseCases.RequestSound(types.SoundIDFire, false)
	}
	return uc.bulletUseCases.ShootBullet(tank)
}

func (uc *TankActionsUseCases) ApplyDecision(
	tank *types.TankEntity,
	decision types.EnemyAIDecision,
) {
	if tank.IsStopped() {
		_ = uc.Rotate(tank, decision.Direction)
		_ = uc.Move(tank)
	}
}

func (uc *TankActionsUseCases) SetMinXPosition(tank *types.TankEntity) {
	tank.Position.X = 0
	uc.stopSlideAtBoundary(tank)
}

func (uc *TankActionsUseCases) SetMaxXPosition(tank *types.TankEntity) {
	mapSizePx := uc.mapUseCases.GetSizePx()
	maxX := float64(mapSizePx.Width - tank.Size.Width)
	tank.Position.X = maxX
	uc.stopSlideAtBoundary(tank)
}

func (uc *TankActionsUseCases) SetMinYPosition(tank *types.TankEntity) {
	tank.Position.Y = 0
	uc.stopSlideAtBoundary(tank)
}

func (uc *TankActionsUseCases) SetMaxYPosition(tank *types.TankEntity) {
	mapSizePx := uc.mapUseCases.GetSizePx()
	maxY := float64(mapSizePx.Height - tank.Size.Height)
	tank.Position.Y = maxY
	uc.stopSlideAtBoundary(tank)
}

// stopSlideAtBoundary прерывает скольжение у края карты: танк клампится
// каждый тик, зафиксированная цель недостижима — без сброса он навсегда
// останется в состоянии торможения
func (uc *TankActionsUseCases) stopSlideAtBoundary(tank *types.TankEntity) {
	if tank.SlideTarget == nil {
		return
	}
	tank.SlideTarget = nil
	tank.State = types.TankStateStopped
}

func (uc *TankActionsUseCases) handleStopByCollision(tank *types.TankEntity) {
	// Позиция уже разрешена коллизией (вплотную/откат) — округление
	// сдвигало бы танк с места контакта и порождало дрожание
	tank.SlideTarget = nil
	tank.State = types.TankStateStopped
}
