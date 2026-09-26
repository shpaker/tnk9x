package stage_select

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// MenuHit — зона меню, в которую попал тап; Prev/Next — левая и
// правая половины экрана в полосе соответствующей строки
type MenuHit int

const (
	MenuHitNone MenuHit = iota
	MenuHitLevelPrev
	MenuHitLevelNext
	MenuHitPlayersPrev
	MenuHitPlayersNext
	MenuHitMaxEnemiesPrev
	MenuHitMaxEnemiesNext
	MenuHitGraphicsPrev
	MenuHitGraphicsNext
	MenuHitQuit
	MenuHitStart
)

type StageSelectRendererAdapter struct {
	selector         *types.StageSelectorEntity
	selectorUseCases interfaces.IStageSelectorUseCases
	fontFace         text.Face
	titleFontSize    int
	regularFontSize  int
	subtitleFontSize int
	gameTitle        string

	// Размер экрана последней отрисовки — для хит-тестов тапов
	lastWidth  float64
	lastHeight float64
	// Видимость строки QUIT в последней отрисовке: без неё
	// тап-зоны QUIT не существует
	lastQuitVisible bool
}

// StageSelectRendererDependencies — готовый граф зависимостей рендера
// меню; собирается composition root'ом, все поля обязательны
type StageSelectRendererDependencies struct {
	Selector         *types.StageSelectorEntity
	SelectorUseCases interfaces.IStageSelectorUseCases
	FontFace         text.Face
	TitleFontSize    int
	RegularFontSize  int
	SubtitleFontSize int
	GameTitle        string
}

func NewStageSelectRendererAdapter(
	deps StageSelectRendererDependencies,
) *StageSelectRendererAdapter {
	return &StageSelectRendererAdapter{
		selector:         deps.Selector,
		selectorUseCases: deps.SelectorUseCases,
		fontFace:         deps.FontFace,
		titleFontSize:    deps.TitleFontSize,
		regularFontSize:  deps.RegularFontSize,
		subtitleFontSize: deps.SubtitleFontSize,
		gameTitle:        deps.GameTitle,
	}
}

// menuRow — строка меню в порядке отрисовки
type menuRow int

const (
	menuRowLevel menuRow = iota
	menuRowPlayers
	menuRowMaxEnemies
	menuRowGraphics
	menuRowQuit
)

// hintLinesCount — строки подсказки управления над строкой запуска
const hintLinesCount = 2

// stageSelectMenuLayout — вертикальная раскладка меню и границы
// полос тап-зон в логических координатах экрана
type stageSelectMenuLayout struct {
	rowHeight float64 // высота строки меню после масштабирования
	rows      []menuRow
	rowTops   []float64

	menuTop    float64
	menuBottom float64
	hintTops   [hintLinesCount]float64
	startTop   float64 // нижняя полоса запуска игры вместе с подсказкой
	subtitleY  float64
}

// menuRows — строки меню; QUIT есть только там, где приложение
// может завершиться
func menuRows(quitVisible bool) []menuRow {
	rows := []menuRow{
		menuRowLevel,
		menuRowPlayers,
		menuRowMaxEnemies,
		menuRowGraphics,
	}
	if quitVisible {
		rows = append(rows, menuRowQuit)
	}
	return rows
}

// menuLayout — единый источник вертикальных позиций меню для
// отрисовки и хит-тестов: строка запуска у нижнего края, над ней
// подсказка, строки меню — по центру между заголовком и подсказкой
func (r *StageSelectRendererAdapter) menuLayout(
	width, height float64,
	quitVisible bool,
) stageSelectMenuLayout {
	stageText := r.selectorUseCases.String(r.selector)
	_, textHeight := text.Measure(stageText, r.fontFace, 0)
	scale := float64(r.regularFontSize) / float64(r.titleFontSize)
	if scale <= 0 {
		scale = 1
	}
	rowHeight := textHeight * scale
	gap := float64(r.regularFontSize)

	subtitleY := height - float64(r.subtitleFontSize)
	hintPitch := rowHeight + gap/2
	var hintTops [hintLinesCount]float64
	for i := range hintTops {
		hintTops[i] = subtitleY - gap - float64(hintLinesCount-i)*hintPitch
	}

	rows := menuRows(quitVisible)
	blockHeight := float64(len(rows))*(rowHeight+gap) - gap
	titleBottom := height/4 + float64(r.titleFontSize)/2
	areaBottom := hintTops[0] - gap
	firstTop := titleBottom + (areaBottom-titleBottom-blockHeight)/2

	rowTops := make([]float64, len(rows))
	for i := range rowTops {
		rowTops[i] = firstTop + float64(i)*(rowHeight+gap)
	}

	return stageSelectMenuLayout{
		rowHeight:  rowHeight,
		rows:       rows,
		rowTops:    rowTops,
		menuTop:    firstTop - gap,
		menuBottom: rowTops[len(rowTops)-1] + rowHeight + gap,
		hintTops:   hintTops,
		startTop:   hintTops[0] - gap/2,
		subtitleY:  subtitleY,
	}
}

// rowAt — строка меню под точкой: полосы строк делятся посередине
// промежутков
func (layout stageSelectMenuLayout) rowAt(y float64) (menuRow, bool) {
	if y < layout.menuTop || y >= layout.menuBottom {
		return 0, false
	}
	for i := len(layout.rowTops) - 1; i > 0; i-- {
		boundary := (layout.rowTops[i-1] + layout.rowHeight +
			layout.rowTops[i]) / 2
		if y >= boundary {
			return layout.rows[i], true
		}
	}
	return layout.rows[0], true
}

