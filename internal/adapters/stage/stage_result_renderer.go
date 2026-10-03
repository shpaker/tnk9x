package stage

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/shpaker/tnk9x/internal/adapters/ui"
	"github.com/shpaker/tnk9x/internal/types"
)

// stageResultLabels — подписи пунктов меню итогов
var stageResultLabels = map[types.StageResultItem]types.TextKey{
	types.StageResultItemNext:       types.TextResultNextStage,
	types.StageResultItemContinue:   types.TextResultContinue,
	types.StageResultItemRetry:      types.TextResultRetry,
	types.StageResultItemLevels:     types.TextResultStages,
	types.StageResultItemRevive:     types.TextResultRevive,
	types.StageResultItemBoostNext:  types.TextResultBoostNext,
	types.StageResultItemBoostRetry: types.TextResultBoostRetry,
}

// resultCaption — пояснение под пунктом меню итогов: текст
// и ряд звёзд за ним
type resultCaption struct {
	label types.TextKey
	stars uint
}

// boostCaption — пояснение усиленного переноса: прокачка и жизнь
// сверху, потолок звёзд как у переноса
var boostCaption = resultCaption{
	label: types.TextResultBoostCaption,
	stars: types.MaxCarryOverStars,
}

// stageResultCaptions — пояснения под пунктами меню итогов:
// CONTINUE отличается от NEXT STAGE переносом и потолком звёзд,
// BOOST — усилением и тем же потолком
var stageResultCaptions = map[types.StageResultItem]resultCaption{
	types.StageResultItemContinue: {
		label: types.TextResultKeepCaption,
		stars: types.MaxCarryOverStars,
	},
	types.StageResultItemBoostNext:  boostCaption,
	types.StageResultItemBoostRetry: boostCaption,
}

// Метка пункта за рекламу: рамка с текстом справа от подписи
const (
	rewardBadgeGap     = 6
	rewardBadgePadding = 2
)

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

