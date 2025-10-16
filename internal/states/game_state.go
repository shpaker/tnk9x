package states

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/services"
)

type GameState struct {
	BattleFieldService services.BattleFieldService
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
		BattleFieldService: services.NewBattleFieldService(level),
	}, nil
}

func (state GameState) Update() (interfaces.State, error) {
	return nil, nil
}

func (state GameState) Draw(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		float32(0),
		float32(0),
		float32(screen.Bounds().Dx()),
		float32(screen.Bounds().Dy()),
		color.Gray{Y: 128},
		false,
	)
	state.BattleFieldService.Draw(screen)
}
