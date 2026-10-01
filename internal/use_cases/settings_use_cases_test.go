package use_cases_test

import (
	"errors"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// recordingSettingsRepository считает сохранения настроек
type recordingSettingsRepository struct {
	saves int
	err   error
}

func (r *recordingSettingsRepository) GetSettings() (
	*types.SettingsEntity,
	error,
) {
	return types.NewSettingsEntity(), nil
}

func (r *recordingSettingsRepository) SaveSettings(
	*types.SettingsEntity,
) error {
	r.saves++
	return r.err
}

func TestSettings_ChangeTogglesAndSaves(t *testing.T) {
	repository := &recordingSettingsRepository{}
	settingsUseCases := use_cases.NewSettingsUseCases(repository, true)
	settings := types.NewSettingsEntity()

	_ = settingsUseCases.Change(settings, types.SettingsItemGraphics, 1)
	_ = settingsUseCases.Change(settings, types.SettingsItemFullscreen, -1)

	if settings.IsEffectsEnabled() || settings.IsFullscreen() {
		t.Error("графика и полный экран должны переключиться")
	}
	if repository.saves != 2 {
		t.Errorf("сохранений %d, ожидалось 2", repository.saves)
	}
}

func TestSettings_ChangeVolume(t *testing.T) {
	repository := &recordingSettingsRepository{}
	settingsUseCases := use_cases.NewSettingsUseCases(repository, true)
	settings := types.NewSettingsEntity()

	_ = settingsUseCases.Change(settings, types.SettingsItemVolume, 2)
	if settings.GetVolumeLevel() != 7 {
		t.Errorf("шаг %d, ожидался 7", settings.GetVolumeLevel())
	}

	// Упор в край шкалы ничего не меняет и не пишет в хранилище
	settings.SetVolumeLevel(types.MaxVolumeLevel)
	_ = settingsUseCases.Change(settings, types.SettingsItemVolume, 1)
	if repository.saves != 1 {
		t.Errorf("сохранений %d, ожидалось 1", repository.saves)
	}
}

func TestSettings_ChangeBackAndSaveError(t *testing.T) {
	repository := &recordingSettingsRepository{err: errors.New("disk full")}
	settingsUseCases := use_cases.NewSettingsUseCases(repository, true)
	settings := types.NewSettingsEntity()

	for _, item := range []types.SettingsItem{
		types.SettingsItemControls, types.SettingsItemBack,
	} {
		if err := settingsUseCases.Change(settings, item, 1); err != nil {
			t.Errorf("пункт %d: ошибка %v", item, err)
		}
	}
	if repository.saves != 0 {
		t.Errorf("сохранений %d, ожидалось 0", repository.saves)
	}

	err := settingsUseCases.Change(settings, types.SettingsItemGraphics, 1)
	if err == nil {
		t.Error("ошибка сохранения должна вернуться вызывающему")
	}
}

func TestSettings_SetPlayers(t *testing.T) {
	repository := &recordingSettingsRepository{}
	settingsUseCases := use_cases.NewSettingsUseCases(repository, true)
	settings := types.NewSettingsEntity()

	_ = settingsUseCases.SetPlayers(settings, 2)
	if settings.GetPlayers() != 2 {
		t.Errorf("игроков %d, ожидалось 2", settings.GetPlayers())
	}
	_ = settingsUseCases.SetPlayers(settings, 1)
	if settings.GetPlayers() != 1 {
		t.Errorf("игроков %d, ожидался 1", settings.GetPlayers())
	}
	// Тот же режим и выход за шкалу не пишут лишний раз
	_ = settingsUseCases.SetPlayers(settings, 1)
	_ = settingsUseCases.SetPlayers(settings, 0)
	if settings.GetPlayers() != 1 || repository.saves != 2 {
		t.Errorf("без смены режима хранилище не трогается: %d", repository.saves)
	}
}

func TestSettings_BuildView(t *testing.T) {
	settingsUseCases := use_cases.NewSettingsUseCases(
		&recordingSettingsRepository{}, true,
	)
	settings := types.NewSettingsEntity()
	settings.SetEffectsEnabled(false)

	view := settingsUseCases.BuildView(settings, 2)

	want := []types.SettingsRow{
		{Item: types.SettingsItemGraphics, Value: "CLASSIC"},
		{Item: types.SettingsItemFullscreen, Value: "ON"},
		{Item: types.SettingsItemVolume, Value: "50%"},
		{Item: types.SettingsItemControls},
		{Item: types.SettingsItemBack},
	}
	if len(view.Rows) != len(want) ||
		len(settingsUseCases.Items()) != len(want) {
		t.Fatalf("строк %d, ожидалось %d", len(view.Rows), len(want))
	}
	for i := range want {
		if view.Rows[i] != want[i] {
			t.Errorf("строка %d: %+v, ожидалась %+v", i, view.Rows[i], want[i])
		}
	}
	if view.ActiveIndex != 2 {
		t.Errorf("активная строка %d, ожидалась 2", view.ActiveIndex)
	}
}

// В браузере полного экрана на старте нет — нет и строки
func TestSettings_NoFullscreenItem(t *testing.T) {
	settingsUseCases := use_cases.NewSettingsUseCases(
		&recordingSettingsRepository{}, false,
	)
	for _, item := range settingsUseCases.Items() {
		if item == types.SettingsItemFullscreen {
			t.Fatal("строки FULLSCREEN быть не должно")
		}
	}
}
