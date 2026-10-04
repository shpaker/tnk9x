package testutil

import (
	"github.com/shpaker/koleya"

	"github.com/shpaker/tnk9x/internal/types"
)

// MovingTank делает танк активным и трогает его туда, куда он смотрит,
// со скоростью из спецификаций, не сдвигая с места: такт нулевой длины
func MovingTank(tank *types.TankEntity) *types.TankEntity {
	tank.State = types.TankStateActive
	tank.Drive(tank.Direction)
	tank.Move(classic(tank), koleya.Ground, 0)
	return tank
}

// DockingTank — едущий танк, которого отпустили: он докатывает до узла
// сетки. Танк должен стоять вне узла, иначе он сразу остановится
func DockingTank(tank *types.TankEntity) *types.TankEntity {
	MovingTank(tank)
	tank.Release()
	tank.Move(classic(tank), koleya.Ground, 0)
	return tank
}

// classic — профиль движения танчиков со скоростью танка
func classic(tank *types.TankEntity) koleya.Profile {
	if tank.GetSpecs() == nil {
		return koleya.Classic(32)
	}
	return koleya.Classic(tank.GetSpecs().GetSpeed())
}
