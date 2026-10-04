package types

import "github.com/shpaker/koleya"

// TankLattice — сетка, на узлах которой останавливаются танки: кратные
// 4px, поэтому танк легко входит в проходы шириной в себя
var TankLattice = koleya.Lattice{Step: 4}

// Dir переводит направление в направление koleya
func (d Direction) Dir() koleya.Dir {
	switch d {
	case DirectionDown:
		return koleya.Down
	case DirectionLeft:
		return koleya.Left
	case DirectionRight:
		return koleya.Right
	default:
		return koleya.Up
	}
}

// directionOf переводит прямое направление koleya в направление танка
func directionOf(d koleya.Dir) (Direction, bool) {
	switch d {
	case koleya.Up:
		return DirectionUp, true
	case koleya.Down:
		return DirectionDown, true
	case koleya.Left:
		return DirectionLeft, true
	case koleya.Right:
		return DirectionRight, true
	}
	return DirectionUp, false
}

// Drive — команда водителя ехать в направлении; едущий танк в другую
// сторону сначала докатывает до узла сетки, затем поворачивает
func (t *TankEntity) Drive(direction Direction) {
	t.drive = direction.Dir()
}

// Release — водитель отпустил: танк докатывает до узла сетки и встаёт
func (t *TankEntity) Release() {
	t.drive = koleya.None
}

// GetDrive возвращает текущую команду водителя; false — команды нет
func (t *TankEntity) GetDrive() (Direction, bool) {
	return directionOf(t.drive)
}

// Face поворачивает стоящий танк на месте; едущий не поворачивается
func (t *TankEntity) Face(direction Direction) bool {
	if t.motion.Moving() {
		return false
	}
	t.motion.Face(direction.Dir())
	t.Direction = direction
	return true
}

// Halt — танк упёрся: встаёт на месте без выравнивания по сетке,
// команда водителя сбрасывается
func (t *TankEntity) Halt() {
	t.motion.Halt()
	t.drive = koleya.None
}

// Move проводит такт движения по сетке и возвращает его события.
// Position и Direction остаются источником правды для остальной игры:
// коллизии правят Position, стоящий танк поворачивают через Direction
func (t *TankEntity) Move(
	profile koleya.Profile,
	surface koleya.Surface,
	dt float64,
) koleya.Events {
	if !t.motion.Moving() {
		t.motion.Face(t.Direction.Dir())
	}
	t.motion.Pos = koleya.Vec2{X: t.Position.X, Y: t.Position.Y}

	events := koleya.Step(
		&t.motion,
		koleya.Intent{Dir: t.drive},
		profile,
		surface,
		dt,
	)

	t.Position = Position{X: t.motion.Pos.X, Y: t.motion.Pos.Y}
	if direction, ok := directionOf(t.motion.Facing()); ok {
		t.Direction = direction
	}
	return events
}

// IsMoving — активный танк едет или докатывает до узла
func (t *TankEntity) IsMoving() bool {
	return t.State == TankStateActive && t.motion.Moving()
}

// IsDriving — активный танк едет по команде, а не докатывает
func (t *TankEntity) IsDriving() bool {
	return t.IsMoving() && !t.motion.Docking()
}

// IsDocking — танк докатывает до узла сетки
func (t *TankEntity) IsDocking() bool {
	return t.IsMoving() && t.motion.Docking()
}

// HasPendingTurn — танк докатывает, чтобы затем повернуть
func (t *TankEntity) HasPendingTurn() bool {
	return t.motion.Pending() != koleya.None
}
