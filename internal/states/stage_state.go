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
	// HitPauseRow и HitResultRow — пункт меню паузы или итогов
	// последней отрисовки под точкой
	HitPauseRow(position types.Position) (int, bool)
	HitResultRow(position types.Position) (int, bool)
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
	// InventoryUseCases — жетоны вместо рекламы для пунктов итогов
	InventoryUseCases interfaces.IInventoryUseCases

	// Adapters
	InputAdapters      [2]interfaces.IInputAdapter
	EnemyInputAdapter  interfaces.IAiInputAdapter
	Renderer           StageRenderer
	SoundPlayerAdapter interfaces.ISoundPlayerAdapter
	MenuInput          interfaces.IMenuInputAdapter
	// RewardAdapter — реклама за вознаграждение для пунктов итогов
	RewardAdapter interfaces.IRewardAdapter

	// Session & Repositories
	StageSession      *session_entities.StageSessionEntity
	BonusesRepository interfaces.IBonusesRepository

	// Экран настроек из меню паузы, общий для приложения
	SettingsOverlay *SettingsOverlay
	// Страница «Как играть» перед первым запуском уровня 1
	HelpOverlay *HelpOverlay
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
	inventoryUseCases     interfaces.IInventoryUseCases

	// Adapters
	inputAdapters      [2]interfaces.IInputAdapter
	enemyInputAdapter  interfaces.IAiInputAdapter
	renderer           StageRenderer
	soundPlayerAdapter interfaces.ISoundPlayerAdapter
	menuInput          interfaces.IMenuInputAdapter
	rewardAdapter      interfaces.IRewardAdapter

	// Session & Repositories
	stageSession      *session_entities.StageSessionEntity
	bonusesRepository interfaces.IBonusesRepository

	// Presentation
	settingsOverlay *SettingsOverlay
	helpOverlay     *HelpOverlay

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
	// Пункт итогов, за который показывается реклама; пока исход
	// не получен, меню итогов ждёт
	rewardItem    types.StageResultItem
	rewardPending bool
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
		inventoryUseCases:     deps.InventoryUseCases,
		inputAdapters:         deps.InputAdapters,
		enemyInputAdapter:     deps.EnemyInputAdapter,
		renderer:              deps.Renderer,
		soundPlayerAdapter:    deps.SoundPlayerAdapter,
		menuInput:             deps.MenuInput,
		rewardAdapter:         deps.RewardAdapter,
		stageSession:          deps.StageSession,
		bonusesRepository:     deps.BonusesRepository,
		settingsOverlay:       deps.SettingsOverlay,
		helpOverlay:           deps.HelpOverlay,
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
	// Пока открыта страница «Как играть», бой не начинается:
	// ни стартового звука, ни танков, ни отсчёта времени
	if state.isHelpOpen() {
		state.helpOverlay.Update(state.stageSession.GetPlayerCount())
		return types.StateTransition{}
	}
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

	// Завершение обрабатывается в том же кадре, где уровень закончился:
	// к Draw итоги уже построены, и меню паузы не мелькает перед ними
	if state.stageUseCases.IsStageFinished() {
		transition = state.handleStageFinish()
	}

	// Эффекты продвигаются и на финальном оверлее: дым и тряска
	// от взрыва штаба доигрывают, а не замирают; в меню паузы стоят
	if !paused || stageFinished {
		state.visualEffectsUseCases.Update()
	}

	playSoundEvents(state.soundUseCases, state.soundPlayerAdapter)

	return transition
}

// handleStageFinish ставит завершённый уровень на паузу, один раз
// глушит все звуки (в т.ч. луп двигателя), при поражении проигрывает
// gameover и ведёт экран итогов
func (state *StageState) handleStageFinish() types.StateTransition {
	if !state.stageUseCases.IsPaused() {
		state.stageUseCases.PauseStageState()
	}
	if !state.endSoundHandled {
		state.soundUseCases.RequestStopAll()
		if !state.stageUseCases.IsStageWon() {
			state.soundUseCases.RequestSound(types.SoundIDGameOver, false)
		}
		state.endSoundHandled = true
	}
	return state.handleStageResult()
}

// Suspend — площадка приостановила игру (реклама, скрытая вкладка):
// идущий бой уходит на паузу, чтобы после возврата игрок продолжил
// из меню паузы, а не оказался сразу в бою
func (state *StageState) Suspend() {
	if !state.isSetUp ||
		state.stageUseCases.IsStageFinished() ||
		state.stageUseCases.IsPaused() {
		return
	}
	state.stageUseCases.PauseStageState()
}