// resultMenuBottomMargin — отступ меню итогов от низа экрана
const resultMenuBottomMargin = 4

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
		title := r.texts.Get(types.TextResultVictory)
		if !view.Won {
			title = r.texts.Get(types.TextResultDefeat)
		}
		titleWidth, _ := text.Measure(title, r.fontFace, 0)
		titleOp := &text.DrawOptions{}
		titleOp.GeoM.Translate(
			(width-titleWidth)/2,
			height*resultTitleY-resultTitleDrop*(1-reveal.Title),
		)
		titleOp.ColorScale.ScaleWithColor(
			withAlpha(
				color.NRGBA{R: 255, G: 255, B: 255, A: 255},
				reveal.Title,
			),
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
	r.resultHits.Reset()
	if !reveal.Menu {
		return
	}

	layout := r.resultMenuLayout(height, view.Items)
	for _, rect := range layout.hitRects(width) {
		r.resultHits.Add(rect)
	}
	for i, item := range view.Items {
		rowColor := color.NRGBA{R: 150, G: 150, B: 150, A: 255}
		if i == view.ActiveIndex {
			rowColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		if item.IsRewarded() {
			r.drawRewardedResultLine(
				screen, r.texts.Get(stageResultLabels[item]),
				r.rewardBadge(view), layout.rowTops[i], rowColor,
			)
		} else {
			r.drawResultLine(
				screen, r.texts.Get(stageResultLabels[item]),
				layout.rowTops[i], rowColor,
			)
		}
		if caption, ok := stageResultCaptions[item]; ok {
			r.drawResultCaption(
				screen, caption, layout.rowTops[i]+layout.captionOffset,
			)
		}
	}
}

// HitResultRow — пункт меню итогов последней отрисовки под точкой;
// до появления меню пунктов нет
func (r *StageRendererAdapter) HitResultRow(
	position types.Position,
) (int, bool) {
	return r.resultHits.Hit(position)
}

// drawResultStats — время, потерянные жизни и заметки с прозрачностью
// появления alpha
func (r *StageRendererAdapter) drawResultStats(
	screen *ebiten.Image,
	view types.StageResultViewData,
	height, alpha float64,
) {
	seconds := view.ElapsedTicks / 60
	stats := r.texts.Format(types.TextResultStats, types.TextArgs{
		"Time":  fmt.Sprintf("%d:%02d", seconds/60, seconds%60),
		"Lives": view.LivesLost,
	})
	r.drawResultLine(
		screen, stats, height*resultStatsY, withAlpha(subtleTextColor, alpha),
	)
	lineStep := float64(r.regularFontSize) + 4
	noteTop := height*resultStatsY + lineStep
	if view.NewBest {
		r.drawResultLine(
			screen, r.texts.Get(types.TextResultNewBest), noteTop,
			withAlpha(newBestColor, alpha),
		)
		noteTop += lineStep
	}
	if view.CarriedOver {
		r.drawResultLine(
			screen,
			r.texts.Plural(
				types.TextResultCarryOver, int(types.MaxCarryOverStars), nil,
			),
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
	// rowBottoms — низ пункта вместе с пояснением
	rowBottoms []float64
	// captionOffset — сдвиг пояснения от верха его пункта
	captionOffset float64
	// edgeGap — зазор над первым и под последним пунктом для хитов
	edgeGap float64
}

// resultMenuMetrics — размеры строк меню итогов в логических
// координатах экрана: меню занимает полосу от top до bottom
type resultMenuMetrics struct {
	top       float64
	bottom    float64
	rowHeight float64
	gap       float64
}

// resultMenuLayout — строки меню итогов в нижней части экрана
func (r *StageRendererAdapter) resultMenuLayout(
	height float64,
	items []types.StageResultItem,
) resultMenuLayout {
	_, textHeight := text.Measure("PAUSED", r.fontFace, 0)
	scale := float64(r.regularFontSize) / float64(r.titleFontSize)
	if scale <= 0 {
		scale = 1
	}
	return layoutResultMenu(resultMenuMetrics{
		top:       height * resultMenuTop,
		bottom:    height - resultMenuBottomMargin,
		rowHeight: textHeight * scale,
		gap:       float64(r.regularFontSize),
	}, items)
}

// layoutResultMenu раскладывает строки сверху вниз; пункт
// с пояснением занимает дополнительную строку. Зазор между строками
// ужимается, только если меню иначе не помещается в свою полосу
func layoutResultMenu(
	metrics resultMenuMetrics,
	items []types.StageResultItem,
) resultMenuLayout {
	captionOffset := metrics.rowHeight + resultCaptionGap

	gap := metrics.gap
	if len(items) > 1 {
		content := float64(len(items)) * metrics.rowHeight
		for _, item := range items {
			if _, ok := stageResultCaptions[item]; ok {
				content += captionOffset
			}
		}
		fit := (metrics.bottom - metrics.top - content) /
			float64(len(items)-1)
		gap = max(0, min(gap, fit))
	}

	rowTops := make([]float64, len(items))
	rowBottoms := make([]float64, len(items))
	top := metrics.top
	for i, item := range items {
		rowTops[i] = top
		top += metrics.rowHeight
		if _, ok := stageResultCaptions[item]; ok {
			top += captionOffset
		}
		rowBottoms[i] = top
		top += gap
	}
	return resultMenuLayout{
		rowTops:       rowTops,
		rowBottoms:    rowBottoms,
		captionOffset: captionOffset,
		edgeGap:       metrics.gap,
	}
}

// hitRects — прямоугольники пунктов на всю ширину экрана: граница
// между соседними пунктами посередине зазора, мёртвых зон нет
func (l resultMenuLayout) hitRects(width float64) []image.Rectangle {
	rects := make([]image.Rectangle, len(l.rowTops))
	for i := range l.rowTops {
		upper := l.rowTops[i] - l.edgeGap/2
		if i > 0 {
			upper = (l.rowBottoms[i-1] + l.rowTops[i]) / 2
		}
		lower := l.rowBottoms[i] + l.edgeGap/2
		if i < len(l.rowTops)-1 {
			lower = (l.rowBottoms[i] + l.rowTops[i+1]) / 2
		}
		rects[i] = image.Rect(0, int(upper), int(width), int(lower))
	}
	return rects
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
	label := r.texts.Get(caption.label)
	labelWidth, _ := text.Measure(label, r.fontFace, 0)
	labelWidth *= scale
	starsWidth := float64(captionStarStep*(caption.stars-1)) +
		2*captionStarRadius
	left := (float64(screen.Bounds().Dx()) -
		labelWidth - captionStarsGap - starsWidth) / 2

	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(left, top)
	op.ColorScale.ScaleWithColor(captionColor)
	text.Draw(screen, label, r.fontFace, op)

	ui.DrawStarsRow(
		screen,
		float32(left+labelWidth+captionStarsGap+starsWidth/2),
		float32(top+float64(r.regularFontSize)/2),
		captionStarRadius, captionStarStep,
		caption.stars, caption.stars,
	)
}

// rewardBadge — метка пунктов за рекламу: чем они оплачиваются
func (r *StageRendererAdapter) rewardBadge(
	view types.StageResultViewData,
) string {
	if view.UseTokens {
		return r.texts.Get(types.TextResultToken)
	}
	return r.texts.Get(types.TextResultAd)
}

// drawRewardedResultLine рисует пункт за рекламу: подпись и метку
// оплаты (реклама или жетон) справа от неё, вместе по центру
func (r *StageRendererAdapter) drawRewardedResultLine(
	screen *ebiten.Image,
	label string,
	badgeLabel string,
	top float64,
	lineColor color.NRGBA,
) {
	scale := float64(r.regularFontSize) / float64(r.titleFontSize)
	if scale <= 0 {
		scale = 1
	}
	labelWidth, _ := text.Measure(label, r.fontFace, 0)
	labelWidth *= scale
	badgeTextWidth, badgeTextHeight := text.Measure(
		badgeLabel, r.fontFace, 0,
	)
	badgeTextWidth *= scale
	badgeTextHeight *= scale
	badgeWidth := badgeTextWidth + 2*rewardBadgePadding
	left := (float64(screen.Bounds().Dx()) -
		labelWidth - rewardBadgeGap - badgeWidth) / 2

	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(left, top)
	op.ColorScale.ScaleWithColor(lineColor)
	text.Draw(screen, label, r.fontFace, op)

	badgeLeft := left + labelWidth + rewardBadgeGap
	vector.StrokeRect(
		screen,
		float32(badgeLeft), float32(top-rewardBadgePadding),
		float32(badgeWidth), float32(badgeTextHeight+2*rewardBadgePadding),
		1, newBestColor, false,
	)
	badgeOp := &text.DrawOptions{}
	badgeOp.GeoM.Scale(scale, scale)
	badgeOp.GeoM.Translate(badgeLeft+rewardBadgePadding, top)
	badgeOp.ColorScale.ScaleWithColor(newBestColor)
	text.Draw(screen, badgeLabel, r.fontFace, badgeOp)
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
