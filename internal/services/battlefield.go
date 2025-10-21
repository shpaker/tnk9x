package services

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/models"
)

type BattlefieldService struct {
	Level models.Level
}

func NewBattlefieldService(
	level models.Level,
) BattlefieldService {
	return BattlefieldService{
		Level: level,
	}
}

func (s *BattlefieldService) Draw(screen *ebiten.Image) {
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
			constants.BattleFieldOffset+block.WorldPosition.X*constants.TileMinSize,
			constants.BattleFieldOffset+block.WorldPosition.Y*constants.TileMinSize,
		)
		screen.DrawImage(block.Image, op)
	}

}

// GetBlocks возвращает список блоков-стен для коллизий
func (s *BattlefieldService) GetBlocks() *models.Level {
	// var colliders []*models.Block
	// for i := range s.Level {
	// 	block := &s.Level[i]
	// 	if block.Data == nil {
	// 		continue
	// 	}
	// 	// Считаем стенами только кирпич и сталь
	// 	if block.Data.Name == types.Brick || block.Data.Name == types.Steel {
	// 		colliders = append(colliders, block)
	// 	}
	// }
	return &s.Level
}
