package states

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/tnk9x/internal/adapters/stage"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

type StageState struct {
	HQEntity *types.HQEntity

	HQUseCases            interfaces.IHQUseCases
	TankActionsUseCases   interfaces.ITankActionsUseCases
	TankCommonUseCases    interfaces.ITankCommonUseCases
	RenderUseCases        interfaces.IRenderUseCases
	TankLifecycleUseCases interfaces.ITankLifecycleUseCases
	BulletUseCases        interfaces.IBulletUseCases
	MapUseCases           interfaces.IMapUseCases
	CollisionUseCases     interfaces.ICollisionUseCases
	TilesUseCases         interfaces.ITilesUseCases

	inputAdapters     []interfaces.IInputAdapter
	RendererAdapter   *stage.StageRendererAdapter
	EnemyInputAdapter interfaces.IAiInputAdapter

	StartTime time.Time
	isSetUp   bool

	stageUseCases interfaces.IStageUseCases

	stageSession      *session_entities.StageSessionEntity
	bonusesRepository interfaces.IBonusesRepository

	soundUseCases      interfaces.ISoundUseCases
	soundPlayerAdapter interfaces.ISoundPlayerAdapter

	defeatSoundPlayed bool
	debugEnabled      bool // Флаг дебаг-режима
}

// SetDebugEnabled устанавливает флаг дебаг-режима
func (state *StageState) SetDebugEnabled(enabled bool) {
	state.debugEnabled = enabled
}

func NewStageState(
	mapsRepository interfaces.IMapsDataRepository,
	scriptsRepository interfaces.IScriptsRepository,
	levelNumber int,
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	config interfaces.IConfigProvider,
	textFace text.Face,
	gameRepository interfaces.IGameRepositoriesRegistry,
	boundaryCollisionService interfaces.IBoundaryCollisionService,
	wallCollisionService interfaces.IWallCollisionService,
	tankBrakingService interfaces.ITankBrakingService,
	scriptEngine interfaces.IAIScriptEngine,
	stageSession *session_entities.StageSessionEntity,
	fileRepository interfaces.IFileRepository,
	audioContext *audio.Context,
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
		tankBrakingService,
		textFace,
		scriptEngine,
		stageSession,
		fileRepository,
		audioContext,
	)
	return builder.Build()
}

func (state *StageState) SetUp() {
	if state.stageUseCases == nil {
		return
	}

	// Сбрасываем флаг проигрывания звука поражения для нового уровня
	state.defeatSoundPlayed = false

	// Запускаем фоновую музыку при старте уровня
	if state.soundUseCases != nil {
		state.soundUseCases.StopAll(state.soundPlayerAdapter)
		state.soundUseCases.RequestSound(types.SoundIDGameStart, false)
	}

	// Сбрасываем сессию перед спавном танков, чтобы восстановить жизни игроков
	if state.stageSession != nil {
		state.stageSession.Reset()
	}

	playerCount := 2
	if state.stageSession != nil {
		playerCount = int(state.stageSession.GetPlayerCount())
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
			if keyboardAdapter, ok := state.inputAdapters[i].(interfaces.IInputAdapterWithTank); ok {
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

func (state *StageState) Update() types.StateTransition {
	if !state.isSetUp {
		state.SetUp()
		state.isSetUp = true
	}

	transition := types.StateTransition{}

	tps := ebiten.ActualTPS()
	var dt float64
	if tps > 0 {
		dt = 1.0 / tps
	} else {
		dt = 1.0 / 60.0
	}

	// Обработка дебаг-команд
	// Клавиша 0 повышает уровень игрока (только в режиме дебага)
	if state.debugEnabled && inpututil.IsKeyJustPressed(ebiten.KeyDigit0) {
		if state.TankCommonUseCases != nil && state.RenderUseCases != nil {
			playerTanks := state.TankCommonUseCases.GetAllPlayerTanks()
			// Повышаем уровень всех активных танков игроков
			for _, tank := range playerTanks {
				if tank != nil && tank.IsActive() {
					state.TankCommonUseCases.LevelUp(tank)
					// UpdateTankAnimation вызывается внутри LevelUp
				}
			}
		}
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
		if stageFinished {
			// Проигрываем звук поражения один раз при завершении уровня поражением
			if !state.stageUseCases.IsStageWon() &&
				!state.defeatSoundPlayed && state.soundUseCases != nil {
				// Останавливаем все звуки перед проигрыванием звука поражения
				if state.soundPlayerAdapter != nil {
					state.soundPlayerAdapter.StopAll()
				}
				// Очищаем очередь событий, чтобы не проигрывать другие звуки

				// Запрашиваем звук поражения
				state.soundUseCases.RequestSound(types.SoundIDGameOver, false)
				state.defeatSoundPlayed = true
			}
			if len(inpututil.AppendJustPressedKeys(nil)) > 0 {
				transition = types.StateTransition{
					Target: types.TransitionToStageSelect,
				}
			}
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
				if keyboardAdapter, ok := state.inputAdapters[i].(interfaces.IInputAdapterWithTank); ok {
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

	// Обновляем мигание бонусов и танков с бонусом
	if !paused && state.RenderUseCases != nil {
		var blinkObjects []types.IBlink

		// Добавляем бонусы
		if state.bonusesRepository != nil {
			bonuses := state.bonusesRepository.GetAllBonuses()
			for _, bonus := range bonuses {
				if bonus != nil {
					blinkObjects = append(blinkObjects, bonus)
				}
			}
		}

		// Добавляем танки с бонусом
		if state.TankCommonUseCases != nil {
			allTanks := state.TankCommonUseCases.GetAllTanks()
			for _, tank := range allTanks {
				if tank != nil && tank.IsEnemy() && tank.GetWithBonus() {
					blinkObjects = append(blinkObjects, tank)
				}
			}
		}

		if len(blinkObjects) > 0 {
			state.RenderUseCases.UpdateBlink(blinkObjects)
		}
	}

	// Управление звуком двигателя
	if !paused && state.TankCommonUseCases != nil &&
		state.soundUseCases != nil && state.soundPlayerAdapter != nil {
		isAnyPlayerMoving := state.TankCommonUseCases.IsAnyPlayerTankMoving()
		if isAnyPlayerMoving {
			// Запускаем звук двигателя с зацикливанием, если он еще не играет
			state.soundUseCases.RequestSound(types.SoundIDEngine, true)
		} else {
			// Останавливаем звук двигателя, когда все игроки остановлены
			_ = state.soundPlayerAdapter.Stop(types.SoundIDEngine)
		}
	}

	// Обработка звуковых событий
	if state.soundUseCases != nil && state.soundPlayerAdapter != nil {
		events := state.soundUseCases.GetEvents()
		for _, event := range events {
			if event.Loop {
				_ = state.soundPlayerAdapter.PlayLoop(event.SoundID)
			} else {
				_ = state.soundPlayerAdapter.Play(event.SoundID)
			}
		}
		_ = state.soundPlayerAdapter.Update()
	}

	return transition
}

func (state *StageState) Draw(screen *ebiten.Image) {
	state.RendererAdapter.DrawAll(screen)
	if state.stageUseCases == nil {
		return
	}

	if state.stageUseCases.IsStageFinished() {
		label := "VICTORY"
		if !state.stageUseCases.IsStageWon() {
			label = "DEFEAT"
		}
		state.RendererAdapter.DrawStageEndOverlay(screen, label)
		return
	}

	if state.stageUseCases.IsPaused() {
		state.RendererAdapter.DrawPauseOverlay(screen)
	}
}
