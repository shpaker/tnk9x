package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawFade затемняет экран: fade 0 — без изменений, 1 — полностью
// чёрный; переходы между экранами идут через чёрный
func DrawFade(screen *ebiten.Image, fade float64) {
	if fade <= 0 {
		return
	}
	bounds := screen.Bounds()
	vector.FillRect(
		screen,
		0, 0,
		float32(bounds.Dx()), float32(bounds.Dy()),
		color.NRGBA{A: uint8(min(fade, 1) * 255)},
		false,
	)
}
