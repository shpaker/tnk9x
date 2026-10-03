package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/adapters/platform"
	"github.com/shpaker/tnk9x/internal/states"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

// Тесты запускаются из каталога пакета, а config.yml и assets
// лежат в корне репозитория — переходим туда до запуска
func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

type stubGameState struct {
	updates int
}

func (s *stubGameState) Update() types.StateTransition {
	s.updates++
	return types.StateTransition{}
}

func (s *stubGameState) Draw(screen *ebiten.Image) {}

// stubStageLikeState — состояние с геймплеем, реагирующее
// на приостановку площадкой
type stubStageLikeState struct {
	stubGameState
	suspends int
	active   bool
}

func (s *stubStageLikeState) Suspend() { s.suspends++ }

func (s *stubStageLikeState) IsGameplayActive() bool { return s.active }

var errLevelUnavailable = errors.New("level unavailable")

// failingMapsRepo валит сборку уровня на первом же шаге newStageState
type failingMapsRepo struct{}

func (r *failingMapsRepo) GetLevel(
	num int,
	tileBaseSize int,
) (*types.LevelEntity, error) {
	return nil, errLevelUnavailable
}

func (r *failingMapsRepo) HasLevel(num int) bool { return false }

func (r *failingMapsRepo) GetSceneLevel(
	string, int,
) (*types.LevelEntity, error) {
	return nil, errLevelUnavailable
}

func (r *failingMapsRepo) GetLevelsCount() (int, error) { return 0, nil }

// newAppTestEnv собирает минимальный App без тяжёлой инфраструктуры:
// приватные поля доступны, так как тест живёт в пакете app.
// Площадка — настоящая десктопная (тесты собираются без тега js)
func newAppTestEnv() (*App, *stubGameState) {
	state := &stubGameState{}
	settings := types.NewSettingsEntity()
	settings.SetPlayers(2)
	platformAdapter := platform.NewPlatformAdapter()
	app := &App{
		config: &Config{
			TileBaseSize: 8,
			BaseSizePx:   16,
		},
		settings:        settings,
		state:           state,
		session:         session_entities.NewGameSessionEntity(),
		mapsRepository:  &failingMapsRepo{},
		platformAdapter: platformAdapter,
		rewardAdapter:   platformAdapter,
		purchaseAdapter: platformAdapter,
		windowAdapter:   &testutil.FakeWindow{},

		inventoryUseCases: &testutil.FakeInventory{},
	}
	return app, state
}

// Полный App через New() собирается один раз на процесс:
// audio.NewContext допускает только один вызов
var (
	fullAppOnce sync.Once
	fullApp     *App
	fullAppErr  error
)

func newFullApp(t *testing.T) *App {
	t.Helper()
	fullAppOnce.Do(func() {
		cfg, err := LoadConfig()
		if err != nil {
			fullAppErr = err
			return
		}
		fullApp = New(cfg)
	})
	if fullAppErr != nil {
		t.Fatalf("конфигурация не загружена: %v", fullAppErr)
	}
	return fullApp
}

func TestApp_ApplyTransition_None(t *testing.T) {
	app, state := newAppTestEnv()

	if err := app.applyTransition(types.StateTransition{}); err != nil {
		t.Fatalf("пустой переход: %v", err)
	}
	if app.state != state {
		t.Error("состояние заменено при пустом переходе")
	}
}

// Параметры уровня записываются в сессию до сборки состояния:
// даже при ошибке сборки сессия уже обновлена
func TestApp_ApplyTransition_ToStage_SessionWrittenBeforeBuild(
	t *testing.T,
) {
	app, state := newAppTestEnv()

	err := app.applyTransition(types.StateTransition{
		Target: types.TransitionToStage,
		Level:  7,
	})
	if !errors.Is(err, errLevelUnavailable) {
		t.Fatalf("ошибка %v, ожидалась errLevelUnavailable", err)
	}

	if app.session.Level != 7 {
		t.Errorf("уровень сессии %d, ожидался 7", app.session.Level)
	}
	stageSession := app.session.StageSession()
	if got := stageSession.GetPlayerCount(); got != 2 {
		t.Errorf("players %d, want 2 from settings", got)
	}
	if got := stageSession.GetStageNumber(); got != 7 {
		t.Errorf("stage number %d, want 7", got)
	}

	// Состояние при ошибке сборки не меняется
	if app.state != state {
		t.Error("состояние заменено при ошибке сборки уровня")
	}
}

// Переход Quit завершает игровой цикл штатно, не трогая состояние
func TestApp_ApplyTransition_Quit(t *testing.T) {
	app, state := newAppTestEnv()

	err := app.applyTransition(types.StateTransition{
		Target: types.TransitionToQuit,
	})
	if !errors.Is(err, ebiten.Termination) {
		t.Fatalf("ошибка %v, ожидалась ebiten.Termination", err)
	}
	if app.state != state {
		t.Error("состояние заменено при переходе Quit")
	}
}

