package curtain

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// curtainBackdropColor — серый фон шторки, как между этапами в NES
var curtainBackdropColor = color.NRGBA{R: 109, G: 109, B: 109, A: 255}

// CurtainRendererAdapter рисует шторку «STAGE N» перед этапом
type CurtainRendererAdapter struct {
	hudFontFace text.Face
}

func NewCurtainRendererAdapter(
	hudFontFace text.Face,
) *CurtainRendererAdapter {
	return &CurtainRendererAdapter{hudFontFace: hudFontFace}
}

func (r *CurtainRendererAdapter) DrawAll(
	screen *ebiten.Image,
	stageLabel string,
) {
	screenBounds := screen.Bounds()
	actualWidth := float64(screenBounds.Dx())
	actualHeight := float64(screenBounds.Dy())

	screen.Fill(curtainBackdropColor)

	textWidth, textHeight := text.Measure(stageLabel, r.hudFontFace, 0)
	op := &text.DrawOptions{}
	op.GeoM.Translate(
		(actualWidth-textWidth)/2,
		(actualHeight-textHeight)/2,
	)
	op.ColorScale.ScaleWithColor(color.Black)
	text.Draw(screen, stageLabel, r.hudFontFace, op)
}
