package settings

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/shpaker/tnk9x/internal/adapters/ui"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// Раскладка страницы «Как играть» в логических координатах 256x224:
// блок из заголовка, цели, таблицы и подсказки центрируется
// по вертикали
const (
	helpTitleGap = 8
	helpGoalGap  = 16
	helpHeadGap  = 6
	helpRowStep  = 14
	helpHintGap  = 18
	helpCellPad  = 3
	// helpBlinkTicks — полупериод мигания подсказки старта
	helpBlinkTicks = 30
)

// helpColumns — левый край подписей и центры столбцов таблицы:
// у одного игрока — клавиатура и геймпад, на двоих — клавиатуры
// обоих игроков и геймпад, на таче — экранные контроллы
type helpColumns struct {
	labelLeft float64
	centers   []float64
}

var (
	helpOnePlayerColumns  = helpColumns{24, []float64{148, 212}}
	helpTwoPlayersColumns = helpColumns{16, []float64{112, 160, 216}}
	helpTouchColumns      = helpColumns{40, []float64{184}}
)

// HelpRendererAdapter рисует страницу «Как играть» поверх уровня
type HelpRendererAdapter struct {
	texts interfaces.ITextsAdapter
	font  ui.MenuFont
}

func NewHelpRendererAdapter(
	texts interfaces.ITextsAdapter,
	fontFace text.Face,
	titleFontSize int,
	regularFontSize int,
) *HelpRendererAdapter {
	return &HelpRendererAdapter{
		texts: texts,
		font: ui.MenuFont{
			Face:            fontFace,
			TitleFontSize:   titleFontSize,
			RegularFontSize: regularFontSize,
		},
	}
}

func (r *HelpRendererAdapter) Draw(
	screen *ebiten.Image,
	view types.HelpViewData,
) {
	columns, header, cells := r.table(view)
	width := float64(screen.Bounds().Dx())
	height := float64(screen.Bounds().Dy())
	lineHeight := float64(r.font.RegularFontSize)

	// Высота блока: заголовок, цель, шапка, строки таблицы, подсказка
	blockHeight := float64(r.font.TitleFontSize) + helpTitleGap +
		lineHeight + helpGoalGap +
		lineHeight + helpHeadGap +
		float64(len(cells)-1)*helpRowStep + lineHeight +
		helpHintGap + lineHeight
	top := math.Round((height - blockHeight) / 2)

	ui.DrawOverlay(screen, r.font, r.texts.Get(types.TextHelpTitle), top)
	top += float64(r.font.TitleFontSize) + helpTitleGap

	r.drawCentered(
		screen, r.texts.Get(types.TextHelpGoal), top, width, controlsRowColor,
	)
	top += lineHeight + helpGoalGap

	for i, label := range header {
		r.drawInColumn(
			screen, label, columns.centers[i], top, controlsHeaderColor,
		)
	}
	top += lineHeight + helpHeadGap

	for _, row := range cells {
		r.font.Draw(
			screen, row.label, columns.labelLeft, top, controlsRowColor,
		)
		for i, value := range row.values {
			if value == "" {
				continue
			}
			center := columns.centers[i]
			// Значение на двоих столбцов — между ними
			if row.span && i+1 < len(columns.centers) {
				center = (center + columns.centers[i+1]) / 2
			}
			r.drawCell(screen, value, center, top)
		}
		top += helpRowStep
	}
	top += lineHeight - helpRowStep + helpHintGap

	if (view.Ticks/helpBlinkTicks)%2 == 0 {
		r.drawCentered(
			screen,
			r.startHint(view),
			top,
			width,
			controlsActiveColor,
		)
	}
}

// helpCells — строка таблицы для отрисовки; span — первое значение
// занимает два столбца
type helpCells struct {
	label  string
	values []string
	span   bool
}

// table — столбцы, шапка и строки для режима страницы
func (r *HelpRendererAdapter) table(
	view types.HelpViewData,
) (helpColumns, []string, []helpCells) {
	if view.TouchActive {
		return helpTouchColumns,
			[]string{r.texts.Get(types.TextHelpTouchColumn)},
			[]helpCells{
				{
					label:  r.texts.Get(types.TextHelpMove),
					values: []string{r.texts.Get(types.TextHelpTouchDPad)},
				},
				{
					label:  r.texts.Get(types.TextControlsFire),
					values: []string{r.texts.Get(types.TextHelpTouchFire)},
				},
				{
					label:  r.texts.Get(types.TextHelpPause),
					values: []string{r.texts.Get(types.TextHelpTouchPause)},
				},
			}
	}

	twoPlayers := view.Players >= types.MaxPlayers
	cells := make([]helpCells, 0, len(view.Rows))
	for _, row := range view.Rows {
		label := r.texts.Get(types.TextHelpPause)
		if row.Kind == types.HelpRowAction {
			label = r.texts.Get(actionLabels[row.Action])
		}
		keys := []string{r.keyLabel(row.Keys[0])}
		span := false
		if twoPlayers {
			// Пауза общая: одна клавиша на оба столбца клавиатуры
			if row.Kind == types.HelpRowPause {
				keys = append(keys, "")
				span = true
			} else {
				keys = append(keys, r.keyLabel(row.Keys[1]))
			}
		}
		cells = append(cells, helpCells{
			label:  label,
			values: append(keys, row.Button),
			span:   span,
		})
	}

	if twoPlayers {
		return helpTwoPlayersColumns, []string{
			r.texts.Get(types.TextHelpPlayer1),
			r.texts.Get(types.TextHelpPlayer2),
			r.texts.Get(types.TextControlsPadColumn),
		}, cells
	}
	return helpOnePlayerColumns, []string{
		r.texts.Get(types.TextControlsKeyColumn),
		r.texts.Get(types.TextControlsPadColumn),
	}, cells
}

// startHint — подсказка старта: на таче — огонь, иначе Enter
func (r *HelpRendererAdapter) startHint(view types.HelpViewData) string {
	button := r.texts.Get(types.TextLevelSelectButtonEnter)
	if view.TouchActive {
		button = r.texts.Get(types.TextLevelSelectButtonFire)
	}
	return r.texts.Format(
		types.TextHelpStart, types.TextArgs{"Button": button},
	)
}

// drawCell — значение на подложке, как ячейка экрана раскладки
func (r *HelpRendererAdapter) drawCell(
	screen *ebiten.Image,
	value string,
	center, top float64,
) {
	cellWidth := math.Round(r.font.TextWidth(value)) + 2*helpCellPad
	vector.FillRect(
		screen,
		float32(math.Round(center-cellWidth/2)), float32(top-2),
		float32(cellWidth), float32(r.font.RegularFontSize+4),
		controlsCellColor, false,
	)
	r.drawInColumn(screen, value, center, top, controlsActiveColor)
}

// keyLabel — подпись клавиши: стрелки на языке интерфейса
func (r *HelpRendererAdapter) keyLabel(name string) string {
	return keyLabel(r.texts, name)
}

func (r *HelpRendererAdapter) drawInColumn(
	screen *ebiten.Image,
	label string,
	center, top float64,
	textColor color.Color,
) {
	left := math.Round(center - r.font.TextWidth(label)/2)
	r.font.Draw(screen, label, left, top, textColor)
}

func (r *HelpRendererAdapter) drawCentered(
	screen *ebiten.Image,
	label string,
	top, width float64,
	textColor color.Color,
) {
	r.drawInColumn(screen, label, width/2, top, textColor)
}
