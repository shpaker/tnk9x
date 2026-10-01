package bindings

import (
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/types"
)

// stickDeadZone — отклонение стика, ниже которого направления нет
const stickDeadZone = 0.5

// StickDirection — направление стика по доминирующей оси вне
// мёртвой зоны; x и y — отклонения -1..1, y растёт вниз
func StickDirection(x, y float64) (types.Direction, bool) {
	if math.Max(math.Abs(x), math.Abs(y)) < stickDeadZone {
		return types.DirectionUp, false
	}
	if math.Abs(x) >= math.Abs(y) {
		if x > 0 {
			return types.DirectionRight, true
		}
		return types.DirectionLeft, true
	}
	if y > 0 {
		return types.DirectionDown, true
	}
	return types.DirectionUp, true
}

// StandardGamepads дописывает в buffer подключённые геймпады
// со стандартной раскладкой по возрастанию ID — в порядке
// подключения
func StandardGamepads(buffer []ebiten.GamepadID) []ebiten.GamepadID {
	all := ebiten.AppendGamepadIDs(buffer)
	pads := all[:len(buffer)]
	for _, id := range all[len(buffer):] {
		if ebiten.IsStandardGamepadLayoutAvailable(id) {
			pads = append(pads, id)
		}
	}
	slices.Sort(pads[len(buffer):])
	return pads
}

// PlayerGamepad — геймпад игрока: первый подключённый у P1,
// второй у P2; buffer переиспользуется между кадрами
func PlayerGamepad(
	player types.PlayerTankNum,
	buffer []ebiten.GamepadID,
) (ebiten.GamepadID, []ebiten.GamepadID, bool) {
	buffer = StandardGamepads(buffer[:0])
	if int(player) < 0 || int(player) >= len(buffer) {
		return 0, buffer, false
	}
	return buffer[player], buffer, true
}
