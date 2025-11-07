package states

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	game "github.com/shpaker/gonflict/internal/adapters/game"
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
	TankSpawnUseCases     interfaces.ITankSpawnUseCases     // Use case для спавна танков уровня
	BulletUseCases        *use_cases.BulletUseCases
	MapUseCases           *use_cases.MapUseCases
	CollisionUseCases     *use_cases.CollisionUseCases
	TilesUseCases         *use_cases.TilesUseCases

	// Адаптеры
	InputAdapter       interfaces.IInputAdapter
	RendererAdapter    *game.GameRendererAdapter
	EnemyInputAdapters []interfaces.IInputAdapter // AI адаптеры врагов

	// Метаданные
	StartTime time.Time
	isSetUp   bool // Флаг для отслеживания, был ли вызван SetUp

	// Сессия
	Session *types.SessionEntity
}

// NewGameState создает новое состояние игры через билдер
func NewGameState(
	mapsRepository interfaces.IMapsDataRepository,
	scriptsRepository interfaces.IScriptsRepository,
	levelNumber int,
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	config interfaces.IConfigProvider,
	gameRepository interfaces.IGameRepositoriesRegistry,
	boundaryCollisionService interfaces.IBoundaryCollisionService,
	wallCollisionService interfaces.IWallCollisionService,
	coordinateService interfaces.ICoordinateService,
	tankBrakingService interfaces.ITankBrakingService,
	rendererAdapter *game.GameRendererAdapter,
	inputAdapter interfaces.IInputAdapter,
	luaEngine interfaces.ILuaEngine,
	session *types.SessionEntity,
) (*GameState, error) {
	builder := NewGameStateBuilder(
		mapsRepository,
		scriptsRepository,
		levelNumber,
		tilesetRegistry,
		config,
		gameRepository,
		boundaryCollisionService,
		wallCollisionService,
		coordinateService,
		tankBrakingService,
		rendererAdapter,
		inputAdapter,
		luaEngine,
		session,
	)
	return builder.Build()
}

// SetUp запускается один раз на старте состояния
func (state *GameState) SetUp() {
	if state.TankSpawnUseCases == nil {
		return
	}
	if err := state.TankSpawnUseCases.StageSetUp(); err != nil {
		_ = err
	}
}

// Update обновляет состояние игры (вызывается Ebiten каждый кадр)
func (state *GameState) Update() {
	// Вызываем SetUp один раз на старте состояния
	if !state.isSetUp {
		state.SetUp()
		state.isSetUp = true
	}

	// Вычисляем delta time из ActualTPS
	tps := ebiten.ActualTPS()
	var dt float64
	if tps > 0 {
		dt = 1.0 / tps
	} else {
		dt = 1.0 / 60.0 // Fallback если TPS равен 0
	}

	// Обновляем жизненный цикл танков (спавн и взрыв)
	if state.TankLifecycleUseCases != nil {
		if err := state.TankLifecycleUseCases.UpdateAllTanksLifecycle(); err != nil {
			_ = err
		}
	}

	// Обновляем ввод
	state.InputAdapter.Update(dt)

	// Обновляем игровые объекты (танки, пули, коллизии, AI)
	state.updateGameObjects(dt)

	// Обновляем все анимации
	if state.TilesUseCases != nil {
		state.TilesUseCases.UpdateAnimations()
	}
}

func (state *GameState) Draw(screen *ebiten.Image) {
	state.RendererAdapter.DrawAll(screen)
}

// updateGameObjects обновляет игровые объекты (танки, пули, коллизии, AI)
func (state *GameState) updateGameObjects(dt float64) {
	// Обновляем все танки (игрок + враги) через use cases
	if state.TankCommonUseCases != nil {
		if err := state.TankCommonUseCases.UpdateAllTanks(dt); err != nil {
			_ = err
		}
	}

	// Обновляем пули
	_ = state.BulletUseCases.UpdateBullets(dt)

	// Проверяем коллизии ПОСЛЕ движения всех объектов
	state.CollisionUseCases.UpdateCollisions()

	// Обновляем AI адаптеры ввода врагов ПОСЛЕ коллизий
	// чтобы AI видел актуальное состояние танка после столкновений
	for _, enemyInputAdapter := range state.EnemyInputAdapters {
		if enemyInputAdapter != nil {
			enemyInputAdapter.Update(dt)
		}
	}

	// Проверяем завершение анимации взрыва базы
	if state.HQUseCases != nil {
		state.HQUseCases.IsExplosionFinished(state.HQEntity)
	}
}