// Полный цикл на настоящем графе зависимостей:
// выбор уровня -> уровень -> выбор уровня
func TestApp_ApplyTransition_FullApp(t *testing.T) {
	app := newFullApp(t)

	if _, ok := app.state.(*states.SplashState); !ok {
		t.Fatalf("initial state %T, want SplashState", app.state)
	}

	// Готовность площадке — с появлением главного меню
	fakePlatform := &testutil.FakePlatform{}
	defaultPlatform := app.platformAdapter
	app.platformAdapter = fakePlatform
	defer func() { app.platformAdapter = defaultPlatform }()

	// Сплеш догружает граф, затем главное меню и выбор уровня
	for _, done := app.loader.Step(); !done; _, done = app.loader.Step() {
	}
	if len(app.frameInputs) != 3 {
		t.Errorf(
			"после загрузки к вводу добавляются хоткеи: %d",
			len(app.frameInputs),
		)
	}
	for _, step := range []struct {
		target types.TransitionTarget
		want   any
	}{
		{types.TransitionToMainMenu, &states.MainMenuState{}},
		{types.TransitionToLevelSelect, &states.LevelSelectState{}},
	} {
		if err := app.applyTransition(types.StateTransition{Target: step.target}); err != nil {
			t.Fatalf("переход %v: %v", step.target, err)
		}
		if fmt.Sprintf("%T", app.state) != fmt.Sprintf("%T", step.want) {
			t.Fatalf("состояние %T, ожидалось %T", app.state, step.want)
		}
	}

	if fakePlatform.ReadyCalls != 1 {
		t.Errorf("Ready calls %d, want 1", fakePlatform.ReadyCalls)
	}

	// У одиночной игры и игры вдвоём раздельное прохождение
	if app.progressionUseCases[0] == app.progressionUseCases[1] ||
		app.levelSelectUseCases[0] == app.levelSelectUseCases[1] {
		t.Error("прогресс режимов должен быть раздельным")
	}

	err := app.applyTransition(types.StateTransition{
		Target: types.TransitionToStage,
		Level:  2,
	})
	if err != nil {
		t.Fatalf("переход на уровень: %v", err)
	}

	if _, ok := app.state.(*states.StageState); !ok {
		t.Fatalf("состояние %T, ожидалось StageState", app.state)
	}
	if app.session.Level != 2 {
		t.Errorf("уровень сессии %d, ожидался 2", app.session.Level)
	}

	if got := app.session.StageSession().GetTotalEnemies(); got == 0 {
		t.Error("the level waves did not reach the stage session")
	}

	err = app.applyTransition(types.StateTransition{
		Target: types.TransitionToLevelSelect,
		Level:  2,
	})
	if err != nil {
		t.Fatalf("back to stage select: %v", err)
	}
	if _, ok := app.state.(*states.LevelSelectState); !ok {
		t.Fatalf("state %T, want LevelSelectState", app.state)
	}
}

// Отмена контекста завершает игровой цикл штатно (ebiten.Termination)
func TestApp_Update_ContextCancelled(t *testing.T) {
	app, _ := newAppTestEnv()

	ctx, cancel := context.WithCancel(context.Background())
	app.ctx = ctx

	cancel()
	if err := app.Update(); !errors.Is(err, ebiten.Termination) {
		t.Fatalf("ожидался ebiten.Termination, получено %v", err)
	}
}

