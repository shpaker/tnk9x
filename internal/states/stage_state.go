package states

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
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
	RenderUseCases        interfaces.IRenderUseCases
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

	session           *session_entities.GameSessionEntity
	bonusesRepository interfaces.IBonusesRepository

	soundUseCases      *use_cases.SoundUseCases
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
	fontUseCases interfaces.IFontUseCases,
	gameRepository interfaces.IGameRepositoriesRegistry,
	boundaryCollisionService interfaces.IBoundaryCollisionService,
	wallCollisionService interfaces.IWallCollisionService,
	coordinateService interfaces.ICoordinateService,
	tankBrakingService interfaces.ITankBrakingService,
	luaEngine interfaces.ILuaEngine,
	session *session_entities.GameSessionEntity,
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
		coordinateService,
		tankBrakingService,
		fontUseCases,
		luaEngine,
		session,
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
	if state.session != nil && state.session.StageSession() != nil {
		state.session.StageSession().Reset()
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
			if state.session != nil &&
				len(inpututil.AppendJustPressedKeys(nil)) > 0 {
				target := types.StateTypeStageSelect
				state.session.SetTargetState(&target)
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
