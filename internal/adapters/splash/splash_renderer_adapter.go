// Package splash рисует сплеш-экран: логотип, полоску загрузки
// ресурсов и затемнение перехода.
package splash

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/shpaker/tnk9x/internal/adapters/ui"
	"github.com/shpaker/tnk9x/internal/types"
)

// Раскладка сплеша в логических координатах 256x224
const (
	logoCenterY    = 96
	progressTop    = 140
	progressWidth  = 128
	progressHeight = 2
)

// Цвета сплеша
var (
	progressBackColor = color.NRGBA{R: 50, G: 50, B: 50, A: 255}
	progressFillColor = color.NRGBA{R: 220, G: 220, B: 220, A: 255}
	textLogoColor     = color.NRGBA{R: 252, G: 160, B: 68, A: 255}
	textShadowColor   = color.NRGBA{R: 120, G: 30, B: 10, A: 255}
)

// Logo — логотип сплеша с центром в (centerX, centerY)
type Logo interface {
	Draw(screen *ebiten.Image, centerX, centerY float64)
}

// ImageLogo — логотип-картинка от дизайнера, пиксель-арт 1:1
type ImageLogo struct {
	image *ebiten.Image
}

func NewImageLogo(img image.Image) *ImageLogo {
	return &ImageLogo{image: ebiten.NewImageFromImage(img)}
}

func (l *ImageLogo) Draw(screen *ebiten.Image, centerX, centerY float64) {
	bounds := l.image.Bounds()
	op := &ebiten.DrawImageOptions{}
	// Целые координаты — пиксели логотипа совпадают с пикселями экрана
	op.GeoM.Translate(
		math.Round(centerX-float64(bounds.Dx())/2),
		math.Round(centerY-float64(bounds.Dy())/2),
	)
	screen.DrawImage(l.image, op)
}

// TextLogo — заглушка, пока нет картинки: надпись крупным шрифтом
// с тенью
type TextLogo struct {
	face  text.Face
	label string
}

func NewTextLogo(face text.Face, label string) *TextLogo {
	return &TextLogo{face: face, label: label}
}

func (l *TextLogo) Draw(screen *ebiten.Image, centerX, centerY float64) {
	width, height := text.Measure(l.label, l.face, 0)
	x := math.Round(centerX - width/2)
	y := math.Round(centerY - height/2)
	for _, layer := range []struct {
		offset float64
		color  color.Color
	}{{2, textShadowColor}, {0, textLogoColor}} {
		op := &text.DrawOptions{}
		op.GeoM.Translate(x+layer.offset, y+layer.offset)
		op.ColorScale.ScaleWithColor(layer.color)
		text.Draw(screen, l.label, l.face, op)
	}
}

// SplashRendererAdapter рисует сплеш на чёрном фоне
type SplashRendererAdapter struct {
	logo Logo
}

func NewSplashRendererAdapter(logo Logo) *SplashRendererAdapter {
	return &SplashRendererAdapter{logo: logo}
}

func (r *SplashRendererAdapter) Draw(
	screen *ebiten.Image,
	view types.SplashViewData,
) {
	screen.Fill(color.Black)
	width := float64(screen.Bounds().Dx())
	r.logo.Draw(screen, width/2, logoCenterY)

	left := float32((width - progressWidth) / 2)
	vector.FillRect(
		screen, left, progressTop,
		progressWidth, progressHeight,
		progressBackColor, false,
	)
	vector.FillRect(
		screen, left, progressTop,
		float32(progressWidth*min(max(view.Progress, 0), 1)), progressHeight,
		progressFillColor, false,
	)

	ui.DrawFade(screen, view.Fade)
}
