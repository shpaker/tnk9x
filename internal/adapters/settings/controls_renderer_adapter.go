package settings

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/shpaker/tnk9x/internal/adapters/bindings"
	"github.com/shpaker/tnk9x/internal/adapters/ui"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// Подписи экрана раскладки
var (
	controlsPageLabels = [types.ControlsPagesCount]types.TextKey{
		types.TextControlsPlayer1,
		types.TextControlsPlayer2,
		types.TextControlsHotkeys,
	}
	actionLabels = [types.InputActionsCount]types.TextKey{
		types.TextControlsUp,
		types.TextControlsDown,
		types.TextControlsLeft,
		types.TextControlsRight,
		types.TextControlsFire,
	}
	hotkeyLabels = [types.HotkeysCount]types.TextKey{
		types.TextControlsHotkeyGraphics,
		types.TextControlsHotkeyFullscreen,
	}
	// arrowKeyLabels — стрелки подписываются на языке интерфейса,
	// остальные клавиши — как на клавиатуре
	arrowKeyLabels = map[string]types.TextKey{
		"ArrowUp":    types.TextKeysUp,
		"ArrowDown":  types.TextKeysDown,
		"ArrowLeft":  types.TextKeysLeft,
		"ArrowRight": types.TextKeysRight,
	}
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
	texts interfaces.ITextsAdapter
	font  ui.MenuFont
	// hits — строки последней отрисовки для мыши и тапов
	hits ui.HitAreas
}

func NewControlsRendererAdapter(
	texts interfaces.ITextsAdapter,
	fontFace text.Face,
	titleFontSize int,
	regularFontSize int,
) *ControlsRendererAdapter {
	return &ControlsRendererAdapter{
		texts: texts,
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
	ui.DrawOverlay(
		screen, r.font, r.texts.Get(types.TextControlsTitle), titleTop,
	)
	width := float64(screen.Bounds().Dx())
	cursor := view.Cursor

	_, isPlayerPage := cursor.Page.Player()
	top := titleTop + float64(r.font.TitleFontSize) + controlsTitleGap
	rowHeight := float64(r.font.RegularFontSize)
	r.hits.Reset()
	for i, row := range view.Rows {
		r.hits.Add(ui.RowRect(
			0, width, top, rowHeight, controlsRowStep-rowHeight,
		))
		active := i == cursor.Row
		rowColor := controlsRowColor
		if active {
			rowColor = controlsActiveColor
		}

		switch row.Kind {
		case types.ControlsRowPage:
			r.drawCentered(
				screen, "< "+r.texts.Get(controlsPageLabels[cursor.Page])+" >",
				top, width, rowColor,
			)
			top += controlsRowStep
			// Шапка колонок под заголовком страницы
			r.drawInColumn(
				screen,
				r.texts.Get(types.TextControlsKeyColumn),
				keyColumnCenter,
				top,
				controlsHeaderColor,
			)
			if isPlayerPage {
				r.drawInColumn(
					screen,
					r.texts.Get(types.TextControlsPadColumn),
					padColumnCenter,
					top,
					controlsHeaderColor,
				)
			}
		case types.ControlsRowAction:
			r.font.Draw(
				screen,
				r.texts.Get(actionLabels[row.Action]),
				controlsLabelLeft,
				top,
				rowColor,
			)
			r.drawCell(screen, view, active, types.ControlsDeviceKeyboard,
				r.keyLabel(row.Key), top)
			r.drawCell(screen, view, active, types.ControlsDeviceGamepad,
				row.Button, top)
		case types.ControlsRowHotkey:
			r.font.Draw(
				screen,
				r.texts.Get(hotkeyLabels[row.Hotkey]),
				controlsLabelLeft,
				top,
				rowColor,
			)
			r.drawCell(screen, view, active, types.ControlsDeviceKeyboard,
				r.keyLabel(row.Key), top)
		case types.ControlsRowReset:
			r.drawCentered(
				screen, r.texts.Get(types.TextControlsResetAll),
				top, width, rowColor,
			)
		case types.ControlsRowBack:
			r.drawCentered(
				screen, r.texts.Get(types.TextControlsBack),
				top, width, rowColor,
			)
		}
		top += controlsRowStep
	}

	r.drawHint(screen, cursor, width)
}

// HitCell — строка последней отрисовки под точкой и колонка: правее
// середины между колонками — геймпад, левее — клавиатура
func (r *ControlsRendererAdapter) HitCell(
	position types.Position,
) (int, types.ControlsDevice, bool) {
	row, ok := r.hits.Hit(position)
	column := types.ControlsDeviceKeyboard
	if position.X >= (keyColumnCenter+padColumnCenter)/2 {
		column = types.ControlsDeviceGamepad
	}
	return row, column, ok
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
	hint := r.texts.Get(types.TextControlsPressKey)
	if cursor.Column == types.ControlsDeviceGamepad {
		hint = r.texts.Get(types.TextControlsPressButton)
	}
	hintColor := controlsHeaderColor
	if cursor.Rejected {
		hint = r.texts.Get(types.TextControlsAlreadyTaken)
		hintColor = controlsRejectColor
	}
	r.drawCentered(screen, hint, controlsHintTop, width, hintColor)
}

// keyLabel — подпись клавиши в ячейке раскладки
func (r *ControlsRendererAdapter) keyLabel(name string) string {
	if key, ok := arrowKeyLabels[name]; ok {
		return r.texts.Get(key)
	}
	return bindings.KeyLabel(name)
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
