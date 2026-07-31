package states

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

// StageRenderer — контракт рендера уровня, определён у потребителя,
// чтобы не тащить ebiten в пакет контрактов
type StageRenderer interface {
	DrawAll(screen *ebiten.Image)
	DrawSidebar(screen *ebiten.Image, hud types.StageHUDData)
	DrawPauseOverlay(screen *ebiten.Image)
	DrawStageEndOverlay(screen *ebiten.Image, label string)
}

// StageStateDependencies — готовый граф зависимостей уровня;
// собирается composition root'ом, все поля обязательны
type StageStateDependencies struct {
	// Use Cases
	TankCommonUseCases    interfaces.ITankCommonUseCases
	RenderUseCases        interfaces.IRenderUseCases
	TankLifecycleUseCases interfaces.ITankLifecycleUseCases
	TilesUseCases         interfaces.ITilesUseCases
	StageUseCases         interfaces.IStageUseCases
	SoundUseCases         interfaces.ISoundUseCases

	// Adapters
	InputAdapters      [2]interfaces.IInputAdapter
	EnemyInputAdapter  interfaces.IAiInputAdapter
	Renderer           StageRenderer
	SoundPlayerAdapter interfaces.ISoundPlayerAdapter

	// Session & Repositories
	StageSession      *session_entities.StageSessionEntity
	BonusesRepository interfaces.IBonusesRepository
}

type StageState struct {
	// Use Cases
	tankCommonUseCases    interfaces.ITankCommonUseCases
	renderUseCases        interfaces.IRenderUseCases
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases
	tilesUseCases         interfaces.ITilesUseCases
	stageUseCases         interfaces.IStageUseCases
	soundUseCases         interfaces.ISoundUseCases

	// Adapters
	inputAdapters      [2]interfaces.IInputAdapter
	enemyInputAdapter  interfaces.IAiInputAdapter
	renderer           StageRenderer
	soundPlayerAdapter interfaces.ISoundPlayerAdapter

	// Session & Repositories
	stageSession      *session_entities.StageSessionEntity
	bonusesRepository interfaces.IBonusesRepository

	isSetUp         bool
	endSoundHandled bool
	debugEnabled    bool // Флаг дебаг-режима
}

func NewStageState(deps StageStateDependencies) *StageState {
	return &StageState{
		tankCommonUseCases:    deps.TankCommonUseCases,
		renderUseCases:        deps.RenderUseCases,
		tankLifecycleUseCases: deps.TankLifecycleUseCases,
		tilesUseCases:         deps.TilesUseCases,
		stageUseCases:         deps.StageUseCases,
		soundUseCases:         deps.SoundUseCases,
		inputAdapters:         deps.InputAdapters,
		enemyInputAdapter:     deps.EnemyInputAdapter,
		renderer:              deps.Renderer,
		soundPlayerAdapter:    deps.SoundPlayerAdapter,
		stageSession:          deps.StageSession,
		bonusesRepository:     deps.BonusesRepository,
	}
}

// SetDebugEnabled устанавливает флаг дебаг-режима
func (state *StageState) SetDebugEnabled(enabled bool) {
	state.debugEnabled = enabled
}

