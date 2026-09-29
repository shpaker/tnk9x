package tank_use_cases

import (
	"errors"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

var _ interfaces.ITankCommonUseCases = (*TankCommonUseCases)(nil)

// Разгон: танк трогается с доли startSpeedFraction полной скорости
// и набирает её за accelerationTime секунд; на льду гусеницы
// буксуют — разгон медленнее в iceAccelerationFactor раз
const (
	startSpeedFraction    = 0.05
	accelerationTime      = 1.0
	iceAccelerationFactor = 0.5
)

type TankCommonUseCases struct {
	brakingService  interfaces.ITankBrakingService
	renderUseCases  interfaces.IRenderUseCases
	tanksRepository interfaces.ITanksRepository
	specsUseCases   interfaces.ISpecsUseCases
	mapUseCases     interfaces.IMapUseCases
	stageSession    *session_entities.StageSessionEntity
	// accelerationEnabled — танки разгоняются с места, а не сразу
	// едут с полной скоростью (game.tank_acceleration)
	accelerationEnabled bool
}

func NewTankCommonUseCases(
	brakingService interfaces.ITankBrakingService,
	renderUseCases interfaces.IRenderUseCases,
	tanksRepository interfaces.ITanksRepository,
	specsUseCases interfaces.ISpecsUseCases,
	mapUseCases interfaces.IMapUseCases,
	stageSession *session_entities.StageSessionEntity,
	accelerationEnabled bool,
) *TankCommonUseCases {
	return &TankCommonUseCases{
		brakingService:      brakingService,
		renderUseCases:      renderUseCases,
		tanksRepository:     tanksRepository,
		specsUseCases:       specsUseCases,
		mapUseCases:         mapUseCases,
		stageSession:        stageSession,
		accelerationEnabled: accelerationEnabled,
	}
}

func (uc *TankCommonUseCases) Update(tank *types.TankEntity, dt float64) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}

	tank.UpdateReload(dt)

	// Замороженный бонусом-таймером враг замирает на месте
	if tank.IsEnemy() && uc.stageSession != nil &&
		uc.stageSession.AreEnemiesFrozen() {
		tank.PrevPosition = tank.Position
		tank.SetSpeed(0)
		return nil
	}

	// Остановившийся танк теряет набранную скорость
	if tank.State == types.TankStateStopped {
		tank.SetSpeed(0)
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
		delta := uc.accelerate(tank, dt) * dt

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

// accelerate возвращает скорость танка на этот тик: с разгоном
// она растёт от доли полной до полной, без разгона сразу полная
func (uc *TankCommonUseCases) accelerate(
	tank *types.TankEntity,
	dt float64,
) float64 {
	maxSpeed := float64(32.0) // Значение по умолчанию
	if tank.GetSpecs() != nil {
		maxSpeed = tank.GetSpecs().GetSpeed()
	}
	if !uc.accelerationEnabled {
		tank.SetSpeed(maxSpeed)
		return maxSpeed
	}

	acceleration := maxSpeed * (1 - startSpeedFraction) / accelerationTime
	if uc.isOnIce(tank) {
		acceleration *= iceAccelerationFactor
	}
	speed := max(tank.GetSpeed(), maxSpeed*startSpeedFraction)
	speed = min(maxSpeed, speed+acceleration*dt)
	tank.SetSpeed(speed)
	return speed
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
