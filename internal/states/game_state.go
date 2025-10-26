package states

import (
	"errors"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/adapters"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/use_cases"
)

type GameState struct {
	gameStateServices *GameStateUseCasesFacade
	inputAdapter      *adapters.InputAdapter
	rendererAdapter   *adapters.RendererAdapter
	startTime         time.Time // Время начала игры для отслеживания спавна
}

// NewGameState создает новое состояние игры с переданным репозиторием карт
func NewGameState(
	mapsRepo processed.IMapsDataRepository,
	mapTilesetRepo processed.ITilesetRepository, // Репозиторий для блоков карты
	playerTilesetRepo processed.ITilesetRepository, // Репозиторий для игрока
	bulletTilesetRepo processed.ITilesetRepository, // Репозиторий для пуль
	spawnerTilesetRepo processed.ITilesetRepository, // Репозиторий для спавна
) (GameState, error) {
	// Создаем GameStateUseCasesFacade
	gameStateServices, err := NewGameStateUseCasesFacade(
		mapsRepo,
		13, // Номер уровня
		mapTilesetRepo,
		playerTilesetRepo,
		bulletTilesetRepo,
		spawnerTilesetRepo,
	)
	if err != nil {
		return GameState{}, err
	}

	// Создаем адаптеры
	rendererAdapter := createRendererAdapter(
		gameStateServices,
		mapTilesetRepo,
		playerTilesetRepo,
		bulletTilesetRepo,
		spawnerTilesetRepo,
	)

	inputAdapter := createInputAdapter(gameStateServices)

	gameState := GameState{
		gameStateServices: gameStateServices,
		inputAdapter:      inputAdapter,
		rendererAdapter:   rendererAdapter,
		startTime:         time.Now(),
	}

	// Запускаем спавн танка на старте
	gameState.StartTankSpawn()

	return gameState, nil
}

// StartTankSpawn запускает спавн танка
func (state GameState) StartTankSpawn() {
	spawnStartTime := 0.0
	state.gameStateServices.StartTankSpawn(spawnStartTime)
}

func (state GameState) Update() (State, error) {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return nil, errors.New("exit application")
	}

	// Обновляем спавн танка
	elapsedTime := time.Since(state.startTime).Seconds()
	state.gameStateServices.UpdateTankSpawn(elapsedTime)

	// Обновляем input
	state.inputAdapter.Update()

	// Обновляем игровое состояние
	state.gameStateServices.Update()

	return nil, nil
}

func (state GameState) Draw(screen *ebiten.Image) {
	state.rendererAdapter.DrawAll(screen)
}

// createInputAdapter создает адаптер ввода
func createInputAdapter(gameStateServices *GameStateUseCasesFacade) *adapters.InputAdapter {
	return adapters.NewInputAdapter(
		gameStateServices.TankUseCases(),
		gameStateServices.BulletUseCases(),
		ebiten.KeyW,     // up
		ebiten.KeyS,     // down
		ebiten.KeyA,     // left
		ebiten.KeyD,     // right
		ebiten.KeySpace, // shoot
	)
}

// createRendererAdapter создает адаптер рендерера
func createRendererAdapter(
	gameStateServices *GameStateUseCasesFacade,
	mapTilesetRepo processed.ITilesetRepository,
	playerTilesetRepo processed.ITilesetRepository,
	bulletTilesetRepo processed.ITilesetRepository,
	spawnerTilesetRepo processed.ITilesetRepository,
) *adapters.RendererAdapter {
	mapTilesUseCases := use_cases.NewTilesUseCases(mapTilesetRepo)
	playerTilesUseCases := use_cases.NewTilesUseCases(playerTilesetRepo)
	bulletTilesUseCases := use_cases.NewTilesUseCases(bulletTilesetRepo)
	spawnerTilesUseCases := use_cases.NewTilesUseCases(spawnerTilesetRepo)

	return adapters.NewRendererAdapter(
		gameStateServices.MapUseCases(),
		gameStateServices.TankUseCases(),
		gameStateServices.BulletUseCases(),
		mapTilesUseCases,
		playerTilesUseCases,
		bulletTilesUseCases,
		spawnerTilesUseCases,
	)
}
