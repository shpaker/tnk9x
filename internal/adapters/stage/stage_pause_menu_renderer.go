package stage

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/shpaker/tnk9x/internal/types"
)

// pauseMenuLabels — подписи пунктов меню паузы
var pauseMenuLabels = map[types.PauseMenuItem]string{
	types.PauseMenuItemContinue:     "CONTINUE",
	types.PauseMenuItemExitToSelect: "EXIT TO MENU",
}

// pauseMenuLayout — вертикальная раскладка строк меню паузы и границы
// полос тап-зон в логических координатах экрана
type pauseMenuLayout struct {
	rowHeight  float64
	rowTops    []float64
	menuTop    float64
	menuBottom float64
}

// pauseMenuLayout — единый источник вертикальных позиций меню паузы
// для отрисовки и хит-тестов
func (r *StageRendererAdapter) pauseMenuLayout(
	height float64,
	rows int,
) pauseMenuLayout {
	_, textHeight := text.Measure("PAUSED", r.fontFace, 0)
	scale := float64(r.regularFontSize) / float64(r.titleFontSize)
	if scale <= 0 {
		scale = 1
	}
	rowHeight := textHeight * scale
	gap := float64(r.regularFontSize)
	firstTop := (height-rowHeight)/2 + gap

	rowTops := make([]float64, rows)
	for i := range rowTops {
		rowTops[i] = firstTop + float64(i)*(rowHeight+gap)
	}

	menuBottom := firstTop + rowHeight + gap
	if rows > 0 {
		menuBottom = rowTops[rows-1] + rowHeight + gap
	}

	return pauseMenuLayout{
		rowHeight:  rowHeight,
		rowTops:    rowTops,
		menuTop:    firstTop - gap,
		menuBottom: menuBottom,
	}
}

// DrawPauseMenu рисует оверлей паузы с заголовком и пунктами меню;
// активный пункт выделяется белым
func (r *StageRendererAdapter) DrawPauseMenu(
	screen *ebiten.Image,
	view types.PauseMenuViewData,
) {
	bounds := screen.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())
	r.lastWidth = width
	r.lastHeight = height
	r.pauseMenuItems = view.Items

	vector.FillRect(
		screen,
		0,
		0,
		float32(width),
		float32(height),
		overlayBackdropColor,
		false,
	)

	titleText := "PAUSED"
	titleWidth, _ := text.Measure(titleText, r.fontFace, 0)
	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(
		(width-titleWidth)/2,
		height/4-float64(r.titleFontSize)/2,
	)
	titleOp.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, titleText, r.fontFace, titleOp)

	layout := r.pauseMenuLayout(height, len(view.Items))
	scale := float64(r.regularFontSize) / float64(r.titleFontSize)
	if scale <= 0 {
		scale = 1
	}

	for i, item := range view.Items {
		label := pauseMenuLabels[item]
		labelWidth, _ := text.Measure(label, r.fontFace, 0)

		rowColor := color.NRGBA{R: 150, G: 150, B: 150, A: 255}
		if i == view.ActiveIndex {
			rowColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}

		op := &text.DrawOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate((width-labelWidth*scale)/2, layout.rowTops[i])
		op.ColorScale.ScaleWithColor(rowColor)
		text.Draw(screen, label, r.fontFace, op)
	}
}

// PauseMenuHitTest определяет пункт меню паузы по тапу в логических
// координатах экрана; полосы строк занимают всю ширину
func (r *StageRendererAdapter) PauseMenuHitTest(
	pos types.Position,
) (types.PauseMenuItem, bool) {
	if r.lastWidth <= 0 || r.lastHeight <= 0 ||
		len(r.pauseMenuItems) == 0 {
		return 0, false
	}
	layout := r.pauseMenuLayout(r.lastHeight, len(r.pauseMenuItems))
	if pos.Y < layout.menuTop || pos.Y >= layout.menuBottom {
		return 0, false
	}
	for i := len(layout.rowTops) - 1; i > 0; i-- {
		boundary := (layout.rowTops[i-1] + layout.rowHeight +
			layout.rowTops[i]) / 2
		if pos.Y >= boundary {
			return r.pauseMenuItems[i], true
		}
	}

	return r.pauseMenuItems[0], true
}
