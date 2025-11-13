package stage_select

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/tnk25/internal/types"
	"github.com/shpaker/tnk25/internal/use_cases"
)

type StageSelectRendererAdapter struct {
	selector         *types.StageSelectorEntity
	selectorUseCases *use_cases.StageSelectorUseCases
	fontFace         text.Face
	screenWidth      int
	screenHeight     int
	titleFontSize    int
	regularFontSize  int
	subtitleFontSize int
	gameTitle        string
}

func NewStageSelectRendererAdapter(
	selector *types.StageSelectorEntity,
	selectorUseCases *use_cases.StageSelectorUseCases,
	fontFace text.Face,
	screenWidth int,
	screenHeight int,
	titleFontSize int,
	regularFontSize int,
	subtitleFontSize int,
	gameTitle string,
) (*StageSelectRendererAdapter, error) {
	if fontFace == nil {
		return nil, fmt.Errorf("font face is nil")
	}

	if titleFontSize <= 0 {
		titleFontSize = 32
	}
	if regularFontSize <= 0 {
		regularFontSize = titleFontSize
	}
	if subtitleFontSize <= 0 {
		subtitleFontSize = regularFontSize
	}

	return &StageSelectRendererAdapter{
		selector:         selector,
		selectorUseCases: selectorUseCases,
		fontFace:         fontFace,
		screenWidth:      screenWidth,
		screenHeight:     screenHeight,
		titleFontSize:    titleFontSize,
		regularFontSize:  regularFontSize,
		subtitleFontSize: subtitleFontSize,
		gameTitle:        gameTitle,
	}, nil
}

func (r *StageSelectRendererAdapter) DrawAll(
	screen *ebiten.Image,
	levelActive bool,
	playerCount int,
	isPlayersActive bool,
	maxActiveEnemies uint,
) {
	if r.selector == nil || r.selectorUseCases == nil {
		return
	}

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
	if levelActive {
		stageColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}

	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(stageColor)
	text.Draw(screen, stageText, r.fontFace, op)

	if playerCount < 1 {
		playerCount = 1
	}
	if playerCount > 2 {
		playerCount = 2
	}

	playerText := fmt.Sprintf("PLAYERS %d", playerCount)
	playerWidth, _ := text.Measure(playerText, r.fontFace, 0)
	playerScale := scale
	playerScaledWidth := playerWidth * playerScale
	playerX := (actualWidth - playerScaledWidth) / 2
	playerY := y + scaledHeight + float64(r.regularFontSize)

	playerColor := color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	if isPlayersActive {
		playerColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}

	playerOp := &text.DrawOptions{}
	playerOp.GeoM.Scale(playerScale, playerScale)
	playerOp.GeoM.Translate(playerX, playerY)
	playerOp.ColorScale.ScaleWithColor(playerColor)
	text.Draw(screen, playerText, r.fontFace, playerOp)

	if maxActiveEnemies < 3 {
		maxActiveEnemies = 3
	}
	if maxActiveEnemies > 10 {
		maxActiveEnemies = 10
	}

	maxEnemiesText := fmt.Sprintf("MAX ENEMIES %d", maxActiveEnemies)
	maxEnemiesWidth, _ := text.Measure(maxEnemiesText, r.fontFace, 0)
	maxEnemiesScale := scale
	maxEnemiesScaledWidth := maxEnemiesWidth * maxEnemiesScale
	maxEnemiesX := (actualWidth - maxEnemiesScaledWidth) / 2
	maxEnemiesY := playerY + scaledHeight + float64(r.regularFontSize)

	maxEnemiesColor := color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	if !levelActive && !isPlayersActive {
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
