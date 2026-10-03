package states_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/states"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// memorySettingsRepository, memoryControlsRepository
// и memoryProgressRepository — хранилища без диска
type memorySettingsRepository struct{}

type memoryProgressRepository struct{}

func (memoryProgressRepository) GetProgress() (*types.ProgressEntity, error) {
	return types.NewProgressEntity(), nil
}

func (memoryProgressRepository) SaveProgress(*types.ProgressEntity) error {
	return nil
}

func (memorySettingsRepository) GetSettings() (*types.SettingsEntity, error) {
	return types.NewSettingsEntity(), nil
}

func (memorySettingsRepository) SaveSettings(*types.SettingsEntity) error {
	return nil
}

type memoryControlsRepository struct{}

func (memoryControlsRepository) GetControls() (*types.ControlsEntity, error) {
	return types.NewControlsEntity(), nil
}

func (memoryControlsRepository) SaveControls(*types.ControlsEntity) error {
	return nil
}

// fakeWindow — окно с переключаемым полным экраном
type fakeWindow struct{ fullscreen bool }

func (w *fakeWindow) IsFullscreen() bool { return w.fullscreen }

func (w *fakeWindow) SetFullscreen(
	fullscreen bool,
) {
	w.fullscreen = fullscreen
}

func (w *fakeWindow) SetCursorVisible(bool) {}

// fakeCapture отдаёт заданное нажатие один раз
type fakeCapture struct{ key, button string }

func (c *fakeCapture) JustPressedKey() (string, bool) {
	key := c.key
	c.key = ""
	return key, key != ""
}

func (c *fakeCapture) JustPressedPadButton() (string, bool) {
	button := c.button
	c.button = ""
	return button, button != ""
}

// fakeHotkeys отдаёт заданные хоткеи
type fakeHotkeys struct{ pressed []types.HotkeyAction }

func (h *fakeHotkeys) JustPressed() []types.HotkeyAction { return h.pressed }

// hitRowByY — хит-тест фейковых рендеров: строка под точкой равна
// её Y
func hitRowByY(position types.Position) (int, bool) {
	return int(position.Y), true
}

// pointAt и clickAt — наведение и клик по строке row; x выбирает
// колонку на экране раскладки
func pointAt(row int, x float64) testutil.FakeMenuInput {
	return testutil.FakeMenuInput{
		PointPosition: types.Position{X: x, Y: float64(row)},
		PointMoved:    true,
		PointerActive: true,
	}
}

func clickAt(row int, x float64) testutil.FakeMenuInput {
	return testutil.FakeMenuInput{
		TapPosition:   types.Position{X: x, Y: float64(row)},
		TapPressed:    true,
		PointerActive: true,
	}
}

// nopRenderer — рендеры экранов без отрисовки; колонка геймпада
// раскладки — при X >= 1
type nopRenderer struct{}

func (nopRenderer) Draw(*ebiten.Image, types.ControlsViewData) {}

func (nopRenderer) HitCell(
	position types.Position,
) (int, types.ControlsDevice, bool) {
	row, ok := hitRowByY(position)
	if position.X >= 1 {
		return row, types.ControlsDeviceGamepad, ok
	}
	return row, types.ControlsDeviceKeyboard, ok
}

type nopSettingsRenderer struct{}

func (nopSettingsRenderer) Draw(*ebiten.Image, types.SettingsViewData) {}

func (nopSettingsRenderer) HitRow(position types.Position) (int, bool) {
	return hitRowByY(position)
}

// menusEnv — экраны настроек и раскладки на фейках
type menusEnv struct {
	input           *testutil.FakeMenuInput
	capture         *fakeCapture
	window          *fakeWindow
	settings        *types.SettingsEntity
	controls        *types.ControlsEntity
	texts           *testutil.FakeTexts
	controlsOverlay *states.ControlsOverlay
	settingsOverlay *states.SettingsOverlay
	purchase        *testutil.FakePurchase
	inventory       *types.InventoryEntity
	shopOverlay     *states.ShopOverlay
}

