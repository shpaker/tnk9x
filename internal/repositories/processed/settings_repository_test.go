package processed

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

func TestSettingsRepository_RoundTrip(t *testing.T) {
	storage := &memoryStorage{data: map[string][]byte{}}
	repository := NewSettingsRepository(storage)

	settings := types.NewSettingsEntity()
	settings.SetEffectsEnabled(false)
	settings.SetFullscreen(false)
	settings.SetVolumeLevel(8)
	settings.SetLanguage("ru")
	settings.SetPlayers(2)
	if err := repository.SaveSettings(settings); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := repository.GetSettings()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.IsEffectsEnabled() || loaded.IsFullscreen() ||
		loaded.GetVolumeLevel() != 8 || loaded.GetLanguage() != "ru" ||
		loaded.GetPlayers() != 2 {
		t.Errorf("loaded settings %+v", loaded)
	}
}

// Пустое хранилище и отсутствующие ключи — значения по умолчанию
func TestSettingsRepository_Defaults(t *testing.T) {
	storage := &memoryStorage{data: map[string][]byte{}}
	repository := NewSettingsRepository(storage)

	settings, err := repository.GetSettings()
	if err != nil || *settings != *types.NewSettingsEntity() {
		t.Errorf("empty storage: %v, settings %+v", err, settings)
	}

	storage.data[settingsKey] = []byte(`{"version":1,"volume":3}`)
	settings, err = repository.GetSettings()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !settings.IsEffectsEnabled() || !settings.IsFullscreen() ||
		settings.GetVolumeLevel() != 3 ||
		settings.GetLanguage() != types.LanguageAuto {
		t.Errorf("partial settings %+v", settings)
	}
}

// Повреждённое или чужое сохранение — ошибка, но игра получает дефолты
func TestSettingsRepository_Broken(t *testing.T) {
	for _, data := range []string{`{broken`, `{"version":99}`} {
		storage := &memoryStorage{data: map[string][]byte{
			settingsKey: []byte(data),
		}}
		settings, err := NewSettingsRepository(storage).GetSettings()
		if err == nil {
			t.Errorf("%s: ожидалась ошибка", data)
		}
		if *settings != *types.NewSettingsEntity() {
			t.Errorf("%s: settings %+v, ожидались дефолты", data, settings)
		}
	}
}
