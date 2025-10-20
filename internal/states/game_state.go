package states

import (
	"errors"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/services"
)

type GameState struct {
	battleFieldService services.BattleFieldService
	playerService      *services.PlayerService
}

// NewGameState создает новое состояние игры с переданным репозиторием спрайтов
func NewGameState(
	levelService interfaces.ILevelsDataService,
	playerOneService *services.PlayerService,
) (GameState, error) {
	level, err := levelService.GetLevel(1)
	if err != nil {
		return GameState{}, err
	}

	if playerOneService == nil {
		return GameState{}, errors.New("playerService is nil")
	}

	return GameState{
		battleFieldService: services.NewBattleFieldService(level),
		playerService:      playerOneService,
	}, nil
}

func (state GameState) Update() (interfaces.State, error) {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return nil, errors.New("exit application")
	}
	state.playerService.KeyPressed()
	state.playerService.Update(constants.DT, state.battleFieldService.GetWallsColliders())
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
	state.battleFieldService.Draw(screen)
	state.playerService.Draw(screen)
}
