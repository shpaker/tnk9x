package states

import (
	"errors"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/gonflict/internal/adapters"
	"github.com/shpaker/gonflict/internal/interfaces"
)

type GameState struct {
	gameStateServices *GameStateUseCasesFacade
	inputAdapter      interfaces.IInputAdapter
	rendererAdapter   *adapters.RendererAdapter
	startTime         time.Time // Время начала игры для отслеживания спавна
}

// NewGameState создает новое состояние игры с переданным реестром тайлсетов
func NewGameState(
	mapsRepo interfaces.IMapsDataRepository,
	scriptsRepo interfaces.IScriptsRepository,
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	gameRepo interfaces.IGameRepositoriesRegistry,
	rendererAdapter *adapters.RendererAdapter,
	inputAdapter interfaces.IInputAdapter,
	gameStateServices *GameStateUseCasesFacade,
) (GameState, error) {
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

	// Вычисляем delta time из ActualTPS
	tps := ebiten.ActualTPS()
	var dt float64
	if tps > 0 {
		dt = 1.0 / tps
	} else {
		dt = 1.0 / 60.0 // Fallback если TPS равен 0
	}

	// Обновляем спавн танка
	elapsedTime := time.Since(state.startTime).Seconds()
	state.gameStateServices.UpdateTankSpawn(elapsedTime)

	// Обновляем спавн врагов
	state.gameStateServices.UpdateEnemiesSpawn(elapsedTime)

	// Обновляем input
	state.inputAdapter.Update(dt)

	// Обновляем игровое состояние
	state.gameStateServices.Update(dt)

	// Обновляем все анимации
	state.gameStateServices.UpdateAnimations()

	return nil, nil
}

func (state GameState) Draw(screen *ebiten.Image) {
	state.rendererAdapter.DrawAll(screen)
}
