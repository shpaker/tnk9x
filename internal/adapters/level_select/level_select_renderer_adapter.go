package level_select

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/shpaker/tnk9x/internal/adapters/ui"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// Раскладка экрана выбора уровня в логических координатах 256x224
const (
	headerY       = 8
	packStarsY    = 22
	cellsTop      = 36
	cellWidth     = 40
	cellHeight    = 30
	cellGap       = 6
	previewTop    = 76
	previewLeft   = 16
	previewScale  = 0.5 // миникарта: тайл 8px -> 4px
	infoLeft      = 132
	hintBottomGap = 22
	startBottom   = 10
	// backPadding — поля рамки тап-кнопки вокруг подписи
	backPadding = 3
	// packArrowHitWidth — ширина хита стрелок пачек от края экрана
	packArrowHitWidth = 32
	// Ряд звёзд рекорда: центр от начала подписи и наименьшее
	// расстояние от конца подписи до центра ряда
	bestStarsCenter = 56
	bestStarsGap    = 24
)

// Цвета экрана выбора уровня
var (
	textColor       = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	dimTextColor    = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	hintColor       = color.NRGBA{R: 110, G: 110, B: 110, A: 255}
	lockedColor     = color.NRGBA{R: 200, G: 70, B: 50, A: 255}
	cellColor       = color.NRGBA{R: 40, G: 40, B: 40, A: 255}
	cellActiveColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	previewBack     = color.NRGBA{R: 16, G: 16, B: 16, A: 255}
	hqColor         = color.NRGBA{R: 252, G: 196, B: 36, A: 255}
	spawnerColor    = color.NRGBA{R: 200, G: 70, B: 50, A: 255}
)

// blockColors — цвета блоков миникарты
var blockColors = map[types.BlockType]color.NRGBA{
	types.Brick:  {R: 164, G: 72, B: 16, A: 255},
	types.Steel:  {R: 188, G: 188, B: 188, A: 255},
	types.Water:  {R: 60, G: 100, B: 252, A: 255},
	types.Forest: {R: 40, G: 140, B: 40, A: 255},
	types.Ice:    {R: 200, G: 220, B: 252, A: 255},
}

// enemyLetters — подписи типов врагов в составе уровня
var enemyLetters = [types.EnemyLevelsCount]string{"B", "F", "P", "A"}

// LevelSelectRendererAdapter рисует экран выбора уровня
// и отвечает на хит-тесты тапов по последней отрисовке
type LevelSelectRendererAdapter struct {
	texts           interfaces.ITextsAdapter
	fontFace        text.Face
	titleFontSize   int
	regularFontSize int
	// Точки штаба и спаунеров в пикселях поля для миникарты
	hqPosition      types.Position
	spawnerPosition []types.Position
	cellSizePx      float64
	// backRect — кнопка выхода в главное меню последней отрисовки;
	// пустая без тача и мыши
	backRect image.Rectangle
	// cellHits — ячейки уровней последней отрисовки по позициям
	cellHits ui.HitAreas
	// Стрелки пачек последней отрисовки; пустые у крайних пачек
	previousPackRect image.Rectangle
	nextPackRect     image.Rectangle
}

// LevelSelectRendererDependencies — зависимости рендера экрана выбора
// уровня; собирается composition root'ом, все поля обязательны
type LevelSelectRendererDependencies struct {
	Texts           interfaces.ITextsAdapter
	FontFace        text.Face
	TitleFontSize   int
	RegularFontSize int
	// HQPosition и EnemySpawners — в клетках танка
	HQPosition    types.Position
	EnemySpawners []types.Position
	// CellSizePx — размер клетки танка в пикселях поля
	CellSizePx int
}

func NewLevelSelectRendererAdapter(
	deps LevelSelectRendererDependencies,
) *LevelSelectRendererAdapter {
	return &LevelSelectRendererAdapter{
		texts:           deps.Texts,
		fontFace:        deps.FontFace,
		titleFontSize:   deps.TitleFontSize,
		regularFontSize: deps.RegularFontSize,
		hqPosition:      deps.HQPosition,
		spawnerPosition: deps.EnemySpawners,
		cellSizePx:      float64(deps.CellSizePx),
	}
}

// Draw рисует экран целиком
func (r *LevelSelectRendererAdapter) Draw(
	screen *ebiten.Image,
	view types.LevelSelectViewData,
) {
	width := float64(screen.Bounds().Dx())
	height := float64(screen.Bounds().Dy())

	screen.Fill(color.Black)

	r.drawHeader(screen, view, width)
	r.drawCells(screen, view, width)
	r.drawPreview(screen, view)
	r.drawInfo(screen, view)
	r.drawFooter(screen, view, width, height)
}

