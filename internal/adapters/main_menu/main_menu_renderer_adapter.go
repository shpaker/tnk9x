// Package main_menu рисует главное меню поверх живой сцены: пункты
// с чёрной обводкой.
package main_menu

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/tnk9x/internal/adapters/ui"
	"github.com/shpaker/tnk9x/internal/types"
)

// mainMenuLabels — подписи пунктов главного меню
var mainMenuLabels = map[types.MainMenuItem]string{
	types.MainMenuItemOnePlayer:  "1 PLAYER",
	types.MainMenuItemTwoPlayers: "2 PLAYERS",
	types.MainMenuItemSettings:   "SETTINGS",
	types.MainMenuItemQuit:       "QUIT",
}

// Раскладка меню в логических координатах 256x224: пункты под
// названием из блоков сцены (поле с 8px, название — ряды 6..10)
const (
	menuTop     = 124
	menuRowStep = 16
)

// Цвета меню
var (
	rowColor       = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	activeRowColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	outlineColor   = color.NRGBA{A: 255}
)

// outlineOffsets — сдвиги чёрной обводки в 1 px вокруг букв
var outlineOffsets = [][2]float64{
	{-1, -1},
	{0, -1},
	{1, -1},
	{-1, 0},
	{1, 0},
	{-1, 1},
	{0, 1},
	{1, 1},
}

// MainMenuRendererAdapter рисует главное меню без затемнения сцены:
// текст с чёрной обводкой читается и поверх кирпича, и поверх танков
type MainMenuRendererAdapter struct {
	font ui.MenuFont
}

func NewMainMenuRendererAdapter(
	fontFace text.Face,
	titleFontSize int,
	regularFontSize int,
) *MainMenuRendererAdapter {
	return &MainMenuRendererAdapter{
		font: ui.MenuFont{
			Face:            fontFace,
			TitleFontSize:   titleFontSize,
			RegularFontSize: regularFontSize,
		},
	}
}

func (r *MainMenuRendererAdapter) Draw(
	screen *ebiten.Image,
	view types.MainMenuViewData,
) {
	bounds := screen.Bounds()
	width := float64(bounds.Dx())

	for i, item := range view.Items {
		label := mainMenuLabels[item]
		textColor := rowColor
		if i == view.ActiveIndex {
			textColor = activeRowColor
		}
		r.drawOutlined(
			screen, label,
			(width-r.font.TextWidth(label))/2,
			float64(menuTop+i*menuRowStep),
			textColor,
		)
	}

	ui.DrawFade(screen, view.Fade)
}

// drawOutlined — строка с чёрной обводкой в 1 px
func (r *MainMenuRendererAdapter) drawOutlined(
	screen *ebiten.Image,
	label string,
	x, y float64,
	textColor color.Color,
) {
	for _, offset := range outlineOffsets {
		r.font.Draw(screen, label, x+offset[0], y+offset[1], outlineColor)
	}
	r.font.Draw(screen, label, x, y, textColor)
}
