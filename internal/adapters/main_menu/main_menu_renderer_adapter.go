// Package main_menu рисует главное меню: заголовок игры, пункты
// и версию.
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

// versionMargin — отступ строки версии от правого нижнего угла
const versionMargin = 6

var versionColor = color.NRGBA{R: 110, G: 110, B: 110, A: 255}

// MainMenuRendererAdapter рисует главное меню
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
	rows := make([]ui.MenuRow, len(view.Items))
	for i, item := range view.Items {
		rows[i] = ui.MenuRow{Label: mainMenuLabels[item]}
	}
	// Подложка как у паузы и итогов; название игры — блоки сцены
	ui.DrawMenu(screen, r.font, "", rows, view.ActiveIndex)

	bounds := screen.Bounds()
	version := "V" + view.Version
	r.font.Draw(
		screen,
		version,
		float64(bounds.Dx())-r.font.TextWidth(version)-versionMargin,
		float64(bounds.Dy()-r.font.RegularFontSize-versionMargin),
		versionColor,
	)

	ui.DrawFade(screen, view.Fade)
}
