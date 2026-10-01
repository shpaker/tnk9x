package states_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/states"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// memorySettingsRepository и memoryControlsRepository — хранилища
// без диска
type memorySettingsRepository struct{}

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

// nopRenderer — рендеры экранов без отрисовки
type nopRenderer struct{}

func (nopRenderer) Draw(*ebiten.Image, types.ControlsViewData) {}

type nopSettingsRenderer struct{}

func (nopSettingsRenderer) Draw(*ebiten.Image, types.SettingsViewData) {}

// menusEnv — экраны настроек и раскладки на фейках
type menusEnv struct {
	input           *testutil.FakeMenuInput
	capture         *fakeCapture
	window          *fakeWindow
	settings        *types.SettingsEntity
	controls        *types.ControlsEntity
	controlsOverlay *states.ControlsOverlay
	settingsOverlay *states.SettingsOverlay
}

func newMenusEnv() *menusEnv {
	env := &menusEnv{
		input:    &testutil.FakeMenuInput{},
		capture:  &fakeCapture{},
		window:   &fakeWindow{},
		settings: types.NewSettingsEntity(),
		controls: types.NewControlsEntity(),
	}
	env.controlsOverlay = states.NewControlsOverlay(
		use_cases.NewControlsUseCases(memoryControlsRepository{}),
		nopRenderer{},
		env.input,
		env.capture,
		env.controls,
	)
	env.settingsOverlay = states.NewSettingsOverlay(
		use_cases.NewSettingsUseCases(memorySettingsRepository{}, true),
		nopSettingsRenderer{},
		env.input,
		&testutil.FakeSoundPlayer{},
		env.window,
		env.controlsOverlay,
		env.settings,
	)
	return env
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
		use_cases.NewSettingsUseCases(memorySettingsRepository{}, true),
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

type nopLevelSelectRenderer struct{}

func (nopLevelSelectRenderer) Draw(*ebiten.Image, types.LevelSelectViewData) {}
func (nopLevelSelectRenderer) DrawMenu(
	*ebiten.Image,
	types.LevelSelectMenuViewData,
) {
}

func TestLevelSelectState_PlayersAndQuit(t *testing.T) {
	env := newMenusEnv()
	state := states.NewLevelSelectState(states.LevelSelectStateDependencies{
		LevelSelectUseCases: stubLevelSelectUseCases{},
		SettingsUseCases: use_cases.NewSettingsUseCases(
			memorySettingsRepository{}, true,
		),
		Renderer:        nopLevelSelectRenderer{},
		MenuInput:       env.input,
		SettingsOverlay: env.settingsOverlay,
		Settings:        env.settings,
		QuitAvailable:   true,
	})
	update := func() { state.Update() }

	// Esc открывает меню, BACK -> PLAYERS, вправо — двое игроков
	env.frame(update, testutil.FakeMenuInput{BackPressed: true})
	env.frame(update, testutil.FakeMenuInput{Down: true})
	env.frame(update, testutil.FakeMenuInput{Side: 1})
	if env.settings.GetPlayers() != 2 {
		t.Errorf("игроков %d, ожидалось 2", env.settings.GetPlayers())
	}

	// PLAYERS -> SETTINGS -> QUIT
	env.frame(update, testutil.FakeMenuInput{Down: true})
	env.frame(update, testutil.FakeMenuInput{Down: true})
	*env.input = testutil.FakeMenuInput{Confirm: true}
	if transition := state.Update(); transition.Target != types.TransitionToQuit {
		t.Errorf("QUIT должен завершать игру, переход %v", transition.Target)
	}
}
