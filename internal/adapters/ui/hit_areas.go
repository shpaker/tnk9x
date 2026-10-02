package ui

import (
	"image"

	"github.com/shpaker/tnk9x/internal/types"
)

// HitAreas — прямоугольники пунктов последней отрисовки в логических
// координатах: рендер заполняет их в Draw, стейт проверяет попадание
// тапа или курсора в следующем Update
type HitAreas struct {
	rects []image.Rectangle
}

// Reset забывает пункты прошлой отрисовки
func (h *HitAreas) Reset() {
	h.rects = h.rects[:0]
}

// Add добавляет прямоугольник следующего пункта; индекс пункта —
// порядок добавления
func (h *HitAreas) Add(rect image.Rectangle) {
	h.rects = append(h.rects, rect)
}

// Hit — индекс пункта под точкой
func (h *HitAreas) Hit(position types.Position) (int, bool) {
	point := image.Pt(int(position.X), int(position.Y))
	for i, rect := range h.rects {
		if point.In(rect) {
			return i, true
		}
	}
	return 0, false
}

// RowRect — прямоугольник строки текста с верхним краем top: по
// высоте захватывает половину зазора сверху и снизу, чтобы между
// соседними строками не было мёртвых зон
func RowRect(left, right, top, rowHeight, gap float64) image.Rectangle {
	return image.Rect(
		int(left), int(top-gap/2),
		int(right), int(top+rowHeight+gap/2),
	)
}
