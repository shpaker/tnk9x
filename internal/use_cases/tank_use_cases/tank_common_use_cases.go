package tank_use_cases

import (
	"errors"
	"fmt"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ITankCommonUseCases = (*TankCommonUseCases)(nil)

type TankCommonUseCases struct {
	brakingService  interfaces.ITankBrakingService
	renderUseCases  interfaces.IRenderUseCases
	tanksRepository interfaces.ITanksRepository
	specsUseCases   interfaces.ISpecsUseCases
}

func NewTankCommonUseCases(
	brakingService interfaces.ITankBrakingService,
	renderUseCases interfaces.IRenderUseCases,
	tanksRepository interfaces.ITanksRepository,
	specsUseCases interfaces.ISpecsUseCases,
) *TankCommonUseCases {
	return &TankCommonUseCases{
		brakingService:  brakingService,
		renderUseCases:  renderUseCases,
		tanksRepository: tanksRepository,
		specsUseCases:   specsUseCases,
	}
}

func (uc *TankCommonUseCases) Update(tank *types.TankEntity, dt float64) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}

	if uc.renderUseCases != nil {
		uc.renderUseCases.SyncTankAnimationWithState(tank)
	}

	tank.PrevPosition = tank.Position

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
		// Получаем скорость танка из спецификаций
		speed := float64(32.0) // Значение по умолчанию
		if tank.GetSpecs() != nil {
			speed = tank.GetSpecs().GetSpeed()
		}
		delta := speed * dt

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

// LevelUp повышает уровень танка на единицу (максимум 3)
func (uc *TankCommonUseCases) LevelUp(tank *types.TankEntity) {
	if tank == nil || tank.GetSpecs() == nil {
		return
	}
	currentLevel := tank.GetSpecs().GetLevel()
	if currentLevel < 3 {
		// Получаем новые спецификации для следующего уровня
		newSpecs := uc.specsUseCases.GetTankSpecs(
			tank.IsEnemy(),
			currentLevel+1,
		)
		if newSpecs != nil {
			tank.SetSpecs(newSpecs)
			// Обновляем анимацию танка для отображения нового уровня
			if uc.renderUseCases != nil {
				uc.renderUseCases.UpdateTankAnimation(tank)
			}
		}
	}
}

// LevelDown понижает уровень танка на единицу (минимум 0)
func (uc *TankCommonUseCases) LevelDown(tank *types.TankEntity) {
	if tank == nil || tank.GetSpecs() == nil {
		return
	}
	currentLevel := tank.GetSpecs().GetLevel()
	if currentLevel > 0 {
		// Получаем новые спецификации для предыдущего уровня
		newSpecs := uc.specsUseCases.GetTankSpecs(
			tank.IsEnemy(),
			currentLevel-1,
		)
		if newSpecs != nil {
			tank.SetSpecs(newSpecs)
			// Обновляем анимацию танка для отображения нового уровня
			if uc.renderUseCases != nil {
				uc.renderUseCases.UpdateTankAnimation(tank)
			}
		}
	}
}

// GetTankAnimationName возвращает имя анимации танка на основе его спецификаций
func (uc *TankCommonUseCases) GetTankAnimationName(
	tank *types.TankEntity,
) string {
	if tank == nil {
		return "player1_level1_tank_up"
	}

	// Получаем уровень танка из спецификаций
	tankLevel := uint(0)
	if tank.GetSpecs() != nil {
		tankLevel = tank.GetSpecs().GetLevel()
	}
	if tankLevel > 3 {
		tankLevel = 3
	}
	modelName := fmt.Sprintf("level%d", tankLevel+1)

	roleStr := string(tank.GetRole())
	if roleStr == "" {
		roleStr = "player1"
	}

	prefix := roleStr + "_" + modelName

	var direction string
	switch tank.Direction {
	case types.DirectionUp:
		direction = "up"
	case types.DirectionDown:
		direction = "down"
	case types.DirectionLeft:
		direction = "left"
	case types.DirectionRight:
		direction = "right"
	default:
		direction = "up"
	}

	return fmt.Sprintf("%s_tank_%s", prefix, direction)
}

// SetRenderUseCases устанавливает renderUseCases
func (uc *TankCommonUseCases) SetRenderUseCases(
	renderUseCases interfaces.IRenderUseCases,
) {
	uc.renderUseCases = renderUseCases
}
