package services

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/models"
)

type MapService struct {
	Level models.Level
}

func NewMapService(
	level models.Level,
) MapService {
	return MapService{
		Level: level,
	}
}

func (s *MapService) Draw(screen *ebiten.Image) {
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
func (s *MapService) GetBlocks() models.Level {
	return s.Level
}

func (s *MapService) RemoveBlock(block *models.Block) {
	if block == nil {
		return
	}

	// Ищем блок в слайсе по указателю на элемент в слайсе
	for i := range s.Level {
		if &s.Level[i] == block {
			// Удаляем блок из слайса
			s.Level = append(s.Level[:i], s.Level[i+1:]...)
			return
		}
	}
}
