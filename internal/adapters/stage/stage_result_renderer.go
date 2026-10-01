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
	types.StageResultItemNext:     "NEXT STAGE",
	types.StageResultItemContinue: "CONTINUE",
	types.StageResultItemRetry:    "RETRY",
	types.StageResultItemLevels:   "STAGES",
}

// resultCaption — пояснение под пунктом меню итогов: текст
// и ряд звёзд за ним
type resultCaption struct {
	label string
	stars uint
}

// stageResultCaptions — пояснения под пунктами меню итогов:
// CONTINUE отличается от NEXT STAGE переносом и потолком звёзд
var stageResultCaptions = map[types.StageResultItem]resultCaption{
	types.StageResultItemContinue: {
		label: "KEEP LIVES & TANK, MAX",
		stars: types.MaxCarryOverStars,
	},
}

// Раскладка пояснения: зазор под пунктом и ряд звёзд после текста
const (
	resultCaptionGap  = 3
	captionStarRadius = 3.5
	captionStarStep   = 9
	captionStarsGap   = 6
)

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
	captionColor    = color.NRGBA{R: 110, G: 110, B: 110, A: 255}
)

// resultTitleDrop — на сколько пикселей заголовок опускается
// при появлении
const resultTitleDrop = 6

// DrawStageResult рисует экран итогов поэтапно, по view.Reveal:
// подложка, исход, звёзды по одной, время и заметки, меню действий
func (r *StageRendererAdapter) DrawStageResult(
	screen *ebiten.Image,
	view types.StageResultViewData,
) {
	bounds := screen.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())
	reveal := view.Reveal

	vector.FillRect(
		screen, 0, 0, float32(width), float32(height),
		withAlpha(ui.OverlayBackdropColor, reveal.Backdrop), false,
	)

	if reveal.Title > 0 {
		title := "VICTORY"
		if !view.Won {
			title = "DEFEAT"
		}
		titleWidth, _ := text.Measure(title, r.fontFace, 0)
		titleOp := &text.DrawOptions{}
		titleOp.GeoM.Translate(
			(width-titleWidth)/2,
			height*resultTitleY-resultTitleDrop*(1-reveal.Title),
		)
		titleOp.ColorScale.ScaleWithColor(
			withAlpha(color.NRGBA{R: 255, G: 255, B: 255, A: 255}, reveal.Title),
		)
		text.Draw(screen, title, r.fontFace, titleOp)
	}

	if view.Won && reveal.Title >= 1 {
		ui.DrawStarsRow(
			screen,
			float32(width/2), float32(height*resultStarsY),
			9, 24,
			reveal.Stars, types.MaxLevelStars,
		)
	}

	if reveal.Stats > 0 {
		r.drawResultStats(screen, view, height, reveal.Stats)
	}
	if !reveal.Menu {
		return
	}

	layout := r.resultMenuLayout(height, view.Items)
	for i, item := range view.Items {
		rowColor := color.NRGBA{R: 150, G: 150, B: 150, A: 255}
		if i == view.ActiveIndex {
			rowColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		r.drawResultLine(
			screen, stageResultLabels[item], layout.rowTops[i], rowColor,
		)
		if caption, ok := stageResultCaptions[item]; ok {
			r.drawResultCaption(
				screen, caption, layout.rowTops[i]+layout.captionOffset,
			)
		}
	}
}

// drawResultStats — время, потерянные жизни и заметки с прозрачностью
// появления alpha
func (r *StageRendererAdapter) drawResultStats(
	screen *ebiten.Image,
	view types.StageResultViewData,
	height, alpha float64,
) {
	seconds := view.ElapsedTicks / 60
	stats := fmt.Sprintf(
		"TIME %d:%02d  LIVES LOST %d",
		seconds/60, seconds%60, view.LivesLost,
	)
	r.drawResultLine(
		screen, stats, height*resultStatsY, withAlpha(subtleTextColor, alpha),
	)
	lineStep := float64(r.regularFontSize) + 4
	noteTop := height*resultStatsY + lineStep
	if view.NewBest {
		r.drawResultLine(
			screen, "NEW BEST", noteTop, withAlpha(newBestColor, alpha),
		)
		noteTop += lineStep
	}
	if view.CarriedOver {
		r.drawResultLine(
			screen,
			fmt.Sprintf("CARRY-OVER: MAX %d STARS", types.MaxCarryOverStars),
			noteTop, withAlpha(subtleTextColor, alpha),
		)
	}
}

// withAlpha — цвет с прозрачностью, умноженной на alpha 0..1
func withAlpha(c color.NRGBA, alpha float64) color.NRGBA {
	c.A = uint8(float64(c.A) * min(max(alpha, 0), 1))
	return c
}

// resultMenuLayout — вертикальная раскладка строк меню итогов
// в логических координатах экрана
type resultMenuLayout struct {
	rowTops []float64
	// captionOffset — сдвиг пояснения от верха его пункта
	captionOffset float64
}

// resultMenuLayout — строки меню итогов в нижней части экрана;
// пункт с пояснением занимает дополнительную строку
func (r *StageRendererAdapter) resultMenuLayout(
	height float64,
	items []types.StageResultItem,
) resultMenuLayout {
	_, textHeight := text.Measure("PAUSED", r.fontFace, 0)
	scale := float64(r.regularFontSize) / float64(r.titleFontSize)
	if scale <= 0 {
		scale = 1
	}
	rowHeight := textHeight * scale
	gap := float64(r.regularFontSize)
	captionOffset := rowHeight + resultCaptionGap

	rowTops := make([]float64, len(items))
	top := height * resultMenuTop
	for i, item := range items {
		rowTops[i] = top
		top += rowHeight + gap
		if _, ok := stageResultCaptions[item]; ok {
			top += captionOffset
		}
	}
	return resultMenuLayout{
		rowTops:       rowTops,
		captionOffset: captionOffset,
	}
}

// drawResultCaption рисует пояснение пункта по центру: тусклый текст
// и за ним ряд звёзд
func (r *StageRendererAdapter) drawResultCaption(
	screen *ebiten.Image,
	caption resultCaption,
	top float64,
) {
	scale := float64(r.regularFontSize) / float64(r.titleFontSize)
	if scale <= 0 {
		scale = 1
	}
	labelWidth, _ := text.Measure(caption.label, r.fontFace, 0)
	labelWidth *= scale
	starsWidth := float64(captionStarStep*(caption.stars-1)) +
		2*captionStarRadius
	left := (float64(screen.Bounds().Dx()) -
		labelWidth - captionStarsGap - starsWidth) / 2

	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(left, top)
	op.ColorScale.ScaleWithColor(captionColor)
	text.Draw(screen, caption.label, r.fontFace, op)

	ui.DrawStarsRow(
		screen,
		float32(left+labelWidth+captionStarsGap+starsWidth/2),
		float32(top+float64(r.regularFontSize)/2),
		captionStarRadius, captionStarStep,
		caption.stars, caption.stars,
	)
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
