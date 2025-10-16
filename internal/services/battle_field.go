package services

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/models"
)

type BattleFieldService struct {
	Level  models.Level
	Player models.Player
}

func NewBattleFieldService(level models.Level, player models.Player) BattleFieldService {
	return BattleFieldService{Level: level, Player: player}
}

func (s *BattleFieldService) Draw(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		float32(constants.BattleFieldOffset),
		float32(constants.BattleFieldOffset),
		float32(constants.BattleFieldWidthHeight),
		float32(constants.BattleFieldWidthHeight),
		color.Black,
		false,
	)

	// Draw blocks on the battle field
	for _, block := range s.Level {
		if block.Image == nil {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		// Предполагаем, что блоки имеют координаты X, Y в WorldPosition
		op.GeoM.Translate(
			float64(constants.BattleFieldOffset+block.WorldPosition.X*constants.TileMinSize),
			float64(constants.BattleFieldOffset+block.WorldPosition.Y*constants.TileMinSize),
		)
		screen.DrawImage(block.Image, op)
	}

	// Draw player on the battle field
	if s.Player.Image != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			float64(constants.BattleFieldOffset+s.Player.WorldPosition.X),
			float64(constants.BattleFieldOffset+s.Player.WorldPosition.Y),
		)
		screen.DrawImage(s.Player.Image, op)
	}
}
