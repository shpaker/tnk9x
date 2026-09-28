package stage

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/shpaker/tnk9x/internal/adapters/ui"
	"github.com/shpaker/tnk9x/internal/types"
)

// stageResultLabels — подписи пунктов меню итогов
var stageResultLabels = map[types.StageResultItem]string{
	types.StageResultItemNext:   "NEXT STAGE",
	types.StageResultItemRetry:  "RETRY",
	types.StageResultItemLevels: "STAGES",
}

// Раскладка экрана итогов: доли высоты экрана
const (
	resultTitleY  = 0.16
	resultStarsY  = 0.36
	resultStatsY  = 0.46
	resultMenuTop = 0.62
)

// Цвета экрана итогов
var (
	newBestColor    = color.NRGBA{R: 252, G: 196, B: 36, A: 255}
	subtleTextColor = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
)

// DrawStageResult рисует экран итогов: исход, звёзды, время
// и меню действий
func (r *StageRendererAdapter) DrawStageResult(
	screen *ebiten.Image,
	view types.StageResultViewData,
) {
	bounds := screen.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())
	r.lastWidth = width
	r.lastHeight = height
	r.resultItems = view.Items

	vector.FillRect(
		screen, 0, 0, float32(width), float32(height),
		overlayBackdropColor, false,
	)

	title := "VICTORY"
	if !view.Won {
		title = "DEFEAT"
	}
	titleWidth, _ := text.Measure(title, r.fontFace, 0)
	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate((width-titleWidth)/2, height*resultTitleY)
	titleOp.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, title, r.fontFace, titleOp)

	if view.Won {
		ui.DrawStarsRow(
			screen,
			float32(width/2), float32(height*resultStarsY),
			9, 24,
			view.Stars, types.MaxLevelStars,
		)
	}

	seconds := view.ElapsedTicks / 60
	stats := fmt.Sprintf(
		"TIME %d:%02d  LIVES LOST %d",
		seconds/60, seconds%60, view.LivesLost,
	)
	r.drawResultLine(screen, stats, height*resultStatsY, subtleTextColor)
	if view.NewBest {
		r.drawResultLine(
			screen, "NEW BEST",
			height*resultStatsY+float64(r.regularFontSize)+4,
			newBestColor,
		)
	}

	layout := r.resultMenuLayout(height, len(view.Items))
	for i, item := range view.Items {
		rowColor := color.NRGBA{R: 150, G: 150, B: 150, A: 255}
		if i == view.ActiveIndex {
			rowColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		r.drawResultLine(
			screen, stageResultLabels[item], layout.rowTops[i], rowColor,
		)
	}
}

// StageResultHitTest определяет пункт меню итогов по тапу
func (r *StageRendererAdapter) StageResultHitTest(
	pos types.Position,
) (types.StageResultItem, bool) {
	if r.lastHeight <= 0 || len(r.resultItems) == 0 {
		return 0, false
	}
	layout := r.resultMenuLayout(r.lastHeight, len(r.resultItems))
	if pos.Y < layout.menuTop || pos.Y >= layout.menuBottom {
		return 0, false
	}
	for i := len(layout.rowTops) - 1; i > 0; i-- {
		boundary := (layout.rowTops[i-1] + layout.rowHeight +
			layout.rowTops[i]) / 2
		if pos.Y >= boundary {
			return r.resultItems[i], true
		}
	}
	return r.resultItems[0], true
}

// resultMenuLayout — строки меню итогов в нижней части экрана
func (r *StageRendererAdapter) resultMenuLayout(
	height float64,
	rows int,
) pauseMenuLayout {
	layout := r.pauseMenuLayout(height, rows)
	shift := height*resultMenuTop - layout.rowTops[0]
	if rows == 0 {
		shift = 0
	}
	for i := range layout.rowTops {
		layout.rowTops[i] += shift
	}
	layout.menuTop += shift
	layout.menuBottom += shift
	return layout
}

func (r *StageRendererAdapter) drawResultLine(
	screen *ebiten.Image,
	label string,
	top float64,
	lineColor color.NRGBA,
) {
	scale := float64(r.regularFontSize) / float64(r.titleFontSize)
	if scale <= 0 {
		scale = 1
	}
	labelWidth, _ := text.Measure(label, r.fontFace, 0)
	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(
		(float64(screen.Bounds().Dx())-labelWidth*scale)/2, top,
	)
	op.ColorScale.ScaleWithColor(lineColor)
	text.Draw(screen, label, r.fontFace, op)
}
