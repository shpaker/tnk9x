package bindings

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

// Все имена раскладки по умолчанию знакомы движку
func TestNames_DefaultsAreKnown(t *testing.T) {
	controls := types.NewControlsEntity()
	for _, slot := range controls.KeySlots() {
		if _, ok := KeyByName(controls.GetName(slot)); !ok {
			t.Errorf("клавиша %q неизвестна", controls.GetName(slot))
		}
	}
	for action := range types.InputActionsCount {
		name := controls.GetButton(types.PlayerTankNumPlayer1, action)
		if _, ok := ButtonByName(name); !ok {
			t.Errorf("кнопка %q неизвестна", name)
		}
	}
	if _, ok := KeyByName(types.ReservedKey); !ok {
		t.Error("зарезервированная клавиша должна быть клавишей движка")
	}
	if _, ok := ButtonByName(types.ReservedButton); !ok {
		t.Error("зарезервированная кнопка должна быть кнопкой движка")
	}
}

func TestNames_ButtonsRoundTrip(t *testing.T) {
	for name, button := range buttonsByName {
		if back, ok := ButtonName(button); !ok || back != name {
			t.Errorf("кнопка %q -> %q", name, back)
		}
	}
}

func TestNames_KeyLabel(t *testing.T) {
	cases := map[string]string{
		"Quote":   "'",
		"W":       "W",
		"Space":   "SPACE",
		"Digit7":  "7",
		"Numpad0": "NUM0",
		"ArrowUp": "UP",
	}
	for name, want := range cases {
		if got := KeyLabel(name); got != want {
			t.Errorf("KeyLabel(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestNames_RepairUnknown(t *testing.T) {
	controls := types.NewControlsEntity()
	fire := types.ControlsSlot{
		Page: types.ControlsPagePlayer1, Action: types.InputActionFire,
	}
	pad := fire
	pad.Device = types.ControlsDeviceGamepad
	controls.SetName(fire, "NoSuchKey")
	controls.SetName(pad, "Z9")

	RepairUnknown(controls)
	if controls.GetName(fire) != "G" || controls.GetName(pad) != "A" {
		t.Errorf("неизвестные имена должны смениться умолчаниями: %q %q",
			controls.GetName(fire), controls.GetName(pad))
	}
}