func (r *LevelSelectRendererAdapter) drawHeader(
	screen *ebiten.Image,
	view types.LevelSelectViewData,
	width float64,
) {
	packName := r.texts.GetOr(
		types.PackNameTextKey(view.PackIndex+1), view.PackName,
	)
	header := fmt.Sprintf(
		"%d/%d %s", view.PackIndex+1, view.PacksCount, packName,
	)
	r.drawCentered(screen, header, headerY, width, textColor)

	r.previousPackRect = image.Rectangle{}
	r.nextPackRect = image.Rectangle{}
	arrowTop := headerY - backPadding
	arrowBottom := headerY + r.regularFontSize + backPadding
	if view.PackIndex > 0 {
		r.drawText(screen, "<", 8, headerY, textColor)
		r.previousPackRect = image.Rect(
			0, arrowTop, packArrowHitWidth, arrowBottom,
		)
	}
	if view.PackIndex < view.PacksCount-1 {
		r.drawText(screen, ">", width-16, headerY, textColor)
		r.nextPackRect = image.Rect(
			int(width)-packArrowHitWidth, arrowTop, int(width), arrowBottom,
		)
	}

	// Счётчики звёзд: пачка слева, кампания справа
	r.drawStarCounter(
		screen, 16, packStarsY, view.Pack.Stars, view.Pack.MaxStars,
	)
	total := fmt.Sprintf("%d/%d", view.Pack.TotalStars, view.CampaignMaxStars)
	totalWidth := r.textWidth(total)
	r.drawStarCounterAt(
		screen, width-16-totalWidth-12, packStarsY, total,
	)
}

func (r *LevelSelectRendererAdapter) drawStarCounter(
	screen *ebiten.Image,
	x, y float64,
	value, maxValue uint,
) {
	r.drawStarCounterAt(screen, x, y, fmt.Sprintf("%d/%d", value, maxValue))
}

func (r *LevelSelectRendererAdapter) drawStarCounterAt(
	screen *ebiten.Image,
	x, y float64,
	label string,
) {
	ui.DrawStar(
		screen, float32(x+4), float32(y+4), 4.5, ui.StarFilledColor,
	)
	r.drawText(screen, label, x+12, y, textColor)
}

// cellsLeft — левый край строки ячеек уровней по центру экрана
func cellsLeft(width float64, count int) float64 {
	rowWidth := float64(count)*cellWidth + float64(count-1)*cellGap
	return (width - rowWidth) / 2
}

func (r *LevelSelectRendererAdapter) drawCells(
	screen *ebiten.Image,
	view types.LevelSelectViewData,
	width float64,
) {
	left := cellsLeft(width, len(view.Entries))
	r.cellHits.Reset()
	for position, entry := range view.Entries {
		x := left + float64(position)*(cellWidth+cellGap)
		// Хит ячейки захватывает половину зазора до соседних
		r.cellHits.Add(image.Rect(
			int(x-cellGap/2), cellsTop,
			int(x+cellWidth+cellGap/2), cellsTop+cellHeight,
		))
		border := cellColor
		if position == view.ActivePosition {
			border = cellActiveColor
		}
		vector.StrokeRect(
			screen,
			float32(x), float32(cellsTop),
			cellWidth, cellHeight,
			1, border, false,
		)

		label := fmt.Sprintf("%02d", entry.Number)
		labelColor := textColor
		if !entry.Unlocked {
			labelColor = dimTextColor
		}
		labelX := x + (cellWidth-r.textWidth(label))/2
		r.drawText(screen, label, labelX, cellsTop+5, labelColor)

		if !entry.Unlocked {
			drawLock(screen, float32(x+cellWidth/2), float32(cellsTop+22))
			continue
		}
		ui.DrawStarsRow(
			screen,
			float32(x+cellWidth/2), float32(cellsTop+22),
			3.5, 10,
			entry.Stars, types.MaxLevelStars,
		)
	}
}

// drawLock рисует замок закрытого уровня
func drawLock(screen *ebiten.Image, cx, cy float32) {
	vector.StrokeRect(screen, cx-2.5, cy-6, 5, 5, 1, dimTextColor, false)
	vector.FillRect(screen, cx-4, cy-2, 8, 6, dimTextColor, false)
}

