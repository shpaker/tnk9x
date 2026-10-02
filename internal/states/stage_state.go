package states

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

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
	DrawStageResult(screen *ebiten.Image, view types.StageResultViewData)
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
	LightingUseCases      interfaces.ILightingUseCases
	VisualEffectsUseCases interfaces.IVisualEffectsUseCases
	ProgressionUseCases   interfaces.IProgressionUseCases

	// Adapters
	InputAdapters      [2]interfaces.IInputAdapter
	EnemyInputAdapter  interfaces.IAiInputAdapter
	Renderer           StageRenderer
	SoundPlayerAdapter interfaces.ISoundPlayerAdapter
	MenuInput          interfaces.IMenuInputAdapter

	// Session & Repositories
	StageSession      *session_entities.StageSessionEntity
	BonusesRepository interfaces.IBonusesRepository

	// Экран настроек из меню паузы, общий для приложения
	SettingsOverlay *SettingsOverlay
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
	visualEffectsUseCases interfaces.IVisualEffectsUseCases
	progressionUseCases   interfaces.IProgressionUseCases

	// Adapters
	inputAdapters      [2]interfaces.IInputAdapter
	enemyInputAdapter  interfaces.IAiInputAdapter
	renderer           StageRenderer
	soundPlayerAdapter interfaces.ISoundPlayerAdapter
	menuInput          interfaces.IMenuInputAdapter

	// Session & Repositories
	stageSession      *session_entities.StageSessionEntity
	bonusesRepository interfaces.IBonusesRepository

	// Presentation
	settingsOverlay *SettingsOverlay

	isSetUp         bool
	endSoundHandled bool

	// Меню паузы — чистая презентация поверх доменного флага паузы:
	// видимо, пока уровень на паузе и не завершён
	pauseMenuItems   []types.PauseMenuItem
	pauseMenuIndex   int
	pauseMenuWasOpen bool

	// Экран итогов: считается один раз при завершении уровня
	level  *types.LevelEntity
	result *types.StageResultViewData
	// resultTicks — тиков с начала появления экрана итогов
	resultTicks uint
	nextLevel   int
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
		visualEffectsUseCases: deps.VisualEffectsUseCases,
		progressionUseCases:   deps.ProgressionUseCases,
		inputAdapters:         deps.InputAdapters,
		enemyInputAdapter:     deps.EnemyInputAdapter,
		renderer:              deps.Renderer,
		soundPlayerAdapter:    deps.SoundPlayerAdapter,
		menuInput:             deps.MenuInput,
		stageSession:          deps.StageSession,
		bonusesRepository:     deps.BonusesRepository,
		settingsOverlay:       deps.SettingsOverlay,
		level:                 deps.Level,
		pauseMenuItems: []types.PauseMenuItem{
			types.PauseMenuItemContinue,
			types.PauseMenuItemRestart,
			types.PauseMenuItemSettings,
			types.PauseMenuItemExitToLevels,
		},
	}
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

	// Пока открыты настройки, ввод идёт только в них: иначе Esc,
	// P или тач-пауза сняли бы паузу под оверлеем
	settingsOpen := state.settingsOverlay.IsOpen()
	if settingsOpen {
		state.settingsOverlay.Update()
	}

	for _, adapter := range state.inputAdapters {
		if adapter != nil && !settingsOpen {
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
	if !stageFinished && !settingsOpen {
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
	}

	// Эффекты продвигаются и на финальном оверлее: дым и тряска
	// от взрыва штаба доигрывают, а не замирают; в меню паузы стоят
	if !paused || stageFinished {
		state.visualEffectsUseCases.Update()
	}

	playSoundEvents(state.soundUseCases, state.soundPlayerAdapter)

	return transition
}

// handlePauseMenu — единственная точка переключения паузы: Esc,
// Start или тач-пауза; в открытом меню «назад» тоже возвращает
// в игру. Меню — стрелки, крестовина или стик, выбор Enter, A
// или огнём
func (state *StageState) handlePauseMenu() types.StateTransition {
	if state.menuInput.PauseJustPressed() ||
		(state.stageUseCases.IsPaused() && state.menuInput.Back()) {
		state.stageUseCases.TogglePause()
	}

	menuOpen := state.stageUseCases.IsPaused()
	// При каждом открытии меню курсор возвращается на первый пункт,
	// а луп двигателя глушится: на паузе управление звуком двигателя
	// в Update не выполняется. После CONTINUE двигатель запросится
	// снова, если игрок едет
	if menuOpen && !state.pauseMenuWasOpen {
		state.pauseMenuIndex = 0
		state.soundUseCases.RequestStop(types.SoundIDEngine)
	}
	state.pauseMenuWasOpen = menuOpen
	if !menuOpen {
		return types.StateTransition{}
	}

	moveUp, moveDown := state.menuInput.Steps()
	state.pauseMenuIndex = stepIndex(
		state.pauseMenuIndex, len(state.pauseMenuItems), moveUp, moveDown,
	)

	if state.menuInput.Confirmed() {
		return state.applyPauseMenuSelection(
			state.pauseMenuItems[state.pauseMenuIndex],
		)
	}

	return types.StateTransition{}
}

// applyPauseMenuSelection применяет выбранный пункт меню паузы
func (state *StageState) applyPauseMenuSelection(
	item types.PauseMenuItem,
) types.StateTransition {
	switch item {
	case types.PauseMenuItemContinue:
		state.stageUseCases.ResumeStageState()
	case types.PauseMenuItemRestart:
		// Глушим звуки уровня перед перезапуском
		state.soundUseCases.RequestStopAll()

		return state.restartTransition()
	case types.PauseMenuItemSettings:
		// Уровень остаётся на паузе; из настроек — обратно в меню
		state.settingsOverlay.Open()
	case types.PauseMenuItemExitToLevels:
		// Глушим звуки уровня при выходе на экран выбора
		state.soundUseCases.RequestStopAll()

		return state.levelsTransition()
	}

	return types.StateTransition{}
}

// updateBlinkObjects обновляет мигание бонусов, танков с бонусом
// и тяжёлых танков
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
		if state.renderUseCases.IsTankBlinking(tank) {
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

	if state.settingsOverlay.IsOpen() {
		state.settingsOverlay.Draw(screen)
		return
	}

	if state.stageUseCases.IsPaused() {
		state.renderer.DrawPauseMenu(screen, types.PauseMenuViewData{
			Items:       state.pauseMenuItems,
			ActiveIndex: state.pauseMenuIndex,
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
	if state.revealResult() {
		return types.StateTransition{}
	}

	result := state.result
	moveUp, moveDown := state.menuInput.Steps()
	result.ActiveIndex = stepIndex(
		result.ActiveIndex, len(result.Items), moveUp, moveDown,
	)

	if state.menuInput.Confirmed() {
		return state.applyStageResultItem(result.Items[result.ActiveIndex])
	}
	// ESC и кнопка паузы — выход к выбору уровней
	if state.menuInput.Back() {
		return state.applyStageResultItem(types.StageResultItemLevels)
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
	state.resultTicks = 0
}

// revealStars — сколько звёзд загорается на экране итогов:
// при поражении звёзд нет
func (state *StageState) revealStars() uint {
	if !state.result.Won {
		return 0
	}
	return state.result.Stars
}

// revealResult продвигает поэтапное появление экрана итогов; каждая
// загоревшаяся звезда звучит. Подтверждение или «назад» сразу
// показывают экран целиком, ничего не выбирая: пункты принимаются
// только после появления меню, чтобы очередь выстрелов не пролистала
// итоги. Возвращает true, пока экран ещё появляется
func (state *StageState) revealResult() bool {
	stars := state.revealStars()
	length := resultRevealLength(stars)
	if state.resultTicks >= length {
		return false
	}

	litBefore := resultReveal(state.resultTicks, stars).Stars
	state.resultTicks++
	skip := state.menuInput.Confirmed() || state.menuInput.Back()
	if skip {
		state.resultTicks = length
	}

	reveal := resultReveal(state.resultTicks, stars)
	if reveal.Stars > litBefore && !skip {
		state.soundUseCases.RequestSound(types.SoundIDBonus, false)
	}
	state.result.Reveal = reveal
	return true
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
		return state.restartTransition()
	default:
		return state.levelsTransition()
	}
}

// restartTransition — этот же уровень заново, без переноса жизней
// и прокачки
func (state *StageState) restartTransition() types.StateTransition {
	return types.StateTransition{
		Target: types.TransitionToStage,
		Level:  state.stageSession.GetStageNumber(),
	}
}

// levelsTransition — выход на экран выбора с курсором на этом уровне
func (state *StageState) levelsTransition() types.StateTransition {
	return types.StateTransition{
		Target: types.TransitionToLevelSelect,
		Level:  state.stageSession.GetStageNumber(),
	}
}
