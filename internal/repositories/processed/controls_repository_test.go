package processed

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

func TestControlsRepository_RoundTrip(t *testing.T) {
	storage := &memoryStorage{data: map[string][]byte{}}
	repository := NewControlsRepository(storage)

	controls := types.NewControlsEntity()
	controls.SetName(types.ControlsSlot{
		Page: types.ControlsPagePlayer2, Action: types.InputActionFire,
	}, "Enter")
	controls.SetName(types.ControlsSlot{
		Page:   types.ControlsPagePlayer1,
		Action: types.InputActionFire,
		Device: types.ControlsDeviceGamepad,
	}, "B")
	controls.SetName(types.ControlsSlot{
		Page: types.ControlsPageHotkeys, Hotkey: types.HotkeyFullscreen,
	}, "F10")
	if err := repository.SaveControls(controls); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := repository.GetControls()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if *loaded != *controls {
		t.Errorf("loaded %+v, want %+v", loaded, controls)
	}
}

// Пустое хранилище и отсутствующие поля — раскладка по умолчанию
func TestControlsRepository_Defaults(t *testing.T) {
	storage := &memoryStorage{data: map[string][]byte{}}
	repository := NewControlsRepository(storage)

	controls, err := repository.GetControls()
	if err != nil || *controls != *types.NewControlsEntity() {
		t.Errorf("empty storage: %v, controls %+v", err, controls)
	}

	storage.data[controlsKey] = []byte(
		`{"version":1,"players":[{"keys":{"fire":"Space"}}]}`,
	)
	controls, err = repository.GetControls()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if controls.GetKey(
		types.PlayerTankNumPlayer1,
		types.InputActionFire,
	) != "Space" ||
		controls.GetKey(
			types.PlayerTankNumPlayer1,
			types.InputActionUp,
		) != "W" ||
		controls.GetHotkey(types.HotkeyGraphics) != "F2" {
		t.Errorf("partial controls %+v", controls)
	}
}

// Повреждённое или чужое сохранение — ошибка, но игра получает
// умолчания
func TestControlsRepository_Broken(t *testing.T) {
	for _, data := range []string{`{broken`, `{"version":99}`} {
		storage := &memoryStorage{data: map[string][]byte{
			controlsKey: []byte(data),
		}}
		controls, err := NewControlsRepository(storage).GetControls()
		if err == nil {
			t.Errorf("%s: ожидалась ошибка", data)
		}
		if *controls != *types.NewControlsEntity() {
			t.Errorf("%s: controls %+v, ожидались умолчания", data, controls)
		}
	}
}
