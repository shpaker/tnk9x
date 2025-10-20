package services

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/models"
	"github.com/shpaker/gonflict/internal/types"
)

type BattleFieldService struct {
	Level models.Level
}

func NewBattleFieldService(
	level models.Level,
) BattleFieldService {
	return BattleFieldService{
		Level: level,
	}
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
			constants.BattleFieldOffset+block.WorldPosition.X*constants.TileMinSize,
			constants.BattleFieldOffset+block.WorldPosition.Y*constants.TileMinSize,
		)
		screen.DrawImage(block.Image, op)
	}

}

// GetWallsColliders возвращает список блоков-стен для коллизий
func (s *BattleFieldService) GetWallsColliders() []*models.Block {
	var colliders []*models.Block
	for i := range s.Level {
		block := &s.Level[i]
		if block.Data == nil {
			continue
		}
		// Считаем стенами только кирпич и сталь
		if block.Data.Name == types.Brick || block.Data.Name == types.Steel {
			colliders = append(colliders, block)
		}
	}
	return colliders
}
