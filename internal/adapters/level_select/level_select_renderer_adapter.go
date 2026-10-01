package level_select

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/shpaker/tnk9x/internal/adapters/ui"
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
	fontFace        text.Face
	titleFontSize   int
	regularFontSize int
	// Точки штаба и спаунеров в пикселях поля для миникарты
	hqPosition      types.Position
	spawnerPosition []types.Position
	cellSizePx      float64
}

// LevelSelectRendererDependencies — зависимости рендера экрана выбора
// уровня; собирается composition root'ом, все поля обязательны
type LevelSelectRendererDependencies struct {
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

// menuLabels — подписи пунктов меню экрана выбора уровня
var menuLabels = map[types.LevelSelectMenuItem]string{
	types.LevelSelectMenuItemBack:     "BACK",
	types.LevelSelectMenuItemPlayers:  "PLAYERS",
	types.LevelSelectMenuItemSettings: "SETTINGS",
	types.LevelSelectMenuItemQuit:     "QUIT",
}

// DrawMenu рисует меню экрана выбора уровня оверлеем поверх экрана;
// у строки PLAYERS — число игроков
func (r *LevelSelectRendererAdapter) DrawMenu(
	screen *ebiten.Image,
	view types.LevelSelectMenuViewData,
) {
	rows := make([]ui.MenuRow, len(view.Items))
	for i, item := range view.Items {
		rows[i] = ui.MenuRow{Label: menuLabels[item]}
		if item == types.LevelSelectMenuItemPlayers {
			rows[i].Value = fmt.Sprint(view.Players)
		}
	}

	ui.DrawMenu(
		screen,
		ui.MenuFont{
			Face:            r.fontFace,
			TitleFontSize:   r.titleFontSize,
			RegularFontSize: r.regularFontSize,
		},
		"MENU",
		rows,
		view.ActiveIndex,
	)
}

func (r *LevelSelectRendererAdapter) drawHeader(
	screen *ebiten.Image,
	view types.LevelSelectViewData,
	width float64,
) {
	header := fmt.Sprintf(
		"%d/%d %s", view.PackIndex+1, view.PacksCount, view.PackName,
	)
	r.drawCentered(screen, header, headerY, width, textColor)

	if view.PackIndex > 0 {
		r.drawText(screen, "<", 8, headerY, textColor)
	}
	if view.PackIndex < view.PacksCount-1 {
		r.drawText(screen, ">", width-16, headerY, textColor)
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
	for position, entry := range view.Entries {
		x := left + float64(position)*(cellWidth+cellGap)
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
		r.drawText(screen, "LOCKED", infoLeft, y, lockedColor)
		y += lineHeight * 1.5
		if view.Pack.TotalStars < view.Pack.RequiredStars {
			r.drawText(screen, "COLLECT", infoLeft, y, dimTextColor)
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
				fmt.Sprintf("WIN STAGE %02d", view.Pack.UnlockAfter),
				infoLeft, y, dimTextColor,
			)
		}
		return
	}

	if view.Level == nil {
		return
	}

	r.drawText(screen, view.Level.GetName(), infoLeft, y, textColor)
	y += lineHeight * 1.5

	r.drawText(screen, "ENEMIES", infoLeft, y, dimTextColor)
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
		r.drawText(screen, "WIN PREVIOUS", infoLeft, y, lockedColor)
		return
	}
	r.drawText(screen, "BEST", infoLeft, y, dimTextColor)
	ui.DrawStarsRow(
		screen,
		float32(infoLeft+56), float32(y+4),
		4.5, 12,
		view.LevelStars, types.MaxLevelStars,
	)
}

// Подсказки управления: клавиатурная и сенсорная
const (
	keyboardHint = "ARROWS SELECT  ESC MENU"
	touchHint    = "D-PAD SELECT  PAUSE MENU"
)

func (r *LevelSelectRendererAdapter) drawFooter(
	screen *ebiten.Image,
	view types.LevelSelectViewData,
	width, height float64,
) {
	hint := keyboardHint
	button := "ENTER"
	if view.TouchActive {
		hint = touchHint
		button = "FIRE"
	}
	// Строка запуска показывает режим: один или двое игроков
	mode := "1 PLAYER"
	if view.PlayerCount > 1 {
		mode = fmt.Sprintf("%d PLAYERS", view.PlayerCount)
	}
	start := fmt.Sprintf("PRESS %s: %s", button, mode)
	r.drawCentered(screen, hint, height-hintBottomGap, width, hintColor)

	startColor := textColor
	if !view.LevelUnlocked {
		startColor = dimTextColor
	}
	r.drawCentered(screen, start, height-startBottom, width, startColor)
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
