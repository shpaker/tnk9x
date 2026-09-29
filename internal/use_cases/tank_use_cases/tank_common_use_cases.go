package tank_use_cases

import (
	"errors"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

var _ interfaces.ITankCommonUseCases = (*TankCommonUseCases)(nil)

type TankCommonUseCases struct {
	brakingService  interfaces.ITankBrakingService
	renderUseCases  interfaces.IRenderUseCases
	tanksRepository interfaces.ITanksRepository
	specsUseCases   interfaces.ISpecsUseCases
	mapUseCases     interfaces.IMapUseCases
	stageSession    *session_entities.StageSessionEntity
}

func NewTankCommonUseCases(
	brakingService interfaces.ITankBrakingService,
	renderUseCases interfaces.IRenderUseCases,
	tanksRepository interfaces.ITanksRepository,
	specsUseCases interfaces.ISpecsUseCases,
	mapUseCases interfaces.IMapUseCases,
	stageSession *session_entities.StageSessionEntity,
) *TankCommonUseCases {
	return &TankCommonUseCases{
		brakingService:  brakingService,
		renderUseCases:  renderUseCases,
		tanksRepository: tanksRepository,
		specsUseCases:   specsUseCases,
		mapUseCases:     mapUseCases,
		stageSession:    stageSession,
	}
}

func (uc *TankCommonUseCases) Update(tank *types.TankEntity, dt float64) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}

	tank.UpdateReload(dt)

	// Замороженный бонусом-таймером танк замирает на месте; игрок
	// останавливается совсем — отпущенная за заморозку клавиша
	// не должна оставить его едущим
	if uc.IsFrozen(tank) {
		tank.PrevPosition = tank.Position
		if !tank.IsEnemy() {
			tank.State = types.TankStateStopped
		}
		return nil
	}

	uc.renderUseCases.SyncTankAnimationWithState(tank)

	tank.PrevPosition = tank.Position

	oldState := tank.State
	oldDirection := tank.Direction

	if tank.State == types.TankStateBraking {
		err := uc.brakingService.HandleBrakingState(tank, dt, uc.isOnIce(tank))

		if oldDirection != tank.Direction {
			uc.renderUseCases.UpdateTankAnimation(tank)
		}

		if oldState != tank.State {
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

	uc.renderUseCases.SyncTankAnimationWithState(tank)

	return nil
}

// isOnIce — центр танка находится на блоке льда
func (uc *TankCommonUseCases) isOnIce(tank *types.TankEntity) bool {
	if uc.mapUseCases == nil {
		return false
	}
	center := types.Position{
		X: tank.Position.X + float64(tank.Size.Width)/2,
		Y: tank.Position.Y + float64(tank.Size.Height)/2,
	}
	return uc.mapUseCases.IsIceAt(center)
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
	return uc.tanksRepository.GetAllTanks()
}

// GetAllPlayerTanks возвращает все танки игроков (не врагов)
func (uc *TankCommonUseCases) GetAllPlayerTanks() []*types.TankEntity {
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
			uc.renderUseCases.UpdateTankAnimation(tank)
		}
	}
}

// IsFrozen реализует ITankCommonUseCases: таймер игрока морозит
// врагов, таймер врага — игроков
func (uc *TankCommonUseCases) IsFrozen(tank *types.TankEntity) bool {
	if tank == nil || uc.stageSession == nil {
		return false
	}
	if tank.IsEnemy() {
		return uc.stageSession.AreEnemiesFrozen()
	}
	return uc.stageSession.ArePlayersFrozen()
}

// SetMaxLevel реализует ITankCommonUseCases
func (uc *TankCommonUseCases) SetMaxLevel(tank *types.TankEntity) {
	if tank == nil || tank.GetSpecs() == nil {
		return
	}
	newSpecs := uc.specsUseCases.GetTankSpecs(tank.IsEnemy(), 3)
	if newSpecs == nil {
		return
	}
	tank.SetSpecs(newSpecs)
	uc.renderUseCases.UpdateTankAnimation(tank)
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
			uc.renderUseCases.UpdateTankAnimation(tank)
		}
	}
}
