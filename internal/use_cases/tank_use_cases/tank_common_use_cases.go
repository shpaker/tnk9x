package tank_use_cases

import (
	"errors"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// TankCommonUseCases фасад для работы с танками, объединяющий два компонента
type TankCommonUseCases struct {
	brakingService    interfaces.ITankBrakingService
	coordinateService interfaces.ICoordinateService
	bulletUseCases    interfaces.IBulletUseCases
	renderUseCases    interfaces.ITankRenderUseCases // Для управления анимацией танка
	tanksRepository   interfaces.ITanksRepository    // Репозиторий танков
}

// ============================================================================
// КОНСТРУКТОР
// ============================================================================

// NewTankCommonUseCases создает новый экземпляр TankCommonUseCases
func NewTankCommonUseCases(
	bulletUseCases interfaces.IBulletUseCases,
	brakingService interfaces.ITankBrakingService,
	coordinateService interfaces.ICoordinateService,
	renderUseCases interfaces.ITankRenderUseCases,
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

// --- Movement ---

// Update обновляет состояние танка (движение)
func (uc *TankCommonUseCases) Update(tank *types.TankEntity, dt float64) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}

	// Сначала синхронизируем анимацию с текущим состоянием (на случай, если состояние изменилось вне Update)
	if uc.renderUseCases != nil {
		uc.renderUseCases.SyncAnimationWithState(tank)
	}

	oldState := tank.State
	oldDirection := tank.Direction

	// Обрабатываем состояние Braking отдельно
	if tank.State == types.TankStateBraking {
		if uc.brakingService == nil {
			return errors.New("brakingService is not initialized")
		}
		err := uc.brakingService.HandleBrakingState(tank, dt)
		// Если изменилось направление (например, применился NextDirection), обновляем анимацию
		if oldDirection != tank.Direction && uc.renderUseCases != nil {
			uc.renderUseCases.UpdateTankAnimation(tank)
		}
		// Синхронизируем анимацию после обновления состояния (если оно изменилось)
		if oldState != tank.State && uc.renderUseCases != nil {
			uc.renderUseCases.SyncAnimationWithState(tank)
		}
		return err
	}

	// Обновляем позицию танка на основе скорости и направления
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

	// Дополнительная синхронизация после обновления позиции
	if uc.renderUseCases != nil {
		uc.renderUseCases.SyncAnimationWithState(tank)
	}

	return nil
}

// UpdateAllTanks обновляет все танки (игрок + враги) из репозитория
func (uc *TankCommonUseCases) UpdateAllTanks(dt float64) error {
	allTanks := uc.GetAllTanks()
	for _, tank := range allTanks {
		if tank != nil {
			if err := uc.Update(tank, dt); err != nil {
				// Продолжаем обновление остальных танков даже при ошибке
				_ = err
			}
		}
	}
	return nil
}

// GetAllTanks возвращает все танки (игрок + враги) из репозитория
func (uc *TankCommonUseCases) GetAllTanks() []*types.TankEntity {
	if uc.tanksRepository == nil {
		return []*types.TankEntity{}
	}
	return uc.tanksRepository.GetAllTanks()
}