func newMenusEnv() *menusEnv {
	env := &menusEnv{
		input:    &testutil.FakeMenuInput{},
		capture:  &fakeCapture{},
		window:   &fakeWindow{},
		settings: types.NewSettingsEntity(),
		controls: types.NewControlsEntity(),
		texts:    &testutil.FakeTexts{Languages: testLanguageConfig.Languages},
	}
	env.controlsOverlay = states.NewControlsOverlay(
		use_cases.NewControlsUseCases(memoryControlsRepository{}),
		nopRenderer{},
		env.input,
		env.capture,
		env.controls,
	)
	env.settingsOverlay = states.NewSettingsOverlay(
		newSettingsUseCases(),
		use_cases.NewLocalizationUseCases(
			env.texts, testLanguageConfig, "",
		),
		nopSettingsRenderer{},
		env.input,
		&testutil.FakeSoundPlayer{},
		env.window,
		env.controlsOverlay,
		env.settings,
	)
	env.purchase = &testutil.FakePurchase{}
	env.inventory = types.NewInventoryEntity()
	repository := &testutil.MemoryInventoryRepository{}
	campaign := testShopCampaign()
	progression := use_cases.NewProgressionUseCases(
		campaign,
		types.NewProgressEntity(),
		env.inventory,
		memoryProgressRepository{},
	)
	env.shopOverlay = states.NewShopOverlay(
		use_cases.NewShopUseCases(
			testShopProducts,
			campaign,
			env.inventory,
			[]interfaces.IProgressionUseCases{progression},
			env.purchase,
			repository,
		),
		use_cases.NewInventoryUseCases(env.inventory, repository),
		nopShopRenderer{},
		env.input,
		&testutil.FakeSoundPlayer{},
	)
	return env
}

// testLanguageConfig — языки интерфейса в тестах меню
var testLanguageConfig = types.LanguageConfig{
	Languages: []types.Language{"en", "ru"},
	Default:   "en",
}

// newSettingsUseCases — use cases настроек с фейковыми текстами
func newSettingsUseCases() *use_cases.SettingsUseCases {
	return use_cases.NewSettingsUseCases(
		memorySettingsRepository{},
		&testutil.FakeTexts{Languages: testLanguageConfig.Languages},
		true,
		testLanguageConfig.Languages,
	)
}

// frame — один кадр с заданным вводом
func (env *menusEnv) frame(update func(), input testutil.FakeMenuInput) {
	*env.input = input
	update()
	env.input.Reset()
}

func TestHotkeysHandler(t *testing.T) {
	env := newMenusEnv()
	hotkeys := &fakeHotkeys{}
	handler := states.NewHotkeysHandler(
		newSettingsUseCases(),
		hotkeys,
		env.window,
		env.settingsOverlay,
		env.settings,
	)

	hotkeys.pressed = []types.HotkeyAction{
		types.HotkeyGraphics, types.HotkeyFullscreen,
	}
	handler.Update()
	if env.settings.IsEffectsEnabled() {
		t.Error("хоткей графики переключает NORMAL на CLASSIC")
	}
	if !env.window.fullscreen || !env.settings.IsFullscreen() {
		t.Error("хоткей полного экрана разворачивает окно")
	}

	// Пока назначают клавишу, хоткеи молчат
	env.settingsOverlay.Open()
	env.controlsOverlay.Open()
	env.frame(env.settingsOverlay.Update, testutil.FakeMenuInput{Down: true})
	env.frame(env.settingsOverlay.Update, testutil.FakeMenuInput{Confirm: true})
	if !env.settingsOverlay.IsCapturing() {
		t.Fatal("выбор ячейки включает ожидание нажатия")
	}
	handler.Update()
	if env.settings.IsEffectsEnabled() {
		t.Error("во время назначения хоткеи не срабатывают")
	}
}