// HitTest определяет зону меню по тапу в логических координатах
// экрана; до первой отрисовки зоны неизвестны
func (r *StageSelectRendererAdapter) HitTest(pos types.Position) MenuHit {
	if r.lastWidth <= 0 || r.lastHeight <= 0 {
		return MenuHitNone
	}
	layout := r.menuLayout(r.lastWidth, r.lastHeight, r.lastQuitVisible)
	if pos.Y >= layout.startTop {
		return MenuHitStart
	}
	row, ok := layout.rowAt(pos.Y)
	if !ok {
		return MenuHitNone
	}

	next := pos.X >= r.lastWidth/2
	switch row {
	case menuRowLevel:
		return pickHit(MenuHitLevelPrev, MenuHitLevelNext, next)
	case menuRowPlayers:
		return pickHit(MenuHitPlayersPrev, MenuHitPlayersNext, next)
	case menuRowMaxEnemies:
		return pickHit(
			MenuHitMaxEnemiesPrev, MenuHitMaxEnemiesNext, next,
		)
	case menuRowGraphics:
		return pickHit(MenuHitGraphicsPrev, MenuHitGraphicsNext, next)
	default:
		return MenuHitQuit
	}
}

func pickHit(prev, next MenuHit, isNext bool) MenuHit {
	if isNext {
		return next
	}

	return prev
}

// Цвета строк меню
var (
	menuRowColor       = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	menuRowActiveColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	subtitleColor      = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	// hintColor — тусклее строк меню и ниже порога свечения bloom
	hintColor = color.NRGBA{R: 110, G: 110, B: 110, A: 255}
)

// Подсказки управления: клавиатурная и сенсорная; вторая строка
// напоминает, что графику можно переключить на классическую
var (
	keyboardHintLines = [hintLinesCount]string{
		"P1 WASD+SPACE  P2 ARROWS+ENTER",
		"ESC/P PAUSE  F2 RTX/CLASSIC",
	}
	touchHintLines = [hintLinesCount]string{
		"D-PAD MOVE  BUTTON FIRE",
		"TAP GRAPHICS FOR RTX/CLASSIC",
	}
)

func (r *StageSelectRendererAdapter) DrawAll(
	screen *ebiten.Image,
	view types.StageSelectViewData,
) {
	screenBounds := screen.Bounds()
	actualWidth := float64(screenBounds.Dx())
	actualHeight := float64(screenBounds.Dy())
	r.lastWidth = actualWidth
	r.lastHeight = actualHeight
	r.lastQuitVisible = view.QuitVisible
	layout := r.menuLayout(actualWidth, actualHeight, view.QuitVisible)

	screen.Fill(color.Black)

	titleText := r.gameTitle
	titleWidth, _ := text.Measure(titleText, r.fontFace, 0)
	titleX := (actualWidth - titleWidth) / 2
	titleY := actualHeight/4 - float64(r.titleFontSize)/2

	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(titleX, titleY)
	titleOp.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, titleText, r.fontFace, titleOp)

	regularScale := float64(r.regularFontSize) / float64(r.titleFontSize)
	for i, row := range layout.rows {
		label, active := r.rowView(row, view)
		rowColor := menuRowColor
		if active {
			rowColor = menuRowActiveColor
		}
		r.drawCentered(
			screen, label, regularScale, layout.rowTops[i], rowColor,
		)
	}

	hintLines := keyboardHintLines
	if view.TouchActive {
		hintLines = touchHintLines
	}
	for i, line := range hintLines {
		r.drawCentered(
			screen, line, regularScale, layout.hintTops[i], hintColor,
		)
	}

	subtitleText := "PRESS ENTER TO START"
	if view.TouchActive {
		subtitleText = "TAP HERE TO START"
	}
	r.drawCentered(
		screen,
		subtitleText,
		float64(r.subtitleFontSize)/float64(r.titleFontSize),
		layout.subtitleY,
		subtitleColor,
	)
}

// rowView — подпись строки меню и её выделение
func (r *StageSelectRendererAdapter) rowView(
	row menuRow,
	view types.StageSelectViewData,
) (string, bool) {
	switch row {
	case menuRowLevel:
		return r.selectorUseCases.String(r.selector), view.LevelActive
	case menuRowPlayers:
		return fmt.Sprintf("PLAYERS %d", view.PlayerCount),
			view.PlayersActive
	case menuRowMaxEnemies:
		return fmt.Sprintf("MAX ENEMIES %d", view.MaxActiveEnemies),
			view.MaxEnemiesActive
	case menuRowGraphics:
		return types.GraphicsLabel(view.EffectsEnabled),
			view.GraphicsActive
	default:
		return "QUIT", view.QuitActive
	}
}

// drawCentered рисует строку по центру экрана в заданном масштабе
// шрифта заголовка
func (r *StageSelectRendererAdapter) drawCentered(
	screen *ebiten.Image,
	label string,
	scale float64,
	top float64,
	textColor color.NRGBA,
) {
	if scale <= 0 {
		scale = 1
	}
	width, _ := text.Measure(label, r.fontFace, 0)

	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(
		(float64(screen.Bounds().Dx())-width*scale)/2,
		top,
	)
	op.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, label, r.fontFace, op)
}
