package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IHUDUseCases = (*HUDUseCases)(nil)

// HUDUseCases — расчёт раскладки элементов боковой панели HUD
type HUDUseCases struct{}

func NewHUDUseCases() *HUDUseCases {
	return &HUDUseCases{}
}

// EnemyIconOffsets возвращает смещения иконок оставшихся в резерве врагов
// относительно левого верхнего угла сетки счётчика. Сетка заполняется
// построчно (слева направо, сверху вниз); count ограничивается ёмкостью
// сетки columns*rows
func (uc *HUDUseCases) EnemyIconOffsets(
	count uint,
	columns int,
	rows int,
	iconSize int,
) []types.Position {
	if columns <= 0 || rows <= 0 || iconSize <= 0 {
		return nil
	}

	capacity := uint(columns) * uint(rows)
	if count > capacity {
		count = capacity
	}

	offsets := make([]types.Position, 0, count)
	for i := 0; i < int(count); i++ {
		offsets = append(offsets, types.Position{
			X: float64(i % columns * iconSize),
			Y: float64(i / columns * iconSize),
		})
	}

	return offsets
}