// Десктоп: площадка не приостанавливает игру, состояние обновляется
// каждый кадр
func TestApp_Update_StateEveryFrame(t *testing.T) {
	app, state := newAppTestEnv()

	for range 3 {
		if err := app.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	if state.updates != 3 {
		t.Errorf("state updates %d, want 3", state.updates)
	}
}

// Пока площадка приостановила игру, кадры пропускаются: состояние
// не обновляется, геймплей не размечается, а начало приостановки
// отдаётся состоянию один раз
func TestApp_Update_SuspendedSkipsFrames(t *testing.T) {
	app, _ := newAppTestEnv()
	state := &stubStageLikeState{active: true}
	app.state = state
	fakePlatform := &testutil.FakePlatform{
		Suspended:     true,
		JustSuspended: true,
	}
	app.platformAdapter = fakePlatform

	_ = app.Update()
	fakePlatform.JustSuspended = false
	_ = app.Update()

	if state.updates != 0 {
		t.Errorf("state updated %d times while suspended", state.updates)
	}
	if state.suspends != 1 {
		t.Errorf("Suspend called %d times, want 1", state.suspends)
	}
	if active, ok := fakePlatform.LastGameplay(); !ok || active {
		t.Error("gameplay must be reported inactive while suspended")
	}

	fakePlatform.Suspended = false
	_ = app.Update()
	if state.updates != 1 {
		t.Errorf("state updates %d after resume, want 1", state.updates)
	}
	if active, _ := fakePlatform.LastGameplay(); !active {
		t.Error("gameplay of the state must be reported after resume")
	}
}

// Приостановка, начавшаяся и закончившаяся между кадрами (скрытая
// вкладка): состояние уходит на паузу, кадр идёт как обычно
func TestApp_Update_MissedSuspension(t *testing.T) {
	app, _ := newAppTestEnv()
	state := &stubStageLikeState{}
	app.state = state
	app.platformAdapter = &testutil.FakePlatform{JustSuspended: true}

	_ = app.Update()

	if state.suspends != 1 || state.updates != 1 {
		t.Errorf("suspends %d, updates %d, want 1 and 1",
			state.suspends, state.updates)
	}
}

// Переход через логическую паузу просит площадку о рекламе
// ещё до сборки нового состояния
func TestApp_ApplyTransition_Intermission(t *testing.T) {
	app, _ := newAppTestEnv()
	fakePlatform := &testutil.FakePlatform{}
	app.platformAdapter = fakePlatform

	_ = app.applyTransition(types.StateTransition{
		Target:       types.TransitionToStage,
		Level:        1,
		Intermission: true,
	})

	if fakePlatform.IntermissionCalls != 1 {
		t.Errorf("intermissions %d, want 1", fakePlatform.IntermissionCalls)
	}
	if active, ok := fakePlatform.LastGameplay(); !ok || active {
		t.Error("gameplay must stop before the intermission")
	}
}

// Купленное «без рекламы»: логическая пауза есть, рекламы нет
func TestApp_ApplyTransition_NoAdsPurchased(t *testing.T) {
	app, _ := newAppTestEnv()
	fakePlatform := &testutil.FakePlatform{}
	app.platformAdapter = fakePlatform
	app.inventoryUseCases = &testutil.FakeInventory{NoAds: true}

	_ = app.applyTransition(types.StateTransition{
		Target:       types.TransitionToStage,
		Level:        1,
		Intermission: true,
	})

	if fakePlatform.IntermissionCalls != 0 {
		t.Errorf("intermissions %d, want 0", fakePlatform.IntermissionCalls)
	}
	if active, ok := fakePlatform.LastGameplay(); !ok || active {
		t.Error("gameplay must stop on the logical pause")
	}
}

func TestApp_ApplyTransition_NoIntermission(t *testing.T) {
	app, _ := newAppTestEnv()
	fakePlatform := &testutil.FakePlatform{}
	app.platformAdapter = fakePlatform

	_ = app.applyTransition(types.StateTransition{
		Target: types.TransitionToStage,
		Level:  1,
	})

	if fakePlatform.IntermissionCalls != 0 {
		t.Errorf("intermissions %d, want 0", fakePlatform.IntermissionCalls)
	}
}

// Сцена главного меню на настоящем графе: название собирается,
// по полю ездят танки с ИИ без игроков и штаба, сцена не падает
func TestApp_MainMenuDemoScene(t *testing.T) {
	app := newFullApp(t)
	for _, done := app.loader.Step(); !done; _, done = app.loader.Step() {
	}

	scene, err := app.newDemoScene()
	if err != nil {
		t.Fatalf("демо-сцена: %v", err)
	}
	// 10 секунд сцены: сборка, выезд танков, стрельба, починка
	for range 600 {
		scene.Update()
	}
	if !scene.IsAssembled() {
		t.Error("название должно собраться")
	}
}

// lostStageWithRewards — уровень на настоящем графе с rewarded
// у площадки, проигранный потерей всех жизней; меню итогов уже
// показано. Подмены ввода и площадки откатываются по завершении теста
func lostStageWithRewards(
	t *testing.T,
) (*App, *states.StageState, *testutil.FakeMenuInput, *testutil.FakeReward) {
	t.Helper()
	app := newFullApp(t)
	for _, done := app.loader.Step(); !done; _, done = app.loader.Step() {
	}

	menuInput := &testutil.FakeMenuInput{}
	reward := &testutil.FakeReward{Available: true}
	defaultMenuInput, defaultReward := app.menuInput, app.rewardAdapter
	app.menuInput, app.rewardAdapter = menuInput, reward
	helpShown := app.settings.IsHelpShown()
	// Страница «Как играть» уже показана: тест начинает сразу с боя
	app.settings.SetHelpShown(true)
	t.Cleanup(func() {
		app.menuInput, app.rewardAdapter = defaultMenuInput, defaultReward
		app.settings.SetHelpShown(helpShown)
	})

	err := app.applyTransition(types.StateTransition{
		Target: types.TransitionToStage,
		Level:  1,
	})
	if err != nil {
		t.Fatalf("переход на уровень: %v", err)
	}
	stage, ok := app.state.(*states.StageState)
	if !ok {
		t.Fatalf("состояние %T, ожидалось StageState", app.state)
	}

	stage.Update()
	stageSession := app.session.StageSession()
	for i := range stageSession.GetPlayerCount() {
		stageSession.SetPlayerLives(types.PlayerTankNum(i), 0)
	}
	stage.Update() // итог уровня
	menuInput.Confirm = true
	stage.Update() // экран итогов показан целиком
	menuInput.Reset()
	stage.Update() // меню итогов принимает ввод
	return app, stage, menuInput, reward
}

// selectResultItem спускается на index-й пункт меню итогов
// и подтверждает его
func selectResultItem(
	stage *states.StageState,
	menuInput *testutil.FakeMenuInput,
	index int,
) types.StateTransition {
	for range index {
		menuInput.Down = true
		stage.Update()
		menuInput.Reset()
	}
	menuInput.Confirm = true
	transition := stage.Update()
	menuInput.Reset()
	return transition
}

// Поражение: RETRY, REVIVE, RETRY+BOOST, STAGES. Второй шанс
// за рекламу продолжает уровень с того же места
func TestApp_StageRevive_FullApp(t *testing.T) {
	_, stage, menuInput, reward := lostStageWithRewards(t)

	selectResultItem(stage, menuInput, 1)
	if reward.Requests != 1 {
		t.Fatalf("reward requests %d, want 1", reward.Requests)
	}

	reward.Status = types.RewardStatusPending
	stage.Update()
	if stage.IsGameplayActive() {
		t.Fatal("stage must wait for the ad")
	}

	reward.Status = types.RewardStatusGranted
	stage.Update()
	if !stage.IsGameplayActive() {
		t.Error("stage must continue after revive")
	}
}

// Реклама не досмотрена: награды нет, меню итогов остаётся
func TestApp_StageReviveDenied_FullApp(t *testing.T) {
	_, stage, menuInput, reward := lostStageWithRewards(t)

	selectResultItem(stage, menuInput, 1)
	reward.Status = types.RewardStatusDenied
	if transition := stage.Update(); transition.Target != types.TransitionNone {
		t.Errorf("transition %v without reward", transition.Target)
	}
	if stage.IsGameplayActive() {
		t.Error("stage must stay on the result screen")
	}
}

// Усиленный перенос: тот же уровень заново, без второй рекламы
func TestApp_StageBoostRetry_FullApp(t *testing.T) {
	app, stage, menuInput, reward := lostStageWithRewards(t)

	selectResultItem(stage, menuInput, 2)
	reward.Status = types.RewardStatusGranted
	transition := stage.Update()

	want := types.StateTransition{
		Target:    types.TransitionToStage,
		Level:     1,
		CarryOver: true,
	}
	if transition != want {
		t.Errorf("transition %+v, want %+v", transition, want)
	}
	if !app.session.StageSession().HasCarryOverAdvantage() {
		t.Error("boost must prepare a carry-over advantage")
	}
}

// Мышь управляет только меню: в бою курсор спрятан, вне боя виден
func TestApp_Update_CursorHiddenInGameplay(t *testing.T) {
	app, _ := newAppTestEnv()
	window := &testutil.FakeWindow{}
	app.windowAdapter = window
	state := &stubStageLikeState{active: true}
	app.state = state

	_ = app.Update()
	if !window.CursorHidden {
		t.Error("в бою курсор спрятан")
	}

	state.active = false
	_ = app.Update()
	if window.CursorHidden {
		t.Error("на паузе и в меню курсор виден")
	}
}

// Перед первым запуском уровня 1 открыта страница «Как играть»:
// бой не начат, пока её не закроют
func TestApp_FirstStageHelp_FullApp(t *testing.T) {
	app := newFullApp(t)
	for _, done := app.loader.Step(); !done; _, done = app.loader.Step() {
	}
	helpShown := app.settings.IsHelpShown()
	app.settings.SetHelpShown(false)
	t.Cleanup(func() { app.settings.SetHelpShown(helpShown) })

	err := app.applyTransition(types.StateTransition{
		Target: types.TransitionToStage,
		Level:  1,
	})
	if err != nil {
		t.Fatalf("переход на уровень: %v", err)
	}
	stage, ok := app.state.(*states.StageState)
	if !ok {
		t.Fatalf("состояние %T, ожидалось StageState", app.state)
	}

	stage.Update()
	if stage.IsGameplayActive() {
		t.Error("пока открыта страница «Как играть», бой не идёт")
	}

	app.settings.SetHelpShown(true)
	stage.Update()
	if !stage.IsGameplayActive() {
		t.Error("после закрытия страницы бой начинается")
	}
}
