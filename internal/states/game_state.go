package states

import (
	"errors"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shpaker/gonflict/internal/adapters"
	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/use_cases"
)

type GameState struct {
	playerUseCases    *use_cases.PlayerUseCases
	bulletUseCases    *use_cases.BulletUseCases
	mapUseCases       *use_cases.MapUseCases
	collisionUseCases *use_cases.CollisionUseCases
	inputAdapter      *adapters.InputAdapter
	rendererAdapter   *adapters.RendererAdapter
}

// NewGameState создает новое состояние игры с переданным репозиторием карт
func NewGameState(
	mapsRepo processed.IMapsDataRepository,
	mapTilesetRepo processed.ITilesetRepository, // Репозиторий для блоков карты
	playerTilesetRepo processed.ITilesetRepository, // Репозиторий для игрока
	bulletTilesetRepo processed.ITilesetRepository, // Репозиторий для пуль
) (GameState, error) {
	level, err := mapsRepo.GetLevel(2)
	if err != nil {
		return GameState{}, err
	}

	// Создаем репозитории
	blocksRepo := game.NewBlocksRepository()
	bulletsRepo := game.NewBulletsRepository()

	// Заполняем репозиторий блоков данными уровня
	for _, block := range level {
		blocksRepo.AddBlock(block)
	}

	// Создаем Use Cases
	playerUseCases := use_cases.NewPlayerUseCases(playerTilesetRepo)
	bulletTilesUseCases := use_cases.NewTilesUseCases(bulletTilesetRepo)
	bulletUseCases := use_cases.NewBulletUseCases(bulletsRepo, bulletTilesUseCases)
	mapUseCases := use_cases.NewMapUseCases(blocksRepo)
	collisionUseCases := use_cases.NewCollisionUseCases(
		bulletUseCases,
		playerUseCases,
		mapUseCases,
	)

	// Создаем TilesUseCases для рендерера
	// Для карты используем репозиторий блоков
	mapTilesUseCases := use_cases.NewTilesUseCases(mapTilesetRepo)
	playerTilesUseCases := use_cases.NewTilesUseCases(playerTilesetRepo)
	bulletTilesUseCasesForRenderer := use_cases.NewTilesUseCases(bulletTilesetRepo)

	rendererAdapter := adapters.NewRendererAdapter(
		mapUseCases,
		playerUseCases,
		bulletUseCases,
		mapTilesUseCases,
		playerTilesUseCases,
		bulletTilesUseCasesForRenderer,
	)
	inputAdapter := adapters.NewInputAdapter(
		playerUseCases,
		bulletUseCases,
		ebiten.KeyW,     // up
		ebiten.KeyS,     // down
		ebiten.KeyA,     // left
		ebiten.KeyD,     // right
		ebiten.KeySpace, // shoot
	)

	return GameState{
		playerUseCases:    playerUseCases,
		bulletUseCases:    bulletUseCases,
		mapUseCases:       mapUseCases,
		collisionUseCases: collisionUseCases,
		inputAdapter:      inputAdapter,
		rendererAdapter:   rendererAdapter,
	}, nil
}

func (state GameState) Update() (State, error) {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return nil, errors.New("exit application")
	}

	state.inputAdapter.Update()
	state.playerUseCases.MovePlayer(state.playerUseCases.GetDirection(), use_cases.DT)
	state.playerUseCases.UpdateTankAnimation()
	state.bulletUseCases.UpdateBullets(use_cases.DT)
	state.collisionUseCases.UpdateCollisions()
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
	state.rendererAdapter.DrawAll(screen)
}