func (state *StageState) SetUp() {
	// Сбрасываем флаг звуков завершения уровня для нового уровня
	state.endSoundHandled = false

	// Глушим остатки прошлого уровня и запускаем стартовый звук
	state.soundUseCases.RequestStopAll()
	state.soundUseCases.RequestSound(types.SoundIDGameStart, false)

	// Сбрасываем сессию перед спавном танков, чтобы восстановить жизни игроков
	state.stageSession.Reset()

	playerCount := int(state.stageSession.GetPlayerCount())
	if playerCount < 1 {
		playerCount = 1
	}
	if playerCount > 2 {
		playerCount = 2
	}

	for i := 0; i < playerCount; i++ {
		num := types.PlayerTankNum(i)
		role := types.PlayerTankNumToRole(num)
		playerTank := state.stageUseCases.SpawnPlayerTank(role)

		if playerTank != nil && state.inputAdapters[i] != nil {
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
		state.enemyInputAdapter.AddTank(enemy)
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
		playerTanks := state.tankCommonUseCases.GetAllPlayerTanks()
		// Повышаем уровень всех активных танков игроков
		for _, tank := range playerTanks {
			if tank != nil && tank.IsActive() {
				state.tankCommonUseCases.LevelUp(tank)
				// UpdateTankAnimation вызывается внутри LevelUp
			}
		}
	}

	for _, adapter := range state.inputAdapters {
		if adapter != nil {
			adapter.Update(dt)
		}
	}

	stageFinished := state.stageUseCases.IsStageFinished()
	if stageFinished && !state.stageUseCases.IsPaused() {
		state.stageUseCases.PauseStageState()
	}
	if stageFinished {
		// Один раз глушим все звуки (в т.ч. луп двигателя) при завершении
		// уровня; при поражении дополнительно проигрываем gameover
		if !state.endSoundHandled {
			state.soundUseCases.RequestStopAll()
			if !state.stageUseCases.IsStageWon() {
				state.soundUseCases.RequestSound(types.SoundIDGameOver, false)
			}
			state.endSoundHandled = true
		}
		// Оверлей закрывается любой клавишей или тапом по экрану
		if len(inpututil.AppendJustPressedKeys(nil)) > 0 ||
			len(inpututil.AppendJustPressedTouchIDs(nil)) > 0 {
			// Глушим звук завершения при выходе на экран выбора уровня
			state.soundUseCases.RequestStopAll()
			transition = types.StateTransition{
				Target: types.TransitionToStageSelect,
			}
		}
	}

	paused := state.stageUseCases.IsPaused()

	if !paused {
		_ = state.tankLifecycleUseCases.UpdateAllTanksLifecycle()
		_ = state.tankCommonUseCases.UpdateAllTanks(dt)
		state.stageUseCases.UpdateGameObjects(dt)

		respawned1, respawned2 := state.stageUseCases.TryRespawnPlayersTanks()
		respawnedTanks := []*types.TankEntity{respawned1, respawned2}

		for i, respawned := range respawnedTanks {
			if respawned != nil && state.inputAdapters[i] != nil {
				if keyboardAdapter, ok := state.inputAdapters[i].(interfaces.IInputAdapterWithTank); ok {
					keyboardAdapter.SetPlayerTank(respawned)
				}
			}
		}

		if state.stageUseCases.IsStageFinished() &&
			!state.stageUseCases.IsPaused() {
			state.stageUseCases.PauseStageState()
		}

		if spawned := state.stageUseCases.TrySpawnEnemy(); spawned != nil {
			state.enemyInputAdapter.AddTank(spawned)
		}

		// Замороженные бонусом-таймером враги не получают команд AI
		if !state.stageSession.AreEnemiesFrozen() {
			state.enemyInputAdapter.Update(dt)
		}

		state.tilesUseCases.UpdateAnimations()

		state.updateBlinkObjects()

		// Управление звуком двигателя
		if state.tankCommonUseCases.IsAnyPlayerTankMoving() {
			// Запускаем звук двигателя с зацикливанием, если он еще не играет
			state.soundUseCases.RequestSound(types.SoundIDEngine, true)
		} else {
			// Останавливаем звук двигателя, когда все игроки остановлены
			state.soundUseCases.RequestStop(types.SoundIDEngine)
		}
	}

	// Единственная точка контакта с звуковым адаптером: применяем
	// накопленные события кадра в порядке добавления
	for _, event := range state.soundUseCases.GetEvents() {
		state.applySoundEvent(event)
	}
	state.soundPlayerAdapter.Update()

	return transition
}

func (state *StageState) applySoundEvent(event types.SoundEntity) {
	var err error
	switch event.Action {
	case types.SoundActionPlay:
		err = state.soundPlayerAdapter.Play(event.SoundID)
	case types.SoundActionPlayLoop:
		err = state.soundPlayerAdapter.PlayLoop(event.SoundID)
	case types.SoundActionStop:
		state.soundPlayerAdapter.Stop(event.SoundID)
	case types.SoundActionStopAll:
		state.soundPlayerAdapter.StopAll()
	}
	// Ошибки воспроизведения не фатальны: логируем и продолжаем
	if err != nil {
		log.Printf("sound %q: %v", event.SoundID, err)
	}
}

// updateBlinkObjects обновляет мигание бонусов, танков с бонусом
// и танков под щитом
func (state *StageState) updateBlinkObjects() {
	var blinkObjects []types.IBlink

	for _, bonus := range state.bonusesRepository.GetAllBonuses() {
		if bonus != nil {
			blinkObjects = append(blinkObjects, bonus)
		}
	}

	for _, tank := range state.tankCommonUseCases.GetAllTanks() {
		if tank == nil {
			continue
		}
		if (tank.IsEnemy() && tank.GetWithBonus()) || tank.HasShield() {
			blinkObjects = append(blinkObjects, tank)
		}
	}

	if len(blinkObjects) > 0 {
		state.renderUseCases.UpdateBlink(blinkObjects)
	}
}

func (state *StageState) Draw(screen *ebiten.Image) {
	state.renderer.DrawAll(screen)

	state.renderer.DrawSidebar(screen, types.StageHUDData{
		EnemiesForSpawn: state.stageSession.EnemiesForSpawnCount(),
		PlayerCount:     state.stageSession.GetPlayerCount(),
		Player1Lives: state.stageSession.GetPlayerLives(
			types.PlayerTankNumPlayer1,
		),
		Player2Lives: state.stageSession.GetPlayerLives(
			types.PlayerTankNumPlayer2,
		),
		StageNumber: state.stageSession.GetStageNumber(),
	})

	if state.stageUseCases.IsStageFinished() {
		label := "VICTORY"
		if !state.stageUseCases.IsStageWon() {
			label = "DEFEAT"
		}
		state.renderer.DrawStageEndOverlay(screen, label)
		return
	}

	if state.stageUseCases.IsPaused() {
		state.renderer.DrawPauseOverlay(screen)
	}
}
