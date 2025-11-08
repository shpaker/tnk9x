package states

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	game "github.com/shpaker/gonflict/internal/adapters/game"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"

	"github.com/shpaker/gonflict/internal/types/session_entities"
)

// StageState представляет состояние уровня
// Объединяет логику управления игровыми объектами и интеграцию с Ebiten
type StageState struct {
	// Entities
	HQEntity *types.HQEntity

	// Use Cases
	HQUseCases            interfaces.IHQUseCases
	TankActionsUseCases   interfaces.ITankActionsUseCases   // Общий use case для действий танков
	TankCommonUseCases    interfaces.ITankCommonUseCases    // Общий use case для движения танков
	TankRenderUseCases    interfaces.ITankRenderUseCases    // Общий use case для рендеринга танков
	TankLifecycleUseCases interfaces.ITankLifecycleUseCases // Use case для жизненного цикла и спавна
	BulletUseCases        interfaces.IBulletUseCases
	MapUseCases           *use_cases.MapUseCases
	CollisionUseCases     *use_cases.CollisionUseCases
	TilesUseCases         *use_cases.TilesUseCases

	// Адаптеры
	InputAdapter      interfaces.IInputAdapter
	RendererAdapter   *game.StageRendererAdapter
	EnemyInputAdapter interfaces.IAiInputAdapter // AI адаптер врагов

	// Метаданные
	StartTime time.Time
	isSetUp   bool // Флаг для отслеживания, был ли вызван SetUp

	// Сессия
	Session *session_entities.GameSessionEntity

	// Use cases
	stageUseCases interfaces.IStageUseCases
}

// NewStageState создает новое состояние уровня через билдер
func NewStageState(
	mapsRepository interfaces.IMapsDataRepository,
	scriptsRepository interfaces.IScriptsRepository,
	levelNumber int,
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	config interfaces.IConfigProvider,
	fontUseCases interfaces.IFontUseCases,
	gameRepository interfaces.IGameRepositoriesRegistry,
	boundaryCollisionService interfaces.IBoundaryCollisionService,
	wallCollisionService interfaces.IWallCollisionService,
	coordinateService interfaces.ICoordinateService,
	tankBrakingService interfaces.ITankBrakingService,
	luaEngine interfaces.ILuaEngine,
	session *session_entities.GameSessionEntity,
) (*StageState, error) {
	builder := NewStageStateBuilder(
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
		fontUseCases,
		luaEngine,
		session,
	)
	return builder.Build()
}

// SetUp запускается один раз на старте состояния
func (state *StageState) SetUp() {
	if state.stageUseCases == nil {
		return
	}
	playerTank := state.stageUseCases.SpawnPlayerTank()
	if playerTank != nil && state.InputAdapter != nil {
		if keyboardAdapter, ok := state.InputAdapter.(interface {
			SetPlayerTank(*types.TankEntity)
		}); ok {
			keyboardAdapter.SetPlayerTank(playerTank)
		}
	}
	state.stageUseCases.SpawnEnemyTanks()
}

// Update обновляет состояние игры (вызывается Ebiten каждый кадр)
func (state *StageState) Update() {
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

	// Обновляем ввод
	if state.InputAdapter != nil {
		state.InputAdapter.Update(dt)
	}

	paused := state.stageUseCases != nil && state.stageUseCases.IsPaused()

	// Обновляем жизненный цикл танков (спавн и взрыв)
	if state.TankLifecycleUseCases != nil && !paused {
		spawned, _ := state.TankLifecycleUseCases.UpdateAllTanksLifecycle()
		if spawned != nil && state.EnemyInputAdapter != nil {
			state.EnemyInputAdapter.AddTank(spawned)
		}
	}

	// Обновляем игровые объекты (танки, пули, коллизии, AI)
	if state.stageUseCases != nil {
		state.stageUseCases.UpdateGameObjects(dt)
	}

	// Обновляем все анимации
	if !paused && state.TilesUseCases != nil {
		state.TilesUseCases.UpdateAnimations()
	}
}

func (state *StageState) Draw(screen *ebiten.Image) {
	state.RendererAdapter.DrawAll(screen)
	if state.stageUseCases != nil && state.stageUseCases.IsPaused() {
		state.RendererAdapter.DrawPauseOverlay(screen)
	}
}