func TestControlsOverlay_BindRejectCancel(t *testing.T) {
	env := newMenusEnv()
	env.controlsOverlay.Open()

	// Заголовок -> UP, выбор ячейки клавиатуры
	env.frame(env.controlsOverlay.Update, testutil.FakeMenuInput{Down: true})
	env.frame(env.controlsOverlay.Update, testutil.FakeMenuInput{Confirm: true})
	if !env.controlsOverlay.IsCapturing() {
		t.Fatal("ожидание нажатия должно включиться")
	}

	// Занятая клавиша отклоняется, ожидание продолжается
	env.capture.key = "S"
	env.frame(env.controlsOverlay.Update, testutil.FakeMenuInput{})
	if !env.controlsOverlay.IsCapturing() ||
		env.controls.GetKey(
			types.PlayerTankNumPlayer1,
			types.InputActionUp,
		) != "W" {
		t.Error("занятая клавиша не назначается")
	}

	env.capture.key = "ArrowUp"
	env.frame(env.controlsOverlay.Update, testutil.FakeMenuInput{})
	if env.controlsOverlay.IsCapturing() ||
		env.controls.GetKey(
			types.PlayerTankNumPlayer1,
			types.InputActionUp,
		) != "ArrowUp" {
		t.Error("свободная клавиша назначается и завершает ожидание")
	}

	// Пауза отменяет ожидание, ничего не меняя
	env.frame(env.controlsOverlay.Update, testutil.FakeMenuInput{Confirm: true})
	env.frame(env.controlsOverlay.Update, testutil.FakeMenuInput{Pause: true})
	if env.controlsOverlay.IsCapturing() || !env.controlsOverlay.IsOpen() {
		t.Error("Esc отменяет ожидание и оставляет экран открытым")
	}
}

func TestControlsOverlay_PadColumnAndPages(t *testing.T) {
	env := newMenusEnv()
	env.controlsOverlay.Open()

	// Вправо на заголовке — страница P2; вниз, вправо — колонка PAD
	env.frame(env.controlsOverlay.Update, testutil.FakeMenuInput{Side: 1})
	env.frame(env.controlsOverlay.Update, testutil.FakeMenuInput{Down: true})
	env.frame(env.controlsOverlay.Update, testutil.FakeMenuInput{Side: 1})
	env.frame(env.controlsOverlay.Update, testutil.FakeMenuInput{Confirm: true})

	env.capture.key = "Q"
	env.capture.button = "Y"
	env.frame(env.controlsOverlay.Update, testutil.FakeMenuInput{})
	if env.controls.GetButton(
		types.PlayerTankNumPlayer2,
		types.InputActionUp,
	) != "Y" ||
		env.controls.GetKey(
			types.PlayerTankNumPlayer2,
			types.InputActionUp,
		) != "I" {
		t.Error(
			"в колонке PAD назначается кнопка геймпада P2, клавиши не трогаются",
		)
	}

	// Назад закрывает экран
	env.frame(
		env.controlsOverlay.Update,
		testutil.FakeMenuInput{BackPressed: true},
	)
	if env.controlsOverlay.IsOpen() {
		t.Error("назад закрывает экран раскладки")
	}
}

// stubLevelSelectUseCases — выбор уровня без кампании
type stubLevelSelectUseCases struct{}

func (stubLevelSelectUseCases) NewSelector(int) *types.LevelSelectorEntity {
	return &types.LevelSelectorEntity{}
}
func (stubLevelSelectUseCases) MoveLevel(*types.LevelSelectorEntity, int)   {}
func (stubLevelSelectUseCases) MovePack(*types.LevelSelectorEntity, int)    {}
func (stubLevelSelectUseCases) SetPosition(*types.LevelSelectorEntity, int) {}

func (stubLevelSelectUseCases) SelectedLevel(
	*types.LevelSelectorEntity,
) (int, bool) {
	return 1, true
}

func (stubLevelSelectUseCases) BuildView(
	*types.LevelSelectorEntity,
	map[int]*types.LevelEntity,
) types.LevelSelectViewData {
	return types.LevelSelectViewData{}
}

// nopLevelSelectRenderer — кнопка выхода занимает левый верхний
// квадрат 10x10
type nopLevelSelectRenderer struct{}

func (nopLevelSelectRenderer) Draw(*ebiten.Image, types.LevelSelectViewData) {}

func (nopLevelSelectRenderer) HitBack(position types.Position) bool {
	return position.X < 10 && position.Y < 10
}

