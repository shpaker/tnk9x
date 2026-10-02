package states

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// stepIndex сдвигает курсор меню на шаг вверх или вниз
// в пределах count пунктов
func stepIndex(index, count int, moveUp, moveDown bool) int {
	if moveUp && index > 0 {
		index--
	}
	if moveDown && index < count-1 {
		index++
	}
	return index
}

// rowHitTest — пункт меню последней отрисовки под точкой
type rowHitTest func(position types.Position) (int, bool)

// pointerIndex — пункт под мышью или тапом: наведение переносит
// курсор меню, тап или клик переносит и выбирает пункт
func pointerIndex(
	menuInput interfaces.IMenuInputAdapter,
	hit rowHitTest,
	index int,
) (int, bool) {
	if position, pointed := menuInput.Pointed(); pointed {
		if row, ok := hit(position); ok {
			index = row
		}
	}
	if position, tapped := menuInput.Tapped(); tapped {
		if row, ok := hit(position); ok {
			return row, true
		}
	}
	return index, false
}

// tappedAnywhere — тап или клик в любом месте игрового экрана
func tappedAnywhere(menuInput interfaces.IMenuInputAdapter) bool {
	_, tapped := menuInput.Tapped()
	return tapped
}
