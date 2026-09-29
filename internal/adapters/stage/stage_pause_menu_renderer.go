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
	types.PauseMenuItemExitToLevels: "EXIT TO LEVELS",
}

// pauseMenuLabel — подпись пункта; у пункта графики она показывает
// текущий режим
func pauseMenuLabel(
	item types.PauseMenuItem,
	view types.PauseMenuViewData,
) string {
	if item == types.PauseMenuItemGraphics {
		return types.GraphicsLabel(view.EffectsEnabled)
	}
	return pauseMenuLabels[item]
}

// pauseMenuLayout — вертикальная раскладка строк меню паузы
// в логических координатах экрана
type pauseMenuLayout struct {
	rowHeight float64
	rowTops   []float64
}

// pauseMenuLayout — вертикальные позиции строк меню паузы
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

	return pauseMenuLayout{
		rowHeight: rowHeight,
		rowTops:   rowTops,
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
		label := pauseMenuLabel(item, view)
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
