package use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

func TestHUDUseCases_EnemyIconOffsets(t *testing.T) {
	uc := use_cases.NewHUDUseCases()

	tests := []struct {
		name     string
		count    uint
		columns  int
		rows     int
		iconSize int
		want     []types.Position
	}{
		{
			name:    "ноль врагов — пустая сетка",
			count:   0,
			columns: 2, rows: 10, iconSize: 8,
			want: []types.Position{},
		},
		{
			name:    "одна иконка в левом верхнем углу",
			count:   1,
			columns: 2, rows: 10, iconSize: 8,
			want: []types.Position{{X: 0, Y: 0}},
		},
		{
			name:    "заполнение построчно",
			count:   3,
			columns: 2, rows: 10, iconSize: 8,
			want: []types.Position{
				{X: 0, Y: 0},
				{X: 8, Y: 0},
				{X: 0, Y: 8},
			},
		},
		{
			name:    "количество сверх ёмкости ограничивается сеткой",
			count:   21,
			columns: 2, rows: 10, iconSize: 8,
			want: fullGridOffsets(),
		},
		{
			name:    "переполнение uint ограничивается сеткой",
			count:   ^uint(0),
			columns: 2, rows: 10, iconSize: 8,
			want: fullGridOffsets(),
		},
		{
			name:    "нулевые колонки — nil",
			count:   5,
			columns: 0, rows: 10, iconSize: 8,
			want: nil,
		},
		{
			name:    "нулевые ряды — nil",
			count:   5,
			columns: 2, rows: 0, iconSize: 8,
			want: nil,
		},
		{
			name:    "нулевой размер иконки — nil",
			count:   5,
			columns: 2, rows: 10, iconSize: 0,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uc.EnemyIconOffsets(
				tt.count,
				tt.columns,
				tt.rows,
				tt.iconSize,
			)
			if len(got) != len(tt.want) {
				t.Fatalf(
					"количество смещений %d, ожидалось %d",
					len(got),
					len(tt.want),
				)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf(
						"смещение %d: %v, ожидалось %v",
						i,
						got[i],
						tt.want[i],
					)
				}
			}
		})
	}
}

// fullGridOffsets — все 20 смещений сетки 2x10 с шагом 8, построчно
func fullGridOffsets() []types.Position {
	offsets := make([]types.Position, 0, 20)
	for i := 0; i < 20; i++ {
		offsets = append(offsets, types.Position{
			X: float64(i % 2 * 8),
			Y: float64(i / 2 * 8),
		})
	}
	return offsets
}
