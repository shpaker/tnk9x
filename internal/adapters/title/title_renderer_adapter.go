package title

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/tnk9x/internal/types"
)

// TitleRendererAdapter рисует титульный экран: название игры,
// HI-SCORE и меню выбора числа игроков
type TitleRendererAdapter struct {
	fontFace         text.Face
	hudFontFace      text.Face
	titleFontSize    int
	regularFontSize  int
	subtitleFontSize int
	gameTitle        string
}

// TitleRendererDependencies — готовый граф зависимостей рендера
// титула; собирается composition root'ом, все поля обязательны
type TitleRendererDependencies struct {
	FontFace         text.Face
	HUDFontFace      text.Face
	TitleFontSize    int
	RegularFontSize  int
	SubtitleFontSize int
	GameTitle        string
}

func NewTitleRendererAdapter(
	deps TitleRendererDependencies,
) *TitleRendererAdapter {
	return &TitleRendererAdapter{
		fontFace:         deps.FontFace,
		hudFontFace:      deps.HUDFontFace,
		titleFontSize:    deps.TitleFontSize,
		regularFontSize:  deps.RegularFontSize,
		subtitleFontSize: deps.SubtitleFontSize,
		gameTitle:        deps.GameTitle,
	}
}

var (
	titleActiveColor   = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	titleInactiveColor = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	titleSubtleColor   = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
)

func (r *TitleRendererAdapter) DrawAll(
	screen *ebiten.Image,
	view types.TitleViewData,
) {
	screenBounds := screen.Bounds()
	actualWidth := float64(screenBounds.Dx())
	actualHeight := float64(screenBounds.Dy())

	screen.Fill(color.Black)

	hiScoreText := fmt.Sprintf("HI-SCORE %d", view.HiScore)
	r.drawCenteredHUDText(
		screen,
		hiScoreText,
		actualWidth,
		actualHeight/8,
		titleSubtleColor,
	)

	titleWidth, _ := text.Measure(r.gameTitle, r.fontFace, 0)
	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(
		(actualWidth-titleWidth)/2,
		actualHeight/4-float64(r.titleFontSize)/2,
	)
	titleOp.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, r.gameTitle, r.fontFace, titleOp)

	onePlayerColor := titleActiveColor
	twoPlayersColor := titleInactiveColor
	if view.TwoPlayersSelected {
		onePlayerColor = titleInactiveColor
		twoPlayersColor = titleActiveColor
	}

	menuY := actualHeight / 2
	r.drawCenteredHUDText(
		screen,
		"1 PLAYER",
		actualWidth,
		menuY,
		onePlayerColor,
	)
	r.drawCenteredHUDText(
		screen,
		"2 PLAYERS",
		actualWidth,
		menuY+float64(r.regularFontSize)*2,
		twoPlayersColor,
	)

	r.drawCenteredHUDText(
		screen,
		"PRESS ENTER TO START",
		actualWidth,
		actualHeight-float64(r.subtitleFontSize)*2,
		titleSubtleColor,
	)
}

func (r *TitleRendererAdapter) drawCenteredHUDText(
	screen *ebiten.Image,
	value string,
	screenWidth float64,
	y float64,
	clr color.Color,
) {
	textWidth, _ := text.Measure(value, r.hudFontFace, 0)
	op := &text.DrawOptions{}
	op.GeoM.Translate((screenWidth-textWidth)/2, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, value, r.hudFontFace, op)
}
