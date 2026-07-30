package score

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/tnk9x/internal/types"
)

// Раскладка экрана итогов на логическом экране 256x224
const (
	scoreHiScoreY   = 24
	scoreStageY     = 40
	scorePlayersY   = 64
	scoreRowsStartY = 88
	scoreRowStepY   = 20
	scoreTotalY     = 172

	scorePlayer1ColumnX = 32.0
	scorePlayer2ColumnX = 160.0
)

var (
	scoreLabelColor  = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	scoreValueColor  = color.NRGBA{R: 255, G: 160, B: 68, A: 255}
	scoreAccentColor = color.NRGBA{R: 228, G: 68, B: 52, A: 255}
)

// ScoreRendererAdapter рисует экран подсчёта очков после этапа:
// счёт каждого игрока и убитые враги по типам с очками
type ScoreRendererAdapter struct {
	hudFontFace text.Face
}

func NewScoreRendererAdapter(hudFontFace text.Face) *ScoreRendererAdapter {
	return &ScoreRendererAdapter{hudFontFace: hudFontFace}
}

func (r *ScoreRendererAdapter) DrawAll(
	screen *ebiten.Image,
	view types.ScoreViewData,
) {
	screenBounds := screen.Bounds()
	actualWidth := float64(screenBounds.Dx())

	screen.Fill(color.Black)

	r.drawCenteredText(
		screen,
		fmt.Sprintf("HI-SCORE %7d", view.HiScore),
		actualWidth,
		scoreHiScoreY,
		scoreAccentColor,
	)
	r.drawCenteredText(
		screen,
		fmt.Sprintf("STAGE %2d", view.StageNumber),
		actualWidth,
		scoreStageY,
		scoreLabelColor,
	)

	r.drawPlayerColumn(
		screen,
		"I-PLAYER",
		view.Player1,
		scorePlayer1ColumnX,
	)
	if view.PlayerCount > 1 {
		r.drawPlayerColumn(
			screen,
			"II-PLAYER",
			view.Player2,
			scorePlayer2ColumnX,
		)
	}
}

// drawPlayerColumn рисует колонку игрока: заголовок, счёт забега,
// строки «очки x убито» по уровням врагов и итог по этапу
func (r *ScoreRendererAdapter) drawPlayerColumn(
	screen *ebiten.Image,
	label string,
	player types.ScorePlayerViewData,
	columnX float64,
) {
	r.drawText(screen, label, columnX, scorePlayersY, scoreAccentColor)
	r.drawText(
		screen,
		fmt.Sprintf("%7d", player.Score),
		columnX,
		scorePlayersY+10,
		scoreValueColor,
	)

	totalKills := uint(0)
	for tier, kills := range player.Kills {
		points := uint(100 * (tier + 1))
		row := fmt.Sprintf("%4d PTS %2d", points*kills, kills)
		rowY := float64(scoreRowsStartY + tier*scoreRowStepY)
		r.drawText(screen, row, columnX, rowY, scoreLabelColor)
		totalKills += kills
	}

	r.drawText(
		screen,
		fmt.Sprintf("TOTAL %2d", totalKills),
		columnX,
		scoreTotalY,
		scoreLabelColor,
	)
}

func (r *ScoreRendererAdapter) drawText(
	screen *ebiten.Image,
	value string,
	x, y float64,
	clr color.Color,
) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, value, r.hudFontFace, op)
}

func (r *ScoreRendererAdapter) drawCenteredText(
	screen *ebiten.Image,
	value string,
	screenWidth float64,
	y float64,
	clr color.Color,
) {
	textWidth, _ := text.Measure(value, r.hudFontFace, 0)
	r.drawText(screen, value, (screenWidth-textWidth)/2, y, clr)
}
