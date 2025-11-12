package states

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	game "github.com/shpaker/gonflict/internal/adapters/stage"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/types/session_entities"
	"github.com/shpaker/gonflict/internal/use_cases"
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
	inputAdapters     []interfaces.IInputAdapter // Адаптеры ввода игроков (массив из двух элементов)
	RendererAdapter   *game.StageRendererAdapter
	EnemyInputAdapter interfaces.IAiInputAdapter // AI адаптер врагов

	// Метаданные
	StartTime time.Time
	isSetUp   bool // Флаг для отслеживания, был ли вызван SetUp

	// Use cases
	stageUseCases interfaces.IStageUseCases

	// Сессия
	session *session_entities.GameSessionEntity
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

	// Определяем количество игроков из StageSessionEntity
	playerCount := 2
	if state.session != nil && state.session.StageSession() != nil {
		playerCount = int(state.session.StageSession().GetPlayerCount())
		if playerCount < 1 {
			playerCount = 1
		}
		if playerCount > 2 {
			playerCount = 2
		}
	}

	// Спавним игроков в зависимости от выбранного количества
	for i := 0; i < playerCount; i++ {
		num := types.PlayerTankNum(i)
		role := types.PlayerTankNumToRole(num)
		playerTank := state.stageUseCases.SpawnPlayerTank(role)

		if playerTank != nil && i < len(state.inputAdapters) &&
			state.inputAdapters[i] != nil {
			if keyboardAdapter, ok := state.inputAdapters[i].(interface {
				SetPlayerTank(*types.TankEntity)
			}); ok {
				keyboardAdapter.SetPlayerTank(playerTank)
			}
		}
	}

	enemies := state.stageUseCases.SpawnInitialEnemyTanks()
	for _, enemy := range enemies {
		if enemy == nil {
			continue
		}
		if state.EnemyInputAdapter != nil {
			state.EnemyInputAdapter.AddTank(enemy)
		}
	}
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

	// Обновляем ввод игроков через итерацию
	for _, adapter := range state.inputAdapters {
		if adapter != nil {
			adapter.Update(dt)
		}
	}

	stageFinished := false
	if state.stageUseCases != nil {
		stageFinished = state.stageUseCases.IsStageFinished()
		if stageFinished && !state.stageUseCases.IsPaused() {
			state.stageUseCases.PauseStageState()
		}
		if stageFinished && state.session != nil &&
			len(inpututil.AppendJustPressedKeys(nil)) > 0 {
			target := types.StateTypeStageSelect
			state.session.SetTargetState(&target)
		}
	}

	paused := state.stageUseCases != nil && state.stageUseCases.IsPaused()

	// Обновляем игровые объекты (танки, пули, коллизии, AI)
	if state.TankLifecycleUseCases != nil && state.stageUseCases != nil &&
		!paused {
		_ = state.TankLifecycleUseCases.UpdateAllTanksLifecycle()
		_ = state.TankCommonUseCases.UpdateAllTanks(dt)
		state.stageUseCases.UpdateGameObjects(dt)

		// Пытаемся респавнить обоих игроков через TryRespawnPlayersTanks
		respawned1, respawned2 := state.stageUseCases.TryRespawnPlayersTanks()
		respawnedTanks := []*types.TankEntity{respawned1, respawned2}

		// Обновляем адаптеры для респавненных игроков через итерацию
		for i, respawned := range respawnedTanks {
			if respawned != nil && i < len(state.inputAdapters) &&
				state.inputAdapters[i] != nil {
				if keyboardAdapter, ok := state.inputAdapters[i].(interface {
					SetPlayerTank(*types.TankEntity)
				}); ok {
					keyboardAdapter.SetPlayerTank(respawned)
				}
			}
		}

		// Проверяем завершение игры после попытки респавна (жизни могли закончиться)
		if state.stageUseCases != nil && state.stageUseCases.IsStageFinished() {
			if !state.stageUseCases.IsPaused() {
				state.stageUseCases.PauseStageState()
			}
		}

		if spawned := state.stageUseCases.TrySpawnEnemy(); spawned != nil {
			if state.EnemyInputAdapter != nil {
				state.EnemyInputAdapter.AddTank(spawned)
			}
		}
	}

	if state.EnemyInputAdapter != nil && !paused {
		state.EnemyInputAdapter.Update(dt)
	}

	// Обновляем все анимации
	if !paused && state.TilesUseCases != nil {
		state.TilesUseCases.UpdateAnimations()
	}
}

func (state *StageState) Draw(screen *ebiten.Image) {
	state.RendererAdapter.DrawAll(screen)
	if state.stageUseCases == nil {
		return
	}

	if state.stageUseCases.IsStageFinished() {
		label := "DEFEAT"
		if state.stageUseCases.IsStageWon() {
			label = "VICTORY"
		}
		state.RendererAdapter.DrawStageEndOverlay(screen, label)
		return
	}

	if state.stageUseCases.IsPaused() {
		state.RendererAdapter.DrawPauseOverlay(screen)
	}
}
