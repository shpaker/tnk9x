package states

import (
	"errors"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/services"
)

type GameState struct {
	battleFieldService         services.MapService
	playerOneService           *services.PlayerService
	playerOneControllerService *services.ControllerService
	bulletsService             *services.BulletsService
	collidersService           *services.CollidersService
	rendererService            *services.RendererService
}

// NewGameState создает новое состояние игры с переданным репозиторием карт
func NewGameState(
	mapsRepo processed.IMapsDataRepository,
	playerOneService *services.PlayerService,
	playerOneControllerService *services.ControllerService,
	bulletsService *services.BulletsService,
) (GameState, error) {
	level, err := mapsRepo.GetLevel(1)
	if err != nil {
		return GameState{}, err
	}

	if playerOneService == nil {
		return GameState{}, errors.New("playerService is nil")
	}

	// Создаем репозиторий блоков и заполняем его данными уровня
	blocksRepo := game.NewBlocksRepository()
	for _, block := range level {
		blocksRepo.AddBlock(block)
	}

	battleFieldService := services.NewMapService(blocksRepo)
	collidersService := services.NewCollidersService(
		bulletsService,
		playerOneService,
		battleFieldService,
	)
	rendererService := services.NewRendererService(
		battleFieldService,
		playerOneService,
		bulletsService,
		collidersService,
	)

	return GameState{
		battleFieldService:         battleFieldService,
		playerOneService:           playerOneService,
		playerOneControllerService: playerOneControllerService,
		bulletsService:             bulletsService,
		collidersService:           collidersService,
		rendererService:            rendererService,
	}, nil
}

func (state GameState) Update() (interfaces.State, error) {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return nil, errors.New("exit application")
	}

	state.playerOneControllerService.Update()
	state.playerOneService.Update(constants.DT)
	state.bulletsService.Update(constants.DT)
	state.collidersService.Update()
	return nil, nil
}

func (state GameState) Draw(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		0,
		0,
		float32(screen.Bounds().Dx()),
		float32(screen.Bounds().Dy()),
		color.Gray{Y: 128},
		false,
	)
	state.rendererService.DrawAll(screen)
}
