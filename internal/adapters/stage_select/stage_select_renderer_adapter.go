package stage_select

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

type StageSelectRendererAdapter struct {
	selector         *types.StageSelectorEntity
	selectorUseCases interfaces.IStageSelectorUseCases
	fontFace         text.Face
	titleFontSize    int
	regularFontSize  int
	subtitleFontSize int
	gameTitle        string
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

func (r *StageSelectRendererAdapter) DrawAll(
	screen *ebiten.Image,
	view types.StageSelectViewData,
) {
	screenBounds := screen.Bounds()
	actualWidth := float64(screenBounds.Dx())
	actualHeight := float64(screenBounds.Dy())

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

	textWidth, textHeight := text.Measure(stageText, r.fontFace, 0)

	scale := float64(r.regularFontSize) / float64(r.titleFontSize)
	if scale <= 0 {
		scale = 1
	}
	scaledWidth := textWidth * scale
	scaledHeight := textHeight * scale
	x := (actualWidth - scaledWidth) / 2
	y := (actualHeight-scaledHeight)/2 + float64(r.regularFontSize)

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
	playerY := y + scaledHeight + float64(r.regularFontSize)

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
	maxEnemiesY := playerY + scaledHeight + float64(r.regularFontSize)

	maxEnemiesColor := color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	if view.MaxEnemiesActive {
		maxEnemiesColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}

	maxEnemiesOp := &text.DrawOptions{}
	maxEnemiesOp.GeoM.Scale(maxEnemiesScale, maxEnemiesScale)
	maxEnemiesOp.GeoM.Translate(maxEnemiesX, maxEnemiesY)
	maxEnemiesOp.ColorScale.ScaleWithColor(maxEnemiesColor)
	text.Draw(screen, maxEnemiesText, r.fontFace, maxEnemiesOp)

	subtitleText := "PRESS ENTER TO START"
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
