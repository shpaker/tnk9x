package settings

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/shpaker/tnk9x/internal/adapters/bindings"
	"github.com/shpaker/tnk9x/internal/adapters/ui"
	"github.com/shpaker/tnk9x/internal/types"
)

// Подписи экрана раскладки
var (
	controlsPageLabels = [types.ControlsPagesCount]string{
		"PLAYER 1", "PLAYER 2", "HOTKEYS",
	}
	actionLabels = [types.InputActionsCount]string{
		"UP", "DOWN", "LEFT", "RIGHT", "FIRE",
	}
	hotkeyLabels = [types.HotkeysCount]string{"GRAPHICS", "FULLSCREEN"}
)

// Раскладка таблицы в логических координатах 256x224
const (
	// controlsTitleGap — отступ таблицы от заголовка
	controlsTitleGap  = 12
	controlsRowStep   = 14
	controlsLabelLeft = 24
	keyColumnCenter   = 148
	padColumnCenter   = 212
	controlsCellWidth = 60
	controlsHintTop   = 210
	// controlsBlinkTicks — полупериод мигания ячейки в ожидании
	controlsBlinkTicks = 20
)

// Цвета экрана раскладки
var (
	controlsRowColor    = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	controlsActiveColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	controlsHeaderColor = color.NRGBA{R: 110, G: 110, B: 110, A: 255}
	controlsCellColor   = color.NRGBA{R: 70, G: 70, B: 70, A: 255}
	controlsRejectColor = color.NRGBA{R: 200, G: 70, B: 50, A: 255}
)

// ControlsRendererAdapter рисует экран раскладки: страница игрока
// или хоткеев, колонки клавиатуры и геймпада
type ControlsRendererAdapter struct {
	font ui.MenuFont
}

func NewControlsRendererAdapter(
	fontFace text.Face,
	titleFontSize int,
	regularFontSize int,
) *ControlsRendererAdapter {
	return &ControlsRendererAdapter{
		font: ui.MenuFont{
			Face:            fontFace,
			TitleFontSize:   titleFontSize,
			RegularFontSize: regularFontSize,
		},
	}
}

func (r *ControlsRendererAdapter) Draw(
	screen *ebiten.Image,
	view types.ControlsViewData,
) {
	// Заголовок на той же высоте, что у SETTINGS
	titleTop := ui.MenuTitleTop(screen, r.font)
	ui.DrawOverlay(screen, r.font, "CONTROLS", titleTop)
	width := float64(screen.Bounds().Dx())
	cursor := view.Cursor

	_, isPlayerPage := cursor.Page.Player()
	top := titleTop + float64(r.font.TitleFontSize) + controlsTitleGap
	for i, row := range view.Rows {
		active := i == cursor.Row
		rowColor := controlsRowColor
		if active {
			rowColor = controlsActiveColor
		}

		switch row.Kind {
		case types.ControlsRowPage:
			r.drawCentered(
				screen, "< "+controlsPageLabels[cursor.Page]+" >",
				top, width, rowColor,
			)
			top += controlsRowStep
			// Шапка колонок под заголовком страницы
			r.drawInColumn(
				screen,
				"KEY",
				keyColumnCenter,
				top,
				controlsHeaderColor,
			)
			if isPlayerPage {
				r.drawInColumn(
					screen,
					"PAD",
					padColumnCenter,
					top,
					controlsHeaderColor,
				)
			}
		case types.ControlsRowAction:
			r.font.Draw(
				screen,
				actionLabels[row.Action],
				controlsLabelLeft,
				top,
				rowColor,
			)
			r.drawCell(screen, view, active, types.ControlsDeviceKeyboard,
				bindings.KeyLabel(row.Key), top)
			r.drawCell(screen, view, active, types.ControlsDeviceGamepad,
				row.Button, top)
		case types.ControlsRowHotkey:
			r.font.Draw(
				screen,
				hotkeyLabels[row.Hotkey],
				controlsLabelLeft,
				top,
				rowColor,
			)
			r.drawCell(screen, view, active, types.ControlsDeviceKeyboard,
				bindings.KeyLabel(row.Key), top)
		case types.ControlsRowReset:
			r.drawCentered(screen, "RESET ALL", top, width, rowColor)
		case types.ControlsRowBack:
			r.drawCentered(screen, "BACK", top, width, rowColor)
		}
		top += controlsRowStep
	}

	r.drawHint(screen, cursor, width)
}

// drawCell — значение ячейки; выбранная ячейка на подложке, в
// ожидании нажатия мигает, после отказа подсвечена красным
func (r *ControlsRendererAdapter) drawCell(
	screen *ebiten.Image,
	view types.ControlsViewData,
	activeRow bool,
	device types.ControlsDevice,
	value string,
	top float64,
) {
	center := float64(keyColumnCenter)
	if device == types.ControlsDeviceGamepad {
		center = padColumnCenter
	}
	cursor := view.Cursor
	selected := activeRow && cursor.Column == device
	textColor := controlsRowColor
	if selected {
		cellColor := controlsCellColor
		if cursor.Rejected {
			cellColor = controlsRejectColor
		}
		vector.FillRect(
			screen,
			float32(center-controlsCellWidth/2), float32(top-2),
			controlsCellWidth, float32(r.font.RegularFontSize+4),
			cellColor, false,
		)
		textColor = controlsActiveColor
		if cursor.Capturing {
			value = ""
			if (cursor.Ticks/controlsBlinkTicks)%2 == 0 {
				value = "?"
			}
		}
	}
	r.drawInColumn(screen, value, center, top, textColor)
}

// drawHint — подсказка внизу экрана в режиме ожидания
func (r *ControlsRendererAdapter) drawHint(
	screen *ebiten.Image,
	cursor types.ControlsCursor,
	width float64,
) {
	if !cursor.Capturing {
		return
	}
	hint := "PRESS A KEY  ESC CANCEL"
	if cursor.Column == types.ControlsDeviceGamepad {
		hint = "PRESS A BUTTON  ESC CANCEL"
	}
	hintColor := controlsHeaderColor
	if cursor.Rejected {
		hint = "ALREADY TAKEN"
		hintColor = controlsRejectColor
	}
	r.drawCentered(screen, hint, controlsHintTop, width, hintColor)
}

func (r *ControlsRendererAdapter) drawInColumn(
	screen *ebiten.Image,
	label string,
	center, top float64,
	textColor color.Color,
) {
	r.font.Draw(screen, label, center-r.font.TextWidth(label)/2, top, textColor)
}

func (r *ControlsRendererAdapter) drawCentered(
	screen *ebiten.Image,
	label string,
	top, width float64,
	textColor color.Color,
) {
	r.font.Draw(
		screen,
		label,
		(width-r.font.TextWidth(label))/2,
		top,
		textColor,
	)
}
