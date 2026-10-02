package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Цвета меню-оверлеев
var (
	// OverlayBackdropColor — полупрозрачная подложка оверлеев:
	// пауза, итоги уровня, меню и настройки
	OverlayBackdropColor = color.NRGBA{R: 40, G: 40, B: 40, A: 240}
	menuRowColor         = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	menuActiveRowColor   = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
)

// MenuFont — шрифт меню: заголовок рисуется в полный размер,
// строки — уменьшенными до основного размера
type MenuFont struct {
	Face            text.Face
	TitleFontSize   int
	RegularFontSize int
}

// MenuRow — строка меню: подпись и необязательное значение
// в правой колонке
type MenuRow struct {
	Label string
	Value string
}

// menuValueGap — зазор между колонками подписей и значений
const menuValueGap = 16

// DrawMenu рисует оверлей меню: подложку, заголовок и строки;
// активная строка белая, её значение обрамлено стрелками
func DrawMenu(
	screen *ebiten.Image,
	font MenuFont,
	title string,
	rows []MenuRow,
	activeIndex int,
) {
	DrawOverlay(screen, font, title, MenuTitleTop(screen, font))
	drawMenuRows(screen, font, rows, activeIndex)
}

// MenuTitleTop — высота заголовка меню и подменю: у всех оверлеев
// заголовок на одной высоте
func MenuTitleTop(screen *ebiten.Image, font MenuFont) float64 {
	height := float64(screen.Bounds().Dy())
	return height/4 - float64(font.TitleFontSize)/2
}

// drawMenuRows — строки меню под заголовком
func drawMenuRows(
	screen *ebiten.Image,
	font MenuFont,
	rows []MenuRow,
	activeIndex int,
) {
	bounds := screen.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())

	// Высота строки — по метрикам шрифта: заголовка может не быть
	_, titleHeight := text.Measure("M", font.Face, 0)
	scale := font.scale()
	rowHeight := titleHeight * scale
	gap := float64(font.RegularFontSize)
	firstTop := (height-rowHeight)/2 + gap

	labelsWidth, valuesWidth := font.columnWidths(rows)
	left := (width - labelsWidth - menuValueGap - valuesWidth) / 2

	for i, row := range rows {
		top := firstTop + float64(i)*(rowHeight+gap)
		rowColor := menuRowColor
		if i == activeIndex {
			rowColor = menuActiveRowColor
		}

		// Строка без значения центрируется по всей ширине
		if row.Value == "" {
			x := (width - font.TextWidth(row.Label)) / 2
			font.Draw(screen, row.Label, x, top, rowColor)
			continue
		}

		font.Draw(screen, row.Label, left, top, rowColor)
		value := row.Value
		if i == activeIndex {
			value = "< " + value + " >"
		}
		valueLeft := left + labelsWidth + menuValueGap +
			(valuesWidth-font.TextWidth(value))/2
		font.Draw(screen, value, valueLeft, top, rowColor)
	}
}

// DrawOverlay рисует подложку оверлея и заголовок крупным шрифтом
// по центру на высоте titleTop
func DrawOverlay(
	screen *ebiten.Image,
	font MenuFont,
	title string,
	titleTop float64,
) {
	bounds := screen.Bounds()
	width := float64(bounds.Dx())

	vector.FillRect(
		screen,
		0,
		0,
		float32(width),
		float32(bounds.Dy()),
		OverlayBackdropColor,
		false,
	)
	drawTitle(screen, font, title, titleTop)
}

// drawTitle — заголовок крупным шрифтом по центру на высоте titleTop
func drawTitle(
	screen *ebiten.Image,
	font MenuFont,
	title string,
	titleTop float64,
) {
	width := float64(screen.Bounds().Dx())
	titleWidth, _ := text.Measure(title, font.Face, 0)
	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate((width-titleWidth)/2, titleTop)
	titleOp.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, title, font.Face, titleOp)
}

// columnWidths — ширина колонки подписей строк со значениями
// и колонки значений с запасом под стрелки активной строки
func (f MenuFont) columnWidths(rows []MenuRow) (float64, float64) {
	var labelsWidth, valuesWidth float64
	for _, row := range rows {
		if row.Value == "" {
			continue
		}
		labelsWidth = max(labelsWidth, f.TextWidth(row.Label))
		valuesWidth = max(valuesWidth, f.TextWidth("< "+row.Value+" >"))
	}
	return labelsWidth, valuesWidth
}

// scale — масштаб строк относительно шрифта заголовка
func (f MenuFont) scale() float64 {
	scale := float64(f.RegularFontSize) / float64(f.TitleFontSize)
	if scale <= 0 {
		return 1
	}
	return scale
}

// TextWidth — ширина строки основного размера
func (f MenuFont) TextWidth(label string) float64 {
	width, _ := text.Measure(label, f.Face, 0)
	return width * f.scale()
}

// Draw рисует строку основного размера с левым верхним углом (x, y)
func (f MenuFont) Draw(
	screen *ebiten.Image,
	label string,
	x, y float64,
	textColor color.Color,
) {
	scale := f.scale()
	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, label, f.Face, op)
}
