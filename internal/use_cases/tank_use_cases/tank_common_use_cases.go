package tank_use_cases

import (
	"errors"

	"github.com/shpaker/koleya"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

var _ interfaces.ITankCommonUseCases = (*TankCommonUseCases)(nil)

type TankCommonUseCases struct {
	renderUseCases  interfaces.IRenderUseCases
	tanksRepository interfaces.ITanksRepository
	specsUseCases   interfaces.ISpecsUseCases
	mapUseCases     interfaces.IMapUseCases
	stageSession    *session_entities.StageSessionEntity
}

func NewTankCommonUseCases(
	renderUseCases interfaces.IRenderUseCases,
	tanksRepository interfaces.ITanksRepository,
	specsUseCases interfaces.ISpecsUseCases,
	mapUseCases interfaces.IMapUseCases,
	stageSession *session_entities.StageSessionEntity,
) *TankCommonUseCases {
	return &TankCommonUseCases{
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
			tank.Halt()
		}
		return nil
	}

	uc.renderUseCases.SyncTankAnimationWithState(tank)

	tank.PrevPosition = tank.Position
	oldDirection := tank.Direction

	tank.Move(uc.motionProfile(tank), uc.surface(tank), dt)

	if oldDirection != tank.Direction {
		uc.renderUseCases.UpdateTankAnimation(tank)
	}
	uc.renderUseCases.SyncTankAnimationWithState(tank)

	return nil
}

// defaultTankSpeed — скорость танка без спецификаций, px/s
const defaultTankSpeed = 32.0

// iceGrip — сцепление льда, px/s²: базовый танк проскальзывает
// по инерции примерно на узел сетки дальше, быстрый — ещё дальше
const iceGrip = 128.0

// motionProfile — классическое движение танчиков: полная скорость
// сразу и докатывание до узла сетки на полной скорости
func (uc *TankCommonUseCases) motionProfile(
	tank *types.TankEntity,
) koleya.Profile {
	speed := defaultTankSpeed
	if tank.GetSpecs() != nil {
		speed = tank.GetSpecs().GetSpeed()
	}
	return koleya.Classic(speed)
}

// surface — грунт под танком: лёд ограничивает разгон и торможение
func (uc *TankCommonUseCases) surface(tank *types.TankEntity) koleya.Surface {
	if uc.isOnIce(tank) {
		return koleya.Surface{Grip: iceGrip}
	}
	return koleya.Ground
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
		if tank != nil && tank.IsDriving() {
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
// всех врагов уровня (заморозка общая, в сессии), таймер врага —
// конкретные танки игроков (заморозка уходит вместе с танком)
func (uc *TankCommonUseCases) IsFrozen(tank *types.TankEntity) bool {
	if tank == nil {
		return false
	}
	if !tank.IsEnemy() {
		return tank.IsFrozen()
	}
	return uc.stageSession != nil && uc.stageSession.AreEnemiesFrozen()
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