func (r *LevelSelectRendererAdapter) drawPreview(
	screen *ebiten.Image,
	view types.LevelSelectViewData,
) {
	if view.Level == nil {
		return
	}
	mapEntity := view.Level.GetMap()
	size := mapEntity.GetSizePx()
	previewWidth := float32(float64(size.Width) * previewScale)
	previewHeight := float32(float64(size.Height) * previewScale)

	// Рамка вокруг поля — тонкая, как у ячеек уровней
	vector.StrokeRect(
		screen,
		previewLeft-1.5, previewTop-1.5,
		previewWidth+3, previewHeight+3,
		1, cellColor, false,
	)
	vector.FillRect(
		screen,
		previewLeft, previewTop,
		previewWidth, previewHeight,
		previewBack, false,
	)

	for _, block := range mapEntity.GetBlocks() {
		if block == nil || block.Data == nil {
			continue
		}
		blockColor, ok := blockColors[block.Data.Name]
		if !ok {
			continue
		}
		position := block.GetPosition()
		blockSize := block.GetSize()
		vector.FillRect(
			screen,
			float32(previewLeft+position.X*previewScale),
			float32(previewTop+position.Y*previewScale),
			float32(float64(blockSize.Width)*previewScale),
			float32(float64(blockSize.Height)*previewScale),
			blockColor, false,
		)
	}

	marker := float32(r.cellSizePx * previewScale)
	for _, spawner := range r.spawnerPosition {
		r.drawMarker(screen, spawner, marker, spawnerColor)
	}
	r.drawMarker(screen, r.hqPosition, marker, hqColor)

	if !view.LevelUnlocked {
		vector.FillRect(
			screen,
			previewLeft, previewTop,
			previewWidth, previewHeight,
			color.NRGBA{A: 170}, false,
		)
	}
}

// drawMarker рисует рамку клетки танка на миникарте
func (r *LevelSelectRendererAdapter) drawMarker(
	screen *ebiten.Image,
	cell types.Position,
	size float32,
	markerColor color.NRGBA,
) {
	vector.StrokeRect(
		screen,
		float32(previewLeft+cell.X*r.cellSizePx*previewScale)+0.5,
		float32(previewTop+cell.Y*r.cellSizePx*previewScale)+0.5,
		size-1, size-1,
		1, markerColor, false,
	)
}

func (r *LevelSelectRendererAdapter) drawInfo(
	screen *ebiten.Image,
	view types.LevelSelectViewData,
) {
	lineHeight := float64(r.regularFontSize) + 4
	y := float64(previewTop)

	if !view.Pack.Unlocked {
		r.drawText(
			screen, r.texts.Get(types.TextLevelSelectLocked),
			infoLeft, y, lockedColor,
		)
		y += lineHeight * 1.5
		if view.Pack.TotalStars < view.Pack.RequiredStars {
			r.drawText(
				screen, r.texts.Get(types.TextLevelSelectCollect),
				infoLeft, y, dimTextColor,
			)
			y += lineHeight
			r.drawStarCounterAt(
				screen, infoLeft, y,
				fmt.Sprintf(
					"%d/%d", view.Pack.TotalStars, view.Pack.RequiredStars,
				),
			)
			y += lineHeight * 1.5
		}
		if view.Pack.UnlockAfter > 0 {
			r.drawText(
				screen,
				r.texts.Format(types.TextLevelSelectWinStage, types.TextArgs{
					"Stage": fmt.Sprintf("%02d", view.Pack.UnlockAfter),
				}),
				infoLeft, y, dimTextColor,
			)
		}
		return
	}

	if view.Level == nil {
		return
	}

	r.drawText(screen, r.levelName(view.Level), infoLeft, y, textColor)
	y += lineHeight * 1.5

	r.drawText(
		screen, r.texts.Get(types.TextLevelSelectEnemies),
		infoLeft, y, dimTextColor,
	)
	y += lineHeight
	for index, count := range view.EnemyCounts {
		r.drawText(
			screen,
			fmt.Sprintf("%s %2d", enemyLetters[index], count),
			infoLeft+float64(index%2)*56,
			y+float64(index/2)*lineHeight,
			textColor,
		)
	}
	y += lineHeight * 2.5

	seconds := view.Time3StarTicks / 60
	r.drawText(screen, "3", infoLeft, y, dimTextColor)
	ui.DrawStar(
		screen, float32(infoLeft+13), float32(y+4), 4, ui.StarFilledColor,
	)
	r.drawText(
		screen,
		fmt.Sprintf("%d:%02d", seconds/60, seconds%60),
		infoLeft+24, y, dimTextColor,
	)
	y += lineHeight * 1.5

	if !view.LevelUnlocked {
		r.drawText(
			screen, r.texts.Get(types.TextLevelSelectWinPrevious),
			infoLeft, y, lockedColor,
		)
		return
	}
	best := r.texts.Get(types.TextLevelSelectBest)
	r.drawText(screen, best, infoLeft, y, dimTextColor)
	// Звёзды рекорда — не ближе зазора к подписи любой длины
	starsCenter := max(bestStarsCenter, r.textWidth(best)+bestStarsGap)
	ui.DrawStarsRow(
		screen,
		float32(infoLeft+starsCenter), float32(y+4),
		4.5, 12,
		view.LevelStars, types.MaxLevelStars,
	)
}

