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
	DrawPauseMenu(screen *ebiten.Image, view types.PauseMenuViewData)
	PauseMenuHitTest(pos types.Position) (types.PauseMenuItem, bool)
	DrawStageResult(screen *ebiten.Image, view types.StageResultViewData)
	StageResultHitTest(pos types.Position) (types.StageResultItem, bool)
}

// resultInputDelayTicks — пауза перед приёмом ввода на экране итогов,
// чтобы очередь выстрела не пролистала результат
const resultInputDelayTicks = 45

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
	LightingUseCases      interfaces.ILightingUseCases
	VisionUseCases        interfaces.IVisionUseCases
	VisualEffectsUseCases interfaces.IVisualEffectsUseCases
	ProgressionUseCases   interfaces.IProgressionUseCases

	// Adapters
	InputAdapters      [2]interfaces.IInputAdapter
	EnemyInputAdapter  interfaces.IAiInputAdapter
	Renderer           StageRenderer
	SoundPlayerAdapter interfaces.ISoundPlayerAdapter
	TouchControls      interfaces.ITouchControlsAdapter

	// Session & Repositories
	StageSession      *session_entities.StageSessionEntity
	BonusesRepository interfaces.IBonusesRepository

	// Включённость графических эффектов, общая для приложения
	EffectsSettings *types.EffectsSettingsEntity
	// Level — сценарий уровня для подсчёта звёзд
	Level *types.LevelEntity
}

type StageState struct {
	// Use Cases
	tankCommonUseCases    interfaces.ITankCommonUseCases
	renderUseCases        interfaces.IRenderUseCases
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases
	tilesUseCases         interfaces.ITilesUseCases
	stageUseCases         interfaces.IStageUseCases
	soundUseCases         interfaces.ISoundUseCases
	lightingUseCases      interfaces.ILightingUseCases
	visionUseCases        interfaces.IVisionUseCases
	visualEffectsUseCases interfaces.IVisualEffectsUseCases
	progressionUseCases   interfaces.IProgressionUseCases

	// Adapters
	inputAdapters      [2]interfaces.IInputAdapter
	enemyInputAdapter  interfaces.IAiInputAdapter
	renderer           StageRenderer
	soundPlayerAdapter interfaces.ISoundPlayerAdapter
	touchControls      interfaces.ITouchControlsAdapter

	// Session & Repositories
	stageSession      *session_entities.StageSessionEntity
	bonusesRepository interfaces.IBonusesRepository

	// Entities
	effectsSettings *types.EffectsSettingsEntity

	isSetUp         bool
	endSoundHandled bool
	debugEnabled    bool // Флаг дебаг-режима

	// Меню паузы — чистая презентация поверх доменного флага паузы:
	// видимо, пока уровень на паузе и не завершён
	pauseMenuItems   []types.PauseMenuItem
	pauseMenuIndex   int
	pauseMenuWasOpen bool

	// Экран итогов: считается один раз при завершении уровня
	level            *types.LevelEntity
	result           *types.StageResultViewData
	resultInputDelay uint
	nextLevel        int
}

func NewStageState(deps StageStateDependencies) *StageState {
	return &StageState{
		tankCommonUseCases:    deps.TankCommonUseCases,
		renderUseCases:        deps.RenderUseCases,
		tankLifecycleUseCases: deps.TankLifecycleUseCases,
		tilesUseCases:         deps.TilesUseCases,
		stageUseCases:         deps.StageUseCases,
		soundUseCases:         deps.SoundUseCases,
		lightingUseCases:      deps.LightingUseCases,
		visionUseCases:        deps.VisionUseCases,
		visualEffectsUseCases: deps.VisualEffectsUseCases,
		progressionUseCases:   deps.ProgressionUseCases,
		inputAdapters:         deps.InputAdapters,
		enemyInputAdapter:     deps.EnemyInputAdapter,
		renderer:              deps.Renderer,
		soundPlayerAdapter:    deps.SoundPlayerAdapter,
		touchControls:         deps.TouchControls,
		stageSession:          deps.StageSession,
		bonusesRepository:     deps.BonusesRepository,
		effectsSettings:       deps.EffectsSettings,
		level:                 deps.Level,
		pauseMenuItems: []types.PauseMenuItem{
			types.PauseMenuItemContinue,
			types.PauseMenuItemGraphics,
			types.PauseMenuItemExitToLevels,
		},
	}
}

