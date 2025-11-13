package states

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	game "github.com/shpaker/tnk25/internal/adapters/stage"
	"github.com/shpaker/tnk25/internal/interfaces"
	"github.com/shpaker/tnk25/internal/types"
	"github.com/shpaker/tnk25/internal/types/session_entities"
	"github.com/shpaker/tnk25/internal/use_cases"
)

type StageState struct {
	HQEntity *types.HQEntity

	HQUseCases            interfaces.IHQUseCases
	TankActionsUseCases   interfaces.ITankActionsUseCases
	TankCommonUseCases    interfaces.ITankCommonUseCases
	TankRenderUseCases    interfaces.ITankRenderUseCases
	TankLifecycleUseCases interfaces.ITankLifecycleUseCases
	BulletUseCases        interfaces.IBulletUseCases
	MapUseCases           *use_cases.MapUseCases
	CollisionUseCases     *use_cases.CollisionUseCases
	TilesUseCases         *use_cases.TilesUseCases

	inputAdapters     []interfaces.IInputAdapter
	RendererAdapter   *game.StageRendererAdapter
	EnemyInputAdapter interfaces.IAiInputAdapter

	StartTime time.Time
	isSetUp   bool

	stageUseCases interfaces.IStageUseCases

	session *session_entities.GameSessionEntity
}

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

func (state *StageState) SetUp() {
	if state.stageUseCases == nil {
		return
	}

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

func (state *StageState) Update() {
	if !state.isSetUp {
		state.SetUp()
		state.isSetUp = true
	}

	tps := ebiten.ActualTPS()
	var dt float64
	if tps > 0 {
		dt = 1.0 / tps
	} else {
		dt = 1.0 / 60.0
	}

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

	if state.TankLifecycleUseCases != nil && state.stageUseCases != nil &&
		!paused {
		_ = state.TankLifecycleUseCases.UpdateAllTanksLifecycle()
		_ = state.TankCommonUseCases.UpdateAllTanks(dt)
		state.stageUseCases.UpdateGameObjects(dt)

		respawned1, respawned2 := state.stageUseCases.TryRespawnPlayersTanks()
		respawnedTanks := []*types.TankEntity{respawned1, respawned2}

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