// HitLevel — ячейки уровней в полосе Y 20..30, позиция равна X/10
func (nopLevelSelectRenderer) HitLevel(position types.Position) (int, bool) {
	if position.Y < 20 || position.Y >= 30 {
		return 0, false
	}
	return int(position.X) / 10, true
}

// HitPack — стрелка следующей пачки в полосе Y 40..50
func (nopLevelSelectRenderer) HitPack(position types.Position) (int, bool) {
	return 1, position.Y >= 40 && position.Y < 50
}

type nopMainMenuRenderer struct{}

func (nopMainMenuRenderer) Draw(*ebiten.Image, types.MainMenuViewData) {}

func (nopMainMenuRenderer) HitRow(position types.Position) (int, bool) {
	return hitRowByY(position)
}

// countingScene — сцена за меню, считающая кадры и заглушения звука
type countingScene struct {
	updates int
	stops   int
}

func (s *countingScene) Update()            { s.updates++ }
func (s *countingScene) Draw(*ebiten.Image) {}
func (s *countingScene) StopSounds()        { s.stops++ }

func newMainMenu(
	env *menusEnv,
	scene *countingScene,
) *states.MainMenuState {
	return states.NewMainMenuState(states.MainMenuStateDependencies{
		SettingsUseCases: newSettingsUseCases(),
		Renderer:         nopMainMenuRenderer{},
		MenuInput:        env.input,
		Scene:            scene,
		SettingsOverlay:  env.settingsOverlay,
		ShopOverlay:      env.shopOverlay,
		Settings:         env.settings,
		QuitAvailable:    true,
	})
}

func TestMainMenuState_ModesAndQuit(t *testing.T) {
	env := newMenusEnv()
	scene := &countingScene{}
	menu := newMainMenu(env, scene)

	// 1 PLAYER -> 2 PLAYERS: выбор режима ведёт к выбору уровня
	env.frame(func() { menu.Update() }, testutil.FakeMenuInput{Down: true})
	*env.input = testutil.FakeMenuInput{Confirm: true}
	transition := menu.Update()
	if transition.Target != types.TransitionToLevelSelect ||
		env.settings.GetPlayers() != 2 {
		t.Errorf("2 PLAYERS: переход %v, игроков %d",
			transition.Target, env.settings.GetPlayers())
	}
	if scene.stops != 1 {
		t.Errorf("уход из меню глушит звуки сцены, заглушений %d", scene.stops)
	}

	// Курсор нового меню встаёт на последний режим; SETTINGS, QUIT
	scene = &countingScene{}
	menu = newMainMenu(env, scene)
	env.frame(func() { menu.Update() }, testutil.FakeMenuInput{Down: true})
	env.frame(func() { menu.Update() }, testutil.FakeMenuInput{Down: true})
	*env.input = testutil.FakeMenuInput{Confirm: true}
	if transition := menu.Update(); transition.Target != types.TransitionToQuit {
		t.Errorf("QUIT должен завершать игру, переход %v", transition.Target)
	}
	if scene.stops != 1 {
		t.Errorf("выход глушит звуки сцены, заглушений %d", scene.stops)
	}
}

func TestLevelSelectState_BackToMainMenu(t *testing.T) {
	env := newMenusEnv()
	state := states.NewLevelSelectState(states.LevelSelectStateDependencies{
		LevelSelectUseCases: stubLevelSelectUseCases{},
		Renderer:            nopLevelSelectRenderer{},
		MenuInput:           env.input,
		Settings:            env.settings,
	})

	*env.input = testutil.FakeMenuInput{Confirm: true}
	if transition := state.Update(); transition.Target != types.TransitionToStage {
		t.Errorf(
			"выбор открытого уровня запускает его, переход %v",
			transition.Target,
		)
	}
	*env.input = testutil.FakeMenuInput{BackPressed: true}
	if transition := state.Update(); transition.Target != types.TransitionToMainMenu {
		t.Errorf("назад ведёт в главное меню, переход %v", transition.Target)
	}

	// Тап мимо кнопки выхода не действует, тап по ней — в главное меню
	*env.input = testutil.FakeMenuInput{
		TapPosition: types.Position{X: 100, Y: 100}, TapPressed: true,
	}
	if transition := state.Update(); transition.Target != types.TransitionNone {
		t.Errorf(
			"тап мимо кнопки не должен уходить, переход %v",
			transition.Target,
		)
	}
	*env.input = testutil.FakeMenuInput{
		TapPosition: types.Position{X: 5, Y: 5}, TapPressed: true,
	}
	if transition := state.Update(); transition.Target != types.TransitionToMainMenu {
		t.Errorf(
			"тап по кнопке ведёт в главное меню, переход %v",
			transition.Target,
		)
	}
}

