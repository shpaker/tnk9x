package states

import (
	"errors"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/gonflict/internal/adapters"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// GameState представляет игровое состояние уровня
// Объединяет логику управления игровыми объектами и интеграцию с Ebiten
type GameState struct {
	// Entities
	PlayerTank *types.TankEntity
	EnemyTanks [3]*types.TankEntity
	HQEntity   *types.HQEntity

	// Use Cases
	HQUseCases            interfaces.IHQUseCases
	TankActionsUseCases   interfaces.ITankActionsUseCases   // Общий use case для действий танков
	TankCommonUseCases    interfaces.ITankCommonUseCases    // Общий use case для движения танков
	TankRenderUseCases    interfaces.ITankRenderUseCases    // Общий use case для рендеринга танков
	TankLifecycleUseCases interfaces.ITankLifecycleUseCases // Общий use case для жизненного цикла танков
	BulletUseCases        *use_cases.BulletUseCases
	MapUseCases           *use_cases.MapUseCases
	CollisionUseCases     *use_cases.CollisionUseCases
	TilesUseCases         *use_cases.TilesUseCases
	AIContext             *types.GameAiContext

	// Адаптеры
	InputAdapter       interfaces.IInputAdapter
	RendererAdapter    *adapters.GameStateRendererAdapter
	EnemyInputAdapters []interfaces.IInputAdapter // AI адаптеры врагов

	// Метаданные
	StartTime time.Time
}

// NewGameState создает новое состояние игры через билдер
func NewGameState(
	mapsRepo interfaces.IMapsDataRepository,
	scriptsRepo interfaces.IScriptsRepository,
	levelNumber int,
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	config interfaces.IConfigProvider,
	gameRepo interfaces.IGameRepositoriesRegistry,
	boundaryCollisionService interfaces.IBoundaryCollisionService,
	wallCollisionService interfaces.IWallCollisionService,
	coordinateService interfaces.ICoordinateService,
	tankBrakingService interfaces.ITankBrakingService,
	rendererAdapter *adapters.GameStateRendererAdapter,
	inputAdapter interfaces.IInputAdapter,
) (*GameState, error) {
	builder := NewGameStateBuilder(
		mapsRepo,
		scriptsRepo,
		levelNumber,
		tilesetRegistry,
		config,
		gameRepo,
		boundaryCollisionService,
		wallCollisionService,
		coordinateService,
		tankBrakingService,
		rendererAdapter,
		inputAdapter,
	)
	return builder.Build()
}

// StartTankSpawn запускает спавн танка игрока
func (state *GameState) StartTankSpawn() {
	spawnStartTime := 0.0
	if state.TankLifecycleUseCases != nil && state.PlayerTank != nil {
		err := state.TankLifecycleUseCases.Spawn(state.PlayerTank)
		if err != nil {
			panic(err)
		}
		state.PlayerTank.SpawnedAt = spawnStartTime
	}
}

func (state *GameState) Update() (State, error) {
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
	elapsedTime := time.Since(state.StartTime).Seconds()
	state.UpdateTankSpawn(elapsedTime)

	// Обновляем спавн врагов
	state.UpdateEnemiesSpawn(elapsedTime)

	// Обновляем input
	state.InputAdapter.Update(dt)

	// Обновляем игровое состояние
	state.update(dt)

	// Обновляем все анимации
	state.UpdateAnimations()

	return nil, nil
}

func (state *GameState) Draw(screen *ebiten.Image) {
	state.RendererAdapter.DrawAll(screen)
}

// Update обновляет игровое состояние
func (state *GameState) update(dt float64) {
	// Обновляем игрока
	if state.TankCommonUseCases != nil && state.PlayerTank != nil {
		if err := state.TankCommonUseCases.Update(state.PlayerTank, dt); err != nil {
			_ = err
		}
	}

	// Обновляем контекст AI с данными об игроке, врагах и пулях
	if state.AIContext != nil {
		bullets := state.BulletUseCases.GetBullets()
		state.AIContext.Player = state.PlayerTank
		state.AIContext.Enemies = state.getEnemyTanks()
		state.AIContext.Bullets = bullets
	}

	// Обновляем AI input адаптеры врагов
	for _, adapter := range state.EnemyInputAdapters {
		if adapter != nil {
			adapter.Update(dt)
		}
	}

	// Обновляем движение врагов
	for i := range state.EnemyTanks {
		if state.TankCommonUseCases != nil &&
			state.EnemyTanks[i] != nil {
			if err := state.TankCommonUseCases.Update(state.EnemyTanks[i], dt); err != nil {
				_ = err
			}
		}
	}

	// Обновляем пули
	_ = state.BulletUseCases.UpdateBullets(dt)

	// Проверяем коллизии ПОСЛЕ движения всех объектов
	_ = state.CollisionUseCases.UpdateCollisions()

	// Проверяем завершение анимации взрыва базы
	if state.HQUseCases != nil {
		state.HQUseCases.IsExplosionFinished()
	}
}

// UpdateAnimations обновляет все анимации из репозитория
func (state *GameState) UpdateAnimations() {
	if state.TilesUseCases != nil {
		state.TilesUseCases.UpdateAnimations()
	}
}

// UpdateTankSpawn обновляет процесс спавна танка
func (state *GameState) UpdateTankSpawn(currentTime float64) {
	if state.TankLifecycleUseCases != nil && state.PlayerTank != nil {
		state.TankLifecycleUseCases.IsSpawnFinished(
			state.PlayerTank,
			currentTime,
		)
		state.TankLifecycleUseCases.IsExplosionFinished(state.PlayerTank)
	}
}

// UpdateEnemiesSpawn обновляет процесс спавна врагов
func (state *GameState) UpdateEnemiesSpawn(currentTime float64) {
	for i := range state.EnemyTanks {
		if state.TankLifecycleUseCases != nil &&
			state.EnemyTanks[i] != nil {
			state.TankLifecycleUseCases.IsSpawnFinished(
				state.EnemyTanks[i],
				currentTime,
			)
			state.TankLifecycleUseCases.IsExplosionFinished(state.EnemyTanks[i])
		}
	}
}

// getEnemyTanks возвращает список не-nil врагов для AI контекста
func (state *GameState) getEnemyTanks() []*types.TankEntity {
	result := make([]*types.TankEntity, 0, 3)
	for _, tank := range state.EnemyTanks {
		if tank != nil {
			result = append(result, tank)
		}
	}
	return result
}
