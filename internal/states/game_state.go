package states

import (
	"errors"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/gonflict/internal/adapters"
	"github.com/shpaker/gonflict/internal/adapters/input_adapters"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// GameConfig представляет конфигурацию игры (для избежания циклических импортов)
type GameConfig struct {
	EnemySpawners         [][2]int `yaml:"enemy_spawners"`
	PlayerSpawners        [][2]int `yaml:"players_spawners"`
	AIUpdateIntervalTicks int      `yaml:"ai_update_interval_ticks"` // Интервал обновления AI в тиках
}

type GameState struct {
	gameStateServices *GameStateUseCasesFacade
	inputAdapter      input_adapters.IInputAdapter
	rendererAdapter   *adapters.RendererAdapter
	startTime         time.Time // Время начала игры для отслеживания спавна
}

// NewGameState создает новое состояние игры с переданным реестром тайлсетов
func NewGameState(
	mapsRepo processed.IMapsDataRepository,
	scriptsRepo processed.IScriptsRepository,
	tilesetRegistry processed.ITilesetRepositoryRegistry,
	gameConfig *GameConfig,
) (GameState, error) {
	// Создаем GameStateUseCasesFacade
	gameStateServices, err := NewGameStateUseCasesFacade(
		mapsRepo,
		scriptsRepo,
		13, // Номер уровня
		tilesetRegistry.Blocks(),
		tilesetRegistry.Player(),
		tilesetRegistry.Bullet(),
		tilesetRegistry.Spawner(),
		tilesetRegistry.Explosion(),
		gameConfig,
	)
	if err != nil {
		return GameState{}, err
	}

	// Создаем адаптеры
	rendererAdapter := createRendererAdapter(
		gameStateServices,
		tilesetRegistry,
	)

	// Используем клавиатурный адаптер
	inputAdapter := input_adapters.NewKeyboardInputAdapter(
		gameStateServices.TankUseCases(),
		ebiten.KeyW,     // up
		ebiten.KeyS,     // down
		ebiten.KeyA,     // left
		ebiten.KeyD,     // right
		ebiten.KeySpace, // shoot
	)

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

	// Обновляем спавн врагов
	state.gameStateServices.UpdateEnemiesSpawn(elapsedTime)

	// Обновляем input
	state.inputAdapter.Update()

	// Обновляем игровое состояние
	state.gameStateServices.Update()

	// Обновляем все анимации
	state.gameStateServices.UpdateAnimations()

	return nil, nil
}

func (state GameState) Draw(screen *ebiten.Image) {
	state.rendererAdapter.DrawAll(screen)
}

// createRendererAdapter создает адаптер рендерера
func createRendererAdapter(
	gameStateServices *GameStateUseCasesFacade,
	tilesetRegistry processed.ITilesetRepositoryRegistry,
) *adapters.RendererAdapter {
	mapTilesUseCases := use_cases.NewTilesUseCases(tilesetRegistry.Blocks())
	playerTilesUseCases := use_cases.NewTilesUseCases(tilesetRegistry.Player())
	bulletTilesUseCases := use_cases.NewTilesUseCases(tilesetRegistry.Bullet())
	spawnerTilesUseCases := use_cases.NewTilesUseCases(
		tilesetRegistry.Spawner(),
	)
	explosionTilesUseCases := use_cases.NewTilesUseCases(
		tilesetRegistry.Explosion(),
	)

	return adapters.NewRendererAdapter(
		gameStateServices.MapUseCases(),
		gameStateServices.TankUseCases(),
		gameStateServices.BulletUseCases(),
		gameStateServices.GetEnemyTanks(),
		gameStateServices.GetEnemyUseCases(),
		mapTilesUseCases,
		playerTilesUseCases,
		bulletTilesUseCases,
		spawnerTilesUseCases,
		explosionTilesUseCases,
	)
}
