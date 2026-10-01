package types_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

func TestControlsEntity_DefaultsAreUnique(t *testing.T) {
	controls := types.NewControlsEntity()

	seen := map[string]bool{}
	for _, slot := range controls.KeySlots() {
		name := controls.GetName(slot)
		if name == "" || seen[name] {
			t.Errorf("клавиша %q пустая или повторяется", name)
		}
		seen[name] = true
	}
	if controls.GetKey(
		types.PlayerTankNumPlayer1,
		types.InputActionFire,
	) != "G" ||
		controls.GetKey(
			types.PlayerTankNumPlayer2,
			types.InputActionUp,
		) != "I" {
		t.Error("умолчания: P1 стреляет на G, P2 едет вверх на I")
	}
}

func TestControlsEntity_SetNameAndReset(t *testing.T) {
	controls := types.NewControlsEntity()
	slot := types.ControlsSlot{
		Page:   types.ControlsPagePlayer2,
		Action: types.InputActionFire,
		Device: types.ControlsDeviceGamepad,
	}

	controls.SetName(slot, "B")
	if controls.GetButton(
		types.PlayerTankNumPlayer2,
		types.InputActionFire,
	) != "B" {
		t.Error("кнопка огня P2 должна смениться на B")
	}

	hotkey := types.ControlsSlot{
		Page: types.ControlsPageHotkeys, Hotkey: types.HotkeyFullscreen,
	}
	controls.SetName(hotkey, "F10")
	if controls.GetHotkey(types.HotkeyFullscreen) != "F10" {
		t.Error("хоткей полного экрана должен смениться на F10")
	}

	controls.Reset()
	if controls.GetName(slot) != "A" ||
		controls.GetHotkey(types.HotkeyFullscreen) != "F11" {
		t.Error("сброс должен вернуть умолчания")
	}
}
