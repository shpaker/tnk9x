package menu_input

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
)

// fakeDevices — клавиатура и один геймпад вместо ebiten-рантайма
type fakeDevices struct {
	keys    map[ebiten.Key]bool
	buttons map[ebiten.StandardGamepadButton]bool
	stickX  float64
	stickY  float64
	padless bool
}

func newTestAdapter() (*MenuInputAdapter, *fakeDevices, *testutil.FakeTouchControls) {
	devices := &fakeDevices{}
	touch := &testutil.FakeTouchControls{}
	adapter := NewMenuInputAdapter(touch)
	adapter.isKeyJustPressed = func(key ebiten.Key) bool {
		return devices.keys[key]
	}
	adapter.appendGamepads = func(ids []ebiten.GamepadID) []ebiten.GamepadID {
		if devices.padless {
			return ids
		}
		return append(ids, 0)
	}
	adapter.isButtonJustPressed = func(
		_ ebiten.GamepadID,
		button ebiten.StandardGamepadButton,
	) bool {
		return devices.buttons[button]
	}
	adapter.axisValue = func(
		_ ebiten.GamepadID,
		axis ebiten.StandardGamepadAxis,
	) float64 {
		if axis == ebiten.StandardGamepadAxisLeftStickHorizontal {
			return devices.stickX
		}
		return devices.stickY
	}
	return adapter, devices, touch
}

func TestMenuInput_Keyboard(t *testing.T) {
	adapter, devices, _ := newTestAdapter()

	devices.keys = map[ebiten.Key]bool{ebiten.KeyS: true, ebiten.KeyEnter: true}
	adapter.Update()
	if up, down := adapter.Steps(); up || !down || !adapter.Confirmed() {
		t.Error("S — вниз, Enter — выбор")
	}

	devices.keys = map[ebiten.Key]bool{ebiten.KeyEscape: true}
	adapter.Update()
	if !adapter.Back() || !adapter.PauseJustPressed() {
		t.Error("Esc — назад и пауза")
	}
	if adapter.Confirmed() {
		t.Error("события прошлого кадра не должны оставаться")
	}
}

func TestMenuInput_GamepadButtons(t *testing.T) {
	adapter, devices, _ := newTestAdapter()

	devices.buttons = map[ebiten.StandardGamepadButton]bool{
		ebiten.StandardGamepadButtonLeftRight:   true,
		ebiten.StandardGamepadButtonRightBottom: true,
	}
	adapter.Update()
	if adapter.SideStep() != 1 || !adapter.Confirmed() {
		t.Error("крестовина вправо — шаг, A — выбор")
	}

	devices.buttons = map[ebiten.StandardGamepadButton]bool{
		ebiten.StandardGamepadButtonRightRight: true,
	}
	adapter.Update()
	if !adapter.Back() || adapter.PauseJustPressed() {
		t.Error("B — назад, но не пауза")
	}

	devices.buttons = map[ebiten.StandardGamepadButton]bool{
		ebiten.StandardGamepadButtonCenterRight: true,
	}
	adapter.Update()
	if !adapter.Back() || !adapter.PauseJustPressed() {
		t.Error("Start — назад и пауза")
	}
}

// Отклонение стика — один шаг, удержание не листает
func TestMenuInput_StickEdges(t *testing.T) {
	adapter, devices, _ := newTestAdapter()

	devices.stickY = -0.9
	adapter.Update()
	if up, _ := adapter.Steps(); !up {
		t.Fatal("стик вверх — шаг вверх")
	}
	adapter.Update()
	if up, _ := adapter.Steps(); up {
		t.Error("удержание стика не должно давать новых шагов")
	}

	devices.stickY = 0
	devices.stickX = 0.8
	adapter.Update()
	if adapter.SideStep() != 1 {
		t.Error("смена направления стика — новый шаг")
	}

	// Отключённый геймпад забывается
	devices.padless = true
	adapter.Update()
	if len(adapter.sticks) != 0 {
		t.Error("состояние отключённого геймпада должно удаляться")
	}
}

func TestMenuInput_TouchBothPlayers(t *testing.T) {
	adapter, _, touch := newTestAdapter()

	touch.Directions[1] = types.DirectionLeft
	touch.DirectionJust[1] = true
	touch.FireJust[1] = true
	adapter.Update()
	if adapter.SideStep() != -1 || !adapter.Confirmed() {
		t.Error("контроллы P2 тоже управляют меню")
	}

	touch.DirectionJust[1] = false
	touch.FireJust[1] = false
	touch.PauseJust = true
	adapter.Update()
	if !adapter.Back() || !adapter.PauseJustPressed() {
		t.Error("тач-пауза — назад и пауза")
	}
}
