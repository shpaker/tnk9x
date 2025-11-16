package tank_use_cases

import (
	"errors"

	"github.com/shpaker/tnk25/internal/interfaces"
	"github.com/shpaker/tnk25/internal/types"
	"github.com/shpaker/tnk25/internal/use_cases"
)

type TankActionsUseCases struct {
	brakingService    interfaces.ITankBrakingService
	coordinateService interfaces.ICoordinateService
	bulletUseCases    interfaces.IBulletUseCases
	commonUseCases    interfaces.ITankCommonUseCases
	renderUseCases    interfaces.IRenderUseCases
	mapUseCases       interfaces.IMapUseCases
	soundUseCases     *use_cases.SoundUseCases
	specsUseCases     interfaces.ISpecsUseCases
}

func NewTankActionsUseCases(
	brakingService interfaces.ITankBrakingService,
	coordinateService interfaces.ICoordinateService,
	bulletUseCases interfaces.IBulletUseCases,
	commonUseCases interfaces.ITankCommonUseCases,
	renderUseCases interfaces.IRenderUseCases,
	mapUseCases interfaces.IMapUseCases,
	soundUseCases *use_cases.SoundUseCases,
	specsUseCases interfaces.ISpecsUseCases,
) *TankActionsUseCases {
	return &TankActionsUseCases{
		brakingService:    brakingService,
		coordinateService: coordinateService,
		bulletUseCases:    bulletUseCases,
		commonUseCases:    commonUseCases,
		renderUseCases:    renderUseCases,
		mapUseCases:       mapUseCases,
		soundUseCases:     soundUseCases,
		specsUseCases:     specsUseCases,
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
	if direction == tank.Direction {
		return nil
	}

	if tank.State == types.TankStateBraking {
		uc.brakingService.HandleRotateWhileBraking(tank, direction)

		if uc.renderUseCases != nil {
			uc.renderUseCases.UpdateTankAnimation(tank)
		}
		return nil
	}

	if tank.State == types.TankStateStopped {
		tank.Direction = direction
		if uc.renderUseCases != nil {
			uc.renderUseCases.UpdateTankAnimation(tank)
		}
		return nil
	}

	directionCopy := direction
	tank.NextDirection = &directionCopy
	tank.State = types.TankStateBraking

	if uc.renderUseCases != nil {
		uc.renderUseCases.UpdateTankAnimation(tank)
	}
	return nil
}

func (uc *TankActionsUseCases) Move(tank *types.TankEntity) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}

	if tank.State == types.TankStateBraking {
		if tank.NextDirection == nil {
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
	if uc.soundUseCases != nil && !tank.IsEnemy() {
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
}

func (uc *TankActionsUseCases) SetMaxXPosition(tank *types.TankEntity) {
	mapSizePx := uc.mapUseCases.GetSizePx()
	maxX := float64(mapSizePx.Width - tank.Size.Width)
	tank.Position.X = maxX
}

func (uc *TankActionsUseCases) SetMinYPosition(tank *types.TankEntity) {
	tank.Position.Y = 0
}

func (uc *TankActionsUseCases) SetMaxYPosition(tank *types.TankEntity) {
	mapSizePx := uc.mapUseCases.GetSizePx()
	maxY := float64(mapSizePx.Height - tank.Size.Height)
	tank.Position.Y = maxY
}

func (uc *TankActionsUseCases) handleStopByCollision(tank *types.TankEntity) {
	tank.Position.X = uc.coordinateService.RoundToNearestMultipleOf4(
		tank.Position.X,
	)
	tank.Position.Y = uc.coordinateService.RoundToNearestMultipleOf4(
		tank.Position.Y,
	)
	tank.State = types.TankStateStopped
}
