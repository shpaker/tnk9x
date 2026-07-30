package game_over

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var gameOverTextColor = color.NRGBA{R: 228, G: 68, B: 52, A: 255}

// GameOverRendererAdapter рисует экран «GAME OVER»
type GameOverRendererAdapter struct {
	fontFace      text.Face
	titleFontSize int
}

func NewGameOverRendererAdapter(
	fontFace text.Face,
	titleFontSize int,
) *GameOverRendererAdapter {
	return &GameOverRendererAdapter{
		fontFace:      fontFace,
		titleFontSize: titleFontSize,
	}
}

func (r *GameOverRendererAdapter) DrawAll(screen *ebiten.Image) {
	screenBounds := screen.Bounds()
	actualWidth := float64(screenBounds.Dx())
	actualHeight := float64(screenBounds.Dy())

	screen.Fill(color.Black)

	label := "GAME OVER"
	textWidth, textHeight := text.Measure(label, r.fontFace, 0)
	op := &text.DrawOptions{}
	op.GeoM.Translate(
		(actualWidth-textWidth)/2,
		(actualHeight-textHeight)/2,
	)
	op.ColorScale.ScaleWithColor(gameOverTextColor)
	text.Draw(screen, label, r.fontFace, op)
}