// IsGameplayActive — идёт бой: уровень запущен и не на паузе;
// завершённый уровень всегда на паузе
func (state *StageState) IsGameplayActive() bool {
	return state.isSetUp && !state.stageUseCases.IsPaused()
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
	pauseMenuIndex, clicked := pointerIndex(
		state.menuInput, state.renderer.HitPauseRow, state.pauseMenuIndex,
	)
	state.pauseMenuIndex = pauseMenuIndex

	if state.menuInput.Confirmed() || clicked {
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

	if state.isHelpOpen() {
		state.helpOverlay.Draw(screen)
		return
	}

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

// isHelpOpen — уровень ещё не начат и перед ним показывается
// страница «Как играть»
func (state *StageState) isHelpOpen() bool {
	return !state.isSetUp &&
		state.helpOverlay.IsDue(state.stageSession.GetStageNumber())
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
	if state.rewardPending {
		return state.pollReward()
	}

	result := state.result
	moveUp, moveDown := state.menuInput.Steps()
	result.ActiveIndex = stepIndex(
		result.ActiveIndex, len(result.Items), moveUp, moveDown,
	)
	activeIndex, clicked := pointerIndex(
		state.menuInput, state.renderer.HitResultRow, result.ActiveIndex,
	)
	result.ActiveIndex = activeIndex

	if state.menuInput.Confirmed() || clicked {
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

	next, unlocked := state.progressionUseCases.NextLevel(levelNumber)
	if result.Won && unlocked {
		state.nextLevel = next
	}
	// Пункты за рекламу оплачиваются и жетонами: с жетонами они есть
	// и там, где площадка рекламу не умеет
	useTokens := state.inventoryUseCases.GetTokens() > 0
	rewardAvailable := useTokens || state.rewardAdapter.IsRewardAvailable()
	items := stageResultMenu{
		won:                result.Won,
		nextUnlocked:       unlocked,
		carryOverAdvantage: state.stageSession.HasCarryOverAdvantage(),
		rewardAvailable:    rewardAvailable,
		canRevive: rewardAvailable &&
			state.stageUseCases.CanRevivePlayers(),
	}.items()

	state.result = &types.StageResultViewData{
		Won:          result.Won,
		Stars:        stars,
		ElapsedTicks: result.ElapsedTicks,
		LivesLost:    result.LivesLost,
		CarriedOver:  result.CarriedOver,
		NewBest:      newBest,
		UseTokens:    useTokens,
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
	skip := state.menuInput.Confirmed() || state.menuInput.Back() ||
		tappedAnywhere(state.menuInput)
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
	// Пункт за рекламу: жетон оплачивает его сразу, иначе экран
	// итогов ждёт исхода рекламы
	if item.IsRewarded() {
		if state.spendToken() {
			return state.grantReward(item)
		}
		state.requestReward(item)
		return types.StateTransition{}
	}

	// Глушим звук завершения при уходе с экрана итогов
	state.soundUseCases.RequestStopAll()

	switch item {
	case types.StageResultItemNext:
		return types.StateTransition{
			Target:       types.TransitionToStage,
			Level:        uint(state.nextLevel),
			Intermission: true,
		}
	case types.StageResultItemContinue:
		return types.StateTransition{
			Target:       types.TransitionToStage,
			Level:        uint(state.nextLevel),
			CarryOver:    true,
			Intermission: true,
		}
	case types.StageResultItemRetry:
		return state.restartTransition()
	default:
		return state.levelsTransition()
	}
}

// restartTransition — этот же уровень заново, без переноса жизней
// и прокачки; уход с уровня — логическая пауза
func (state *StageState) restartTransition() types.StateTransition {
	return types.StateTransition{
		Target:       types.TransitionToStage,
		Level:        state.stageSession.GetStageNumber(),
		Intermission: true,
	}
}

// levelsTransition — выход на экран выбора с курсором на этом уровне;
// уход с уровня — логическая пауза
func (state *StageState) levelsTransition() types.StateTransition {
	return types.StateTransition{
		Target:       types.TransitionToLevelSelect,
		Level:        state.stageSession.GetStageNumber(),
		Intermission: true,
	}
}