// levelName — название уровня на языке интерфейса: перевод
// из локали, иначе название из файла карты, а у карты без
// названия — номер
func (r *LevelSelectRendererAdapter) levelName(
	level *types.LevelEntity,
) string {
	name := r.texts.GetOr(
		types.LevelNameTextKey(level.GetNumber()), level.GetName(),
	)
	if name != "" {
		return name
	}
	return r.texts.Format(types.TextLevelSelectStageFallback, types.TextArgs{
		"Stage": fmt.Sprintf("%02d", level.GetNumber()),
	})
}

// HitBack — тап попал в кнопку выхода последней отрисовки
func (r *LevelSelectRendererAdapter) HitBack(position types.Position) bool {
	return image.Pt(int(position.X), int(position.Y)).In(r.backRect)
}

// HitLevel — позиция ячейки уровня последней отрисовки под точкой
func (r *LevelSelectRendererAdapter) HitLevel(
	position types.Position,
) (int, bool) {
	return r.cellHits.Hit(position)
}

// HitPack — стрелка пачки последней отрисовки под точкой:
// -1 — предыдущая, +1 — следующая
func (r *LevelSelectRendererAdapter) HitPack(
	position types.Position,
) (int, bool) {
	point := image.Pt(int(position.X), int(position.Y))
	switch {
	case point.In(r.previousPackRect):
		return -1, true
	case point.In(r.nextPackRect):
		return 1, true
	}
	return 0, false
}

func (r *LevelSelectRendererAdapter) drawFooter(
	screen *ebiten.Image,
	view types.LevelSelectViewData,
	width, height float64,
) {
	r.backRect = image.Rectangle{}
	button := r.texts.Get(types.TextLevelSelectButtonEnter)
	if view.TouchActive {
		button = r.texts.Get(types.TextLevelSelectButtonFire)
	}
	if view.TouchActive || view.PointerActive {
		r.drawBackButton(screen, width, height-hintBottomGap)
	} else {
		r.drawCentered(
			screen, r.texts.Get(types.TextLevelSelectKeyboardHint),
			height-hintBottomGap, width, hintColor,
		)
	}
	// Строка запуска показывает режим: один или двое игроков
	startKey := types.TextLevelSelectStartOne
	if view.PlayerCount > 1 {
		startKey = types.TextLevelSelectStartTwo
	}
	start := r.texts.Format(startKey, types.TextArgs{"Button": button})

	startColor := textColor
	if !view.LevelUnlocked {
		startColor = dimTextColor
	}
	r.drawCentered(screen, start, height-startBottom, width, startColor)
}

// drawBackButton — тап-кнопка выхода в главное меню в рамке, как
// у ячеек уровней; запоминает прямоугольник для HitBack
func (r *LevelSelectRendererAdapter) drawBackButton(
	screen *ebiten.Image,
	width, top float64,
) {
	backLabel := r.texts.Get(types.TextLevelSelectMainMenu)
	labelWidth := r.textWidth(backLabel)
	left := (width - labelWidth) / 2
	r.backRect = image.Rect(
		int(left-backPadding), int(top-backPadding),
		int(left+labelWidth+backPadding),
		int(top+float64(r.regularFontSize)+backPadding),
	)
	vector.StrokeRect(
		screen,
		float32(r.backRect.Min.X)+0.5, float32(r.backRect.Min.Y)+0.5,
		float32(r.backRect.Dx()-1), float32(r.backRect.Dy()-1),
		1, dimTextColor, false,
	)
	r.drawText(screen, backLabel, left, top, textColor)
}

// textScale — масштаб основного шрифта относительно шрифта заголовка
func (r *LevelSelectRendererAdapter) textScale() float64 {
	scale := float64(r.regularFontSize) / float64(r.titleFontSize)
	if scale <= 0 {
		return 1
	}
	return scale
}

func (r *LevelSelectRendererAdapter) textWidth(label string) float64 {
	width, _ := text.Measure(label, r.fontFace, 0)
	return width * r.textScale()
}

func (r *LevelSelectRendererAdapter) drawText(
	screen *ebiten.Image,
	label string,
	x, y float64,
	textColor color.NRGBA,
) {
	scale := r.textScale()
	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, label, r.fontFace, op)
}

func (r *LevelSelectRendererAdapter) drawCentered(
	screen *ebiten.Image,
	label string,
	y, width float64,
	textColor color.NRGBA,
) {
	r.drawText(screen, label, (width-r.textWidth(label))/2, y, textColor)
}