// SetDebugEnabled устанавливает флаг дебаг-режима
func (state *StageState) SetDebugEnabled(enabled bool) {
	state.debugEnabled = enabled
}

func (state *StageState) SetUp() {
	// Сбрасываем флаг звуков завершения уровня для нового уровня
	state.endSoundHandled = false
	state.result = nil

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
		playerTank := state.stageUseCases.PlacePlayerTank(role)

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
		transition = state.handleStageResult()
	}

	// Меню паузы работает только пока уровень не завершён:
	// на экране итогов действует своё меню
	if !stageFinished {
		transition = state.handlePauseMenu()
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

		state.lightingUseCases.UpdateHeadlights()
		state.visionUseCases.UpdateVisibility()
	}

	// Эффекты продвигаются и на финальном оверлее: дым и тряска
	// от взрыва штаба доигрывают, а не замирают; в меню паузы стоят
	if !paused || stageFinished {
		state.visualEffectsUseCases.Update()
	}

	// Единственная точка контакта с звуковым адаптером: применяем
	// накопленные события кадра в порядке добавления
	for _, event := range state.soundUseCases.GetEvents() {
		state.applySoundEvent(event)
	}
	state.soundPlayerAdapter.Update()

	return transition
}

// handlePauseMenu — переключение паузы по ESC и управление меню:
// клавиатурная навигация, выбор Enter/Space и тап по пункту
func (state *StageState) handlePauseMenu() types.StateTransition {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		state.stageUseCases.TogglePause()
	}

	menuOpen := state.stageUseCases.IsPaused()
	// При каждом открытии меню курсор возвращается на первый пункт
	if menuOpen && !state.pauseMenuWasOpen {
		state.pauseMenuIndex = 0
	}
	state.pauseMenuWasOpen = menuOpen
	if !menuOpen {
		return types.StateTransition{}
	}

	state.handlePauseMenuNavigation()

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return state.applyPauseMenuSelection(
			state.pauseMenuItems[state.pauseMenuIndex],
		)
	}

	// Тап сразу активирует пункт: это командное меню без степперов
	if pos, ok := state.touchControls.TapJustPressed(); ok {
		if item, hit := state.renderer.PauseMenuHitTest(pos); hit {
			return state.applyPauseMenuSelection(item)
		}
	}

	return types.StateTransition{}
}

func (state *StageState) handlePauseMenuNavigation() {
	moveUp := inpututil.IsKeyJustPressed(ebiten.KeyUp) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW)
	moveDown := inpututil.IsKeyJustPressed(ebiten.KeyDown) ||
		inpututil.IsKeyJustPressed(ebiten.KeyS)

	if moveUp && state.pauseMenuIndex > 0 {
		state.pauseMenuIndex--
	}
	if moveDown && state.pauseMenuIndex < len(state.pauseMenuItems)-1 {
		state.pauseMenuIndex++
	}
}