func TestMainMenuState_Mouse(t *testing.T) {
	env := newMenusEnv()
	menu := newMainMenu(env, &countingScene{})

	// Наведение на 2 PLAYERS только выделяет пункт
	*env.input = pointAt(1, 0)
	if transition := menu.Update(); transition.Target != types.TransitionNone {
		t.Errorf("наведение не выбирает пункт, переход %v", transition.Target)
	}

	// Клик по QUIT выбирает его сразу
	*env.input = clickAt(3, 0)
	if transition := menu.Update(); transition.Target != types.TransitionToQuit {
		t.Errorf("клик по QUIT завершает игру, переход %v", transition.Target)
	}
}

func TestSettingsOverlay_Mouse(t *testing.T) {
	env := newMenusEnv()
	env.settingsOverlay.Open()
	volume := env.settings.GetVolume()

	// Наведение на VOLUME и колесо меняют громкость
	env.frame(env.settingsOverlay.Update, pointAt(2, 0))
	env.frame(env.settingsOverlay.Update, testutil.FakeMenuInput{Side: -1})
	if env.settings.GetVolume() >= volume {
		t.Errorf(
			"колесо над VOLUME убавляет громкость: %v",
			env.settings.GetVolume(),
		)
	}

	// Клик по CONTROLS открывает раскладку, правая кнопка закрывает
	env.frame(env.settingsOverlay.Update, clickAt(4, 0))
	if !env.controlsOverlay.IsOpen() {
		t.Fatal("клик по CONTROLS открывает раскладку")
	}
	env.frame(
		env.settingsOverlay.Update,
		testutil.FakeMenuInput{BackPressed: true},
	)
	env.frame(
		env.settingsOverlay.Update,
		testutil.FakeMenuInput{BackPressed: true},
	)
	if env.settingsOverlay.IsOpen() {
		t.Error("назад закрывает раскладку, затем настройки")
	}
}

// Смена языка в настройках сразу делает его активным
func TestSettingsOverlay_Language(t *testing.T) {
	env := newMenusEnv()
	env.settingsOverlay.Open()

	env.frame(env.settingsOverlay.Update, pointAt(3, 0))
	env.frame(env.settingsOverlay.Update, testutil.FakeMenuInput{Side: 1})
	env.frame(env.settingsOverlay.Update, testutil.FakeMenuInput{Side: 1})
	if env.settings.GetLanguage() != "ru" || env.texts.GetLanguage() != "ru" {
		t.Errorf(
			"выбран %q, активен %q, ожидался ru",
			env.settings.GetLanguage(), env.texts.GetLanguage(),
		)
	}
}

func TestControlsOverlay_Mouse(t *testing.T) {
	env := newMenusEnv()
	env.controlsOverlay.Open()

	// Клик по заголовку листает страницу: P2
	env.frame(env.controlsOverlay.Update, clickAt(0, 0))

	// Клик по ячейке PAD строки UP начинает ожидание кнопки
	env.frame(env.controlsOverlay.Update, clickAt(1, 1))
	if !env.controlsOverlay.IsCapturing() {
		t.Fatal("клик по ячейке включает ожидание нажатия")
	}
	env.capture.button = "Y"
	env.frame(env.controlsOverlay.Update, testutil.FakeMenuInput{})
	if env.controls.GetButton(
		types.PlayerTankNumPlayer2,
		types.InputActionUp,
	) != "Y" {
		t.Error("кликнутая ячейка PAD P2 получает кнопку")
	}

	// Клик во время ожидания отменяет его
	env.frame(env.controlsOverlay.Update, clickAt(2, 0))
	env.frame(env.controlsOverlay.Update, clickAt(2, 0))
	if env.controlsOverlay.IsCapturing() || !env.controlsOverlay.IsOpen() {
		t.Error("клик отменяет ожидание и оставляет экран открытым")
	}
}

