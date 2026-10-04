package tank_use_cases

import (
	"errors"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ITankActionsUseCases = (*TankActionsUseCases)(nil)

type TankActionsUseCases struct {
	bulletUseCases        interfaces.IBulletUseCases
	commonUseCases        interfaces.ITankCommonUseCases
	renderUseCases        interfaces.IRenderUseCases
	mapUseCases           interfaces.IMapUseCases
	soundUseCases         interfaces.ISoundUseCases
	visualEffectsUseCases interfaces.IVisualEffectsUseCases
}

func NewTankActionsUseCases(
	bulletUseCases interfaces.IBulletUseCases,
	commonUseCases interfaces.ITankCommonUseCases,
	renderUseCases interfaces.IRenderUseCases,
	mapUseCases interfaces.IMapUseCases,
	soundUseCases interfaces.ISoundUseCases,
	visualEffectsUseCases interfaces.IVisualEffectsUseCases,
) *TankActionsUseCases {
	return &TankActionsUseCases{
		bulletUseCases:        bulletUseCases,
		commonUseCases:        commonUseCases,
		renderUseCases:        renderUseCases,
		mapUseCases:           mapUseCases,
		soundUseCases:         soundUseCases,
		visualEffectsUseCases: visualEffectsUseCases,
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

// Rotate поворачивает стоящий танк на месте; едущий сначала докатывает
// до узла сетки, затем едет в новом направлении
func (uc *TankActionsUseCases) Rotate(
	tank *types.TankEntity,
	direction types.Direction,
) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}
	if direction == tank.Direction || uc.commonUseCases.IsFrozen(tank) {
		return nil
	}

	if tank.IsMoving() {
		tank.Drive(direction)
	} else {
		tank.Face(direction)
	}
	uc.renderUseCases.UpdateTankAnimation(tank)
	return nil
}

// Move трогает танк в направлении, куда он смотрит; докатывающий танк
// едет дальше, если не ждёт поворота
func (uc *TankActionsUseCases) Move(tank *types.TankEntity) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}
	if uc.commonUseCases.IsFrozen(tank) {
		return nil
	}
	if _, ok := tank.GetDrive(); !ok {
		tank.Drive(tank.Direction)
	}
	return nil
}

// Stop: водитель отпустил — танк докатывает до узла сетки; упёрся
// (byCollision) — встаёт на месте: позиция уже разрешена коллизией,
// округление сдвигало бы танк с места контакта и порождало дрожание
func (uc *TankActionsUseCases) Stop(tank *types.TankEntity, byCollision bool) {
	if !tank.IsActive() {
		return
	}
	if byCollision {
		tank.Halt()
		return
	}
	tank.Release()
}

func (uc *TankActionsUseCases) Shoot(tank *types.TankEntity) error {
	if !tank.IsActive() {
		return errors.New("tank is not active")
	}
	if uc.commonUseCases.IsFrozen(tank) {
		return nil
	}
	fired, err := uc.bulletUseCases.ShootBullet(tank)
	if err != nil || !fired {
		return err
	}
	// Звук, вспышка у ствола и отдача — только у вылетевшей пули:
	// пока пуля в полёте, повторные нажатия танк не дёргают
	if !tank.IsEnemy() {
		uc.soundUseCases.RequestSound(types.SoundIDFire, false)
	}
	uc.visualEffectsUseCases.RequestEffect(types.VisualEventEntity{
		Kind:      types.VisualEventShot,
		Direction: tank.Direction,
		Tank:      tank,
	})
	return nil
}

// ApplyDecision поворачивает остановленный танк по решению AI
// и трогает его с места, если решение требует движения
func (uc *TankActionsUseCases) ApplyDecision(
	tank *types.TankEntity,
	decision types.EnemyAIDecision,
) {
	if !tank.IsStopped() {
		return
	}
	_ = uc.Rotate(tank, decision.Direction)
	if decision.Move {
		_ = uc.Move(tank)
	}
}

func (uc *TankActionsUseCases) SetMinXPosition(tank *types.TankEntity) {
	tank.Position.X = 0
	uc.haltAtBoundary(tank)
}

func (uc *TankActionsUseCases) SetMaxXPosition(tank *types.TankEntity) {
	mapSizePx := uc.mapUseCases.GetSizePx()
	maxX := float64(mapSizePx.Width - tank.Size.Width)
	tank.Position.X = maxX
	uc.haltAtBoundary(tank)
}

func (uc *TankActionsUseCases) SetMinYPosition(tank *types.TankEntity) {
	tank.Position.Y = 0
	uc.haltAtBoundary(tank)
}

func (uc *TankActionsUseCases) SetMaxYPosition(tank *types.TankEntity) {
	mapSizePx := uc.mapUseCases.GetSizePx()
	maxY := float64(mapSizePx.Height - tank.Size.Height)
	tank.Position.Y = maxY
	uc.haltAtBoundary(tank)
}

// haltAtBoundary прерывает докатывание у края карты: танк клампится
// каждый тик, узел за краем недостижим — без остановки он навсегда
// останется докатывающим
func (uc *TankActionsUseCases) haltAtBoundary(tank *types.TankEntity) {
	if tank.IsDocking() {
		tank.Halt()
	}
}
