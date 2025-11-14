package tank_use_cases

import (
	"errors"

	"github.com/shpaker/tnk25/internal/interfaces"
	"github.com/shpaker/tnk25/internal/types"
)

type TankCommonUseCases struct {
	brakingService    interfaces.ITankBrakingService
	coordinateService interfaces.ICoordinateService
	bulletUseCases    interfaces.IBulletUseCases
	renderUseCases    interfaces.IRenderUseCases
	tanksRepository   interfaces.ITanksRepository
}

func NewTankCommonUseCases(
	bulletUseCases interfaces.IBulletUseCases,
	brakingService interfaces.ITankBrakingService,
	coordinateService interfaces.ICoordinateService,
	renderUseCases interfaces.IRenderUseCases,
	tanksRepository interfaces.ITanksRepository,
) *TankCommonUseCases {
	uc := &TankCommonUseCases{
		brakingService:    brakingService,
		coordinateService: coordinateService,
		bulletUseCases:    bulletUseCases,
		renderUseCases:    renderUseCases,
		tanksRepository:   tanksRepository,
	}

	return uc
}

func (uc *TankCommonUseCases) Update(tank *types.TankEntity, dt float64) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}

	if uc.renderUseCases != nil {
		uc.renderUseCases.SyncTankAnimationWithState(tank)
	}

	oldState := tank.State
	oldDirection := tank.Direction

	if tank.State == types.TankStateBraking {
		if uc.brakingService == nil {
			return errors.New("brakingService is not initialized")
		}
		err := uc.brakingService.HandleBrakingState(tank, dt)

		if oldDirection != tank.Direction && uc.renderUseCases != nil {
			uc.renderUseCases.UpdateTankAnimation(tank)
		}

		if oldState != tank.State && uc.renderUseCases != nil {
			uc.renderUseCases.SyncTankAnimationWithState(tank)
		}
		return err
	}

	if tank.State == types.TankStateMoving {
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
	}

	if uc.renderUseCases != nil {
		uc.renderUseCases.SyncTankAnimationWithState(tank)
	}

	return nil
}

func (uc *TankCommonUseCases) UpdateAllTanks(dt float64) error {
	allTanks := uc.GetAllTanks()
	for _, tank := range allTanks {
		if tank != nil {
			if err := uc.Update(tank, dt); err != nil {
				_ = err
			}
		}
	}
	return nil
}

func (uc *TankCommonUseCases) GetAllTanks() []*types.TankEntity {
	if uc.tanksRepository == nil {
		return []*types.TankEntity{}
	}
	return uc.tanksRepository.GetAllTanks()
}

// GetAllPlayerTanks возвращает все танки игроков (не врагов)
func (uc *TankCommonUseCases) GetAllPlayerTanks() []*types.TankEntity {
	if uc.tanksRepository == nil {
		return []*types.TankEntity{}
	}
	return uc.tanksRepository.GetActivePlayerTanks()
}

// IsAnyPlayerTankMoving проверяет, двигается ли хотя бы один танк игрока
func (uc *TankCommonUseCases) IsAnyPlayerTankMoving() bool {
	playerTanks := uc.GetAllPlayerTanks()
	for _, tank := range playerTanks {
		if tank != nil && tank.State == types.TankStateMoving {
			return true
		}
	}
	return false
}
