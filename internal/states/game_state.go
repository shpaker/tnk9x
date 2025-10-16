package states

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/services"
)

type GameState struct {
	worldService services.WorldService
}

// NewGameState создает новое состояние игры с переданным репозиторием спрайтов
func NewGameState(
	levelService interfaces.ILevelsService,
) (GameState, error) {
	level, err := levelService.GetLevel(1)
	if err != nil {
		return GameState{}, err
	}
	return GameState{
		worldService: services.NewWorldService(level),
	}, nil
}

func (state GameState) Update() (interfaces.State, error) {
	return nil, nil
}

func (state GameState) Draw(screen *ebiten.Image) {
	level := state.worldService.Level
	for _, block := range level {
		if block.Image == nil {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		// Предполагаем, что блоки имеют координаты X, Y в WorldPosition
		op.GeoM.Translate(float64(block.WorldPosition.X*8), float64(block.WorldPosition.Y*8))
		screen.DrawImage(block.Image, op)
	}
}