// recordingLevelSelectUseCases — запоминает позицию и пачку
type recordingLevelSelectUseCases struct {
	stubLevelSelectUseCases
	position, packSteps int
}

func (uc *recordingLevelSelectUseCases) SetPosition(
	_ *types.LevelSelectorEntity,
	position int,
) {
	uc.position = position
}

func (uc *recordingLevelSelectUseCases) MovePack(
	_ *types.LevelSelectorEntity,
	delta int,
) {
	uc.packSteps += delta
}

func TestLevelSelectState_Mouse(t *testing.T) {
	env := newMenusEnv()
	useCases := &recordingLevelSelectUseCases{}
	state := states.NewLevelSelectState(states.LevelSelectStateDependencies{
		LevelSelectUseCases: useCases,
		Renderer:            nopLevelSelectRenderer{},
		MenuInput:           env.input,
		Settings:            env.settings,
	})

	// Наведение выбирает уровень, но не запускает
	*env.input = testutil.FakeMenuInput{
		PointPosition: types.Position{X: 25, Y: 25}, PointMoved: true,
	}
	if transition := state.Update(); transition.Target != types.TransitionNone ||
		useCases.position != 2 {
		t.Errorf("наведение: переход %v, позиция %d",
			transition.Target, useCases.position)
	}

	// Клик по стрелке листает пачку
	*env.input = testutil.FakeMenuInput{
		TapPosition: types.Position{Y: 45}, TapPressed: true,
	}
	if transition := state.Update(); transition.Target != types.TransitionNone ||
		useCases.packSteps != 1 {
		t.Errorf("стрелка: переход %v, шагов пачки %d",
			transition.Target, useCases.packSteps)
	}

	// Клик по ячейке запускает уровень
	*env.input = testutil.FakeMenuInput{
		TapPosition: types.Position{X: 15, Y: 25}, TapPressed: true,
	}
	if transition := state.Update(); transition.Target != types.TransitionToStage ||
		useCases.position != 1 {
		t.Errorf("клик по ячейке: переход %v, позиция %d",
			transition.Target, useCases.position)
	}
}

// fakeLoader — загрузка из steps шагов
type fakeLoader struct{ done, steps int }

func (l *fakeLoader) Step() (float64, bool) {
	l.done = min(l.done+1, l.steps)
	return float64(l.done) / float64(l.steps), l.done == l.steps
}

type nopSplashRenderer struct{}

func (nopSplashRenderer) Draw(*ebiten.Image, types.SplashViewData) {}

// splashTicksToMenu — кадров до перехода в главное меню
func splashTicksToMenu(
	splash *states.SplashState,
	input *testutil.FakeMenuInput,
	skip bool,
) int {
	for frame := 1; frame < 1000; frame++ {
		*input = testutil.FakeMenuInput{Confirm: skip}
		if splash.Update().Target == types.TransitionToMainMenu {
			return frame
		}
	}
	return -1
}

func TestSplashState_LoadsThenFades(t *testing.T) {
	input := &testutil.FakeMenuInput{}
	loader := &fakeLoader{steps: 5}
	splash := states.NewSplashState(loader, nopSplashRenderer{}, input)

	frames := splashTicksToMenu(splash, input, false)
	if loader.done != loader.steps {
		t.Fatal("сплеш уходит только после загрузки")
	}
	if frames < 120 {
		t.Errorf("сплеш виден не меньше 2 с, ушёл через %d кадров", frames)
	}

	// После загрузки ожидание можно пропустить
	input = &testutil.FakeMenuInput{}
	skipped := states.NewSplashState(
		&fakeLoader{steps: 5},
		nopSplashRenderer{},
		input,
	)
	if frames := splashTicksToMenu(skipped, input, true); frames < 5 ||
		frames > 60 {
		t.Errorf("пропуск: переход через %d кадров", frames)
	}
}
