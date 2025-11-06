package use_cases

import (
	"errors"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

// TankCommonUseCases фасад для работы с танками, объединяющий два компонента
type TankCommonUseCases struct {
	brakingService    interfaces.ITankBrakingService
	coordinateService interfaces.ICoordinateService
	bulletUseCases    interfaces.IBulletUseCases
	tilesUseCases     *TilesUseCases
	renderUseCases    interfaces.ITankRenderUseCases // Для управления анимацией танка
	tanksRepository   interfaces.ITanksRepository    // Репозиторий танков
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
	renderUseCases interfaces.ITankRenderUseCases,
	tanksRepository interfaces.ITanksRepository,
) *TankCommonUseCases {
	uc := &TankCommonUseCases{
		brakingService:    brakingService,
		coordinateService: coordinateService,
		bulletUseCases:    bulletUseCases,
		tilesUseCases:     tilesUseCases,
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
	uc.syncAnimationWithState(tank, tank.State)

	oldState := tank.State

	// Обрабатываем состояние Braking отдельно
	if tank.State == types.TankStateBraking {
		if uc.brakingService == nil {
			return errors.New("brakingService is not initialized")
		}
		err := uc.brakingService.HandleBrakingState(tank, dt)
		// Синхронизируем анимацию после обновления состояния (если оно изменилось)
		if oldState != tank.State {
			uc.syncAnimationWithState(tank, tank.State)
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
	uc.syncAnimationWithState(tank, tank.State)

	return nil
}

// syncAnimationWithState синхронизирует анимацию гусениц с состоянием танка
func (uc *TankCommonUseCases) syncAnimationWithState(
	tank *types.TankEntity,
	tankState types.TankState,
) {
	if uc.tilesUseCases == nil {
		return
	}

	if tank.Image == nil {
		return
	}
	anim, ok := tank.Image.(*image_providers.AnimationProvider)
	if !ok {
		return
	}

	// Если танк стоит - анимация должна быть остановлена
	if tankState == types.TankStateStopped {
		if anim.IsAnimating {
			uc.tilesUseCases.StopAnimation(anim)
		}
		return
	}

	// Определяем, должна ли анимация быть запущена
	shouldAnimate := tankState == types.TankStateMoving ||
		tankState == types.TankStateBraking

	// Синхронизируем: если состояние требует анимации, но она остановлена - запускаем
	// Если состояние не требует анимации, но она запущена - останавливаем
	if shouldAnimate && !anim.IsAnimating {
		uc.tilesUseCases.StartAnimation(anim)
	} else if !shouldAnimate && anim.IsAnimating {
		uc.tilesUseCases.StopAnimation(anim)
	}
}

// GetAllTanks возвращает все танки (игрок + враги) из репозитория
func (uc *TankCommonUseCases) GetAllTanks() []*types.TankEntity {
	if uc.tanksRepository == nil {
		return []*types.TankEntity{}
	}
	return uc.tanksRepository.GetAllTanks()
}
