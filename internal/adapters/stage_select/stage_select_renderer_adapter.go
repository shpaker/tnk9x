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

// stageSelectMenuLayout — вертикальная раскладка меню и границы
// полос тап-зон в логических координатах экрана
type stageSelectMenuLayout struct {
	rowHeight  float64 // высота строки меню после масштабирования
	stageTop   float64
	playersTop float64
	enemiesTop float64
	quitTop    float64

	menuTop        float64
	stagePlayers   float64 // граница полос LEVEL / PLAYERS
	playersEnemies float64 // граница полос PLAYERS / MAX ENEMIES
	enemiesQuit    float64 // граница полос MAX ENEMIES / QUIT
	menuBottom     float64
	startTop       float64 // нижняя полоса запуска игры
}

// menuLayout — единый источник вертикальных позиций меню для
// отрисовки и хит-тестов
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
	stageTop := (height-rowHeight)/2 + gap
	playersTop := stageTop + rowHeight + gap
	enemiesTop := playersTop + rowHeight + gap
	quitTop := enemiesTop + rowHeight + gap

	menuBottom := enemiesTop + rowHeight + gap
	if quitVisible {
		menuBottom = quitTop + rowHeight + gap
	}

	return stageSelectMenuLayout{
		rowHeight:      rowHeight,
		stageTop:       stageTop,
		playersTop:     playersTop,
		enemiesTop:     enemiesTop,
		quitTop:        quitTop,
		menuTop:        stageTop - gap,
		stagePlayers:   (stageTop + rowHeight + playersTop) / 2,
		playersEnemies: (playersTop + rowHeight + enemiesTop) / 2,
		enemiesQuit:    (enemiesTop + rowHeight + quitTop) / 2,
		menuBottom:     menuBottom,
		startTop:       height - 3*float64(r.subtitleFontSize),
	}
}

// HitTest определяет зону меню по тапу в логических координатах
// экрана; до первой отрисовки зоны неизвестны
func (r *StageSelectRendererAdapter) HitTest(pos types.Position) MenuHit {
	if r.lastWidth <= 0 || r.lastHeight <= 0 {
		return MenuHitNone
	}
	layout := r.menuLayout(r.lastWidth, r.lastHeight, r.lastQuitVisible)
	next := pos.X >= r.lastWidth/2
	switch {
	case pos.Y >= layout.startTop:
		return MenuHitStart
	case pos.Y < layout.menuTop || pos.Y >= layout.menuBottom:
		return MenuHitNone
	case pos.Y < layout.stagePlayers:
		return pickHit(MenuHitLevelPrev, MenuHitLevelNext, next)
	case pos.Y < layout.playersEnemies:
		return pickHit(MenuHitPlayersPrev, MenuHitPlayersNext, next)
	case !r.lastQuitVisible || pos.Y < layout.enemiesQuit:
		return pickHit(
			MenuHitMaxEnemiesPrev, MenuHitMaxEnemiesNext, next,
		)
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

	stageText := r.selectorUseCases.String(r.selector)

	textWidth, _ := text.Measure(stageText, r.fontFace, 0)

	scale := float64(r.regularFontSize) / float64(r.titleFontSize)
	if scale <= 0 {
		scale = 1
	}
	scaledWidth := textWidth * scale
	x := (actualWidth - scaledWidth) / 2
	y := layout.stageTop

	stageColor := color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	if view.LevelActive {
		stageColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}

	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(stageColor)
	text.Draw(screen, stageText, r.fontFace, op)

	playerText := fmt.Sprintf("PLAYERS %d", view.PlayerCount)
	playerWidth, _ := text.Measure(playerText, r.fontFace, 0)
	playerScale := scale
	playerScaledWidth := playerWidth * playerScale
	playerX := (actualWidth - playerScaledWidth) / 2
	playerY := layout.playersTop

	playerColor := color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	if view.PlayersActive {
		playerColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}

	playerOp := &text.DrawOptions{}
	playerOp.GeoM.Scale(playerScale, playerScale)
	playerOp.GeoM.Translate(playerX, playerY)
	playerOp.ColorScale.ScaleWithColor(playerColor)
	text.Draw(screen, playerText, r.fontFace, playerOp)

	maxEnemiesText := fmt.Sprintf("MAX ENEMIES %d", view.MaxActiveEnemies)
	maxEnemiesWidth, _ := text.Measure(maxEnemiesText, r.fontFace, 0)
	maxEnemiesScale := scale
	maxEnemiesScaledWidth := maxEnemiesWidth * maxEnemiesScale
	maxEnemiesX := (actualWidth - maxEnemiesScaledWidth) / 2
	maxEnemiesY := layout.enemiesTop

	maxEnemiesColor := color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	if view.MaxEnemiesActive {
		maxEnemiesColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}

	maxEnemiesOp := &text.DrawOptions{}
	maxEnemiesOp.GeoM.Scale(maxEnemiesScale, maxEnemiesScale)
	maxEnemiesOp.GeoM.Translate(maxEnemiesX, maxEnemiesY)
	maxEnemiesOp.ColorScale.ScaleWithColor(maxEnemiesColor)
	text.Draw(screen, maxEnemiesText, r.fontFace, maxEnemiesOp)

	if view.QuitVisible {
		quitText := "QUIT"
		quitWidth, _ := text.Measure(quitText, r.fontFace, 0)
		quitX := (actualWidth - quitWidth*scale) / 2
		quitY := layout.quitTop

		quitColor := color.NRGBA{R: 150, G: 150, B: 150, A: 255}
		if view.QuitActive {
			quitColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}

		quitOp := &text.DrawOptions{}
		quitOp.GeoM.Scale(scale, scale)
		quitOp.GeoM.Translate(quitX, quitY)
		quitOp.ColorScale.ScaleWithColor(quitColor)
		text.Draw(screen, quitText, r.fontFace, quitOp)
	}

	subtitleText := "PRESS ENTER TO START"
	if view.TouchActive {
		subtitleText = "TAP HERE TO START"
	}
	subtitleWidth, _ := text.Measure(subtitleText, r.fontFace, 0)
	subtitleScale := float64(r.subtitleFontSize) / float64(r.titleFontSize)
	if subtitleScale <= 0 {
		subtitleScale = 1
	}
	subtitleScaledWidth := subtitleWidth * subtitleScale
	subtitleX := (actualWidth - subtitleScaledWidth) / 2
	subtitleY := actualHeight - float64(r.subtitleFontSize)

	subtitleOp := &text.DrawOptions{}
	subtitleOp.GeoM.Scale(subtitleScale, subtitleScale)
	subtitleOp.GeoM.Translate(subtitleX, subtitleY)
	subtitleOp.ColorScale.ScaleWithColor(
		color.NRGBA{R: 200, G: 200, B: 200, A: 255},
	)
	text.Draw(screen, subtitleText, r.fontFace, subtitleOp)
}