// applyPauseMenuSelection применяет выбранный пункт меню паузы
func (state *StageState) applyPauseMenuSelection(
	item types.PauseMenuItem,
) types.StateTransition {
	switch item {
	case types.PauseMenuItemContinue:
		state.stageUseCases.ResumeStageState()
	case types.PauseMenuItemGraphics:
		// Меню остаётся открытым: смена видна сразу за оверлеем
		state.effectsSettings.Toggle()
	case types.PauseMenuItemExitToLevels:
		// Глушим звуки уровня при выходе на экран выбора
		state.soundUseCases.RequestStopAll()

		return state.levelsTransition()
	}

	return types.StateTransition{}
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
		if tank.IsEnemy() && tank.GetWithBonus() {
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

	if state.result != nil {
		state.renderer.DrawStageResult(screen, *state.result)
		return
	}

	if state.stageUseCases.IsPaused() {
		state.renderer.DrawPauseMenu(screen, types.PauseMenuViewData{
			Items:          state.pauseMenuItems,
			ActiveIndex:    state.pauseMenuIndex,
			EffectsEnabled: state.effectsSettings.IsEnabled(),
		})
	}
}

// handleStageResult считает итог уровня при первом кадре после
// завершения и обрабатывает меню итогов
func (state *StageState) handleStageResult() types.StateTransition {
	if state.result == nil {
		state.buildStageResult()
		return types.StateTransition{}
	}
	if state.resultInputDelay > 0 {
		state.resultInputDelay--
		return types.StateTransition{}
	}

	result := state.result
	moveUp := inpututil.IsKeyJustPressed(ebiten.KeyUp) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW)
	moveDown := inpututil.IsKeyJustPressed(ebiten.KeyDown) ||
		inpututil.IsKeyJustPressed(ebiten.KeyS)
	if moveUp && result.ActiveIndex > 0 {
		result.ActiveIndex--
	}
	if moveDown && result.ActiveIndex < len(result.Items)-1 {
		result.ActiveIndex++
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return state.applyStageResultItem(result.Items[result.ActiveIndex])
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return state.applyStageResultItem(types.StageResultItemLevels)
	}

	if pos, ok := state.touchControls.TapJustPressed(); ok {
		if item, hit := state.renderer.StageResultHitTest(pos); hit {
			return state.applyStageResultItem(item)
		}
	}

	return types.StateTransition{}
}

// buildStageResult считает звёзды, сохраняет лучший результат
// и собирает пункты меню итогов
func (state *StageState) buildStageResult() {
	result := state.stageUseCases.GetStageResult()
	stars := state.progressionUseCases.CalcStars(result, state.level)
	levelNumber := int(state.stageSession.GetStageNumber())
	newBest := result.Won &&
		stars > state.progressionUseCases.GetLevelStars(levelNumber)

	if result.Won {
		// Снимок жизней и прокачки — для пункта Continue
		state.stageUseCases.SaveCarryOver()
		if err := state.progressionUseCases.RecordResult(
			levelNumber, stars,
		); err != nil {
			log.Printf("save progress: %v", err)
		}
	}

	items := []types.StageResultItem{
		types.StageResultItemRetry,
		types.StageResultItemLevels,
	}
	if next, unlocked := state.progressionUseCases.NextLevel(
		levelNumber,
	); result.Won && unlocked {
		state.nextLevel = next
		head := []types.StageResultItem{types.StageResultItemNext}
		if state.stageSession.HasCarryOverAdvantage() {
			head = append(head, types.StageResultItemContinue)
		}
		items = append(head, items...)
	}

	state.result = &types.StageResultViewData{
		Won:          result.Won,
		Stars:        stars,
		ElapsedTicks: result.ElapsedTicks,
		LivesLost:    result.LivesLost,
		CarriedOver:  result.CarriedOver,
		NewBest:      newBest,
		Items:        items,
	}
	state.resultInputDelay = resultInputDelayTicks
}

func (state *StageState) applyStageResultItem(
	item types.StageResultItem,
) types.StateTransition {
	// Глушим звук завершения при уходе с экрана итогов
	state.soundUseCases.RequestStopAll()

	switch item {
	case types.StageResultItemNext:
		return types.StateTransition{
			Target: types.TransitionToStage,
			Level:  uint(state.nextLevel),
		}
	case types.StageResultItemContinue:
		return types.StateTransition{
			Target:    types.TransitionToStage,
			Level:     uint(state.nextLevel),
			CarryOver: true,
		}
	case types.StageResultItemRetry:
		return types.StateTransition{
			Target: types.TransitionToStage,
			Level:  state.stageSession.GetStageNumber(),
		}
	default:
		return state.levelsTransition()
	}
}

// levelsTransition — выход на экран выбора с курсором на этом уровне
func (state *StageState) levelsTransition() types.StateTransition {
	return types.StateTransition{
		Target: types.TransitionToLevelSelect,
		Level:  state.stageSession.GetStageNumber(),
	}
}
