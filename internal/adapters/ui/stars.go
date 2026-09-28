// Package ui содержит общие примитивы отрисовки меню.
package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Цвета звёзд: набранная и пустая
var (
	StarFilledColor = color.NRGBA{R: 252, G: 196, B: 36, A: 255}
	StarEmptyColor  = color.NRGBA{R: 70, G: 70, B: 70, A: 255}
)

// DrawStar рисует пятиконечную звезду с центром (cx, cy)
// и внешним радиусом radius
func DrawStar(
	screen *ebiten.Image,
	cx, cy, radius float32,
	starColor color.Color,
) {
	const points = 5
	inner := radius * 0.45

	var path vector.Path
	for i := 0; i < points*2; i++ {
		r := radius
		if i%2 == 1 {
			r = inner
		}
		angle := -math.Pi/2 + float64(i)*math.Pi/points
		x := cx + r*float32(math.Cos(angle))
		y := cy + r*float32(math.Sin(angle))
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	path.Close()

	drawOptions := &vector.DrawPathOptions{AntiAlias: true}
	drawOptions.ColorScale.ScaleWithColor(starColor)
	vector.FillPath(screen, &path, &vector.FillOptions{}, drawOptions)
}

// DrawStarsRow рисует ряд из total звёзд, filled из них набраны;
// cx — центр ряда, step — расстояние между центрами
func DrawStarsRow(
	screen *ebiten.Image,
	cx, cy, radius, step float32,
	filled, total uint,
) {
	left := cx - step*float32(total-1)/2
	for i := uint(0); i < total; i++ {
		starColor := StarEmptyColor
		if i < filled {
			starColor = StarFilledColor
		}
		DrawStar(screen, left+step*float32(i), cy, radius, starColor)
	}
}
