package use_cases

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
	tilesUseCases     *TilesUseCases
}

// ============================================================================
// КОНСТРУКТОР
// ============================================================================

// NewTankCommonUseCases создает новый экземпляр TankCommonUseCases
func NewTankCommonUseCases(
	bulletUseCases interfaces.IBulletUseCases,
	tilesUseCases *TilesUseCases,
	brakingService interfaces.ITankBrakingService,
	coordinateService interfaces.ICoordinateService,
) *TankCommonUseCases {
	uc := &TankCommonUseCases{
		brakingService:    brakingService,
		coordinateService: coordinateService,
		bulletUseCases:    bulletUseCases,
		tilesUseCases:     tilesUseCases,
	}

	return uc
}

// --- Movement ---

// Update обновляет состояние танка (движение)
func (uc *TankCommonUseCases) Update(tank *types.TankEntity, dt float64) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}

	// Обрабатываем состояние Braking отдельно
	if tank.State == types.TankStateBraking {
		if uc.brakingService == nil {
			return errors.New("brakingService is not initialized")
		}
		return uc.brakingService.HandleBrakingState(tank, dt)
	}

	// Обновляем позицию танка на основе скорости и направления
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

	return nil
}
