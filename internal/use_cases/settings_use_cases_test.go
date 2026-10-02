package use_cases_test

import (
	"errors"
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// testLanguages — языки конфигурации в тестах
var testLanguages = []types.Language{"en", "ru", "tr"}

// newSettingsUseCases — use cases настроек с фейковыми текстами:
// значение — ключ текста
func newSettingsUseCases(
	repository *recordingSettingsRepository,
	fullscreenAvailable bool,
) *use_cases.SettingsUseCases {
	return use_cases.NewSettingsUseCases(
		repository,
		&testutil.FakeTexts{Languages: testLanguages},
		fullscreenAvailable,
		testLanguages,
	)
}

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
	settingsUseCases := newSettingsUseCases(repository, true)
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
	settingsUseCases := newSettingsUseCases(repository, true)
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
	settingsUseCases := newSettingsUseCases(repository, true)
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
	settingsUseCases := newSettingsUseCases(repository, true)
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
		t.Errorf(
			"без смены режима хранилище не трогается: %d",
			repository.saves,
		)
	}
}

func TestSettings_BuildView(t *testing.T) {
	settingsUseCases := newSettingsUseCases(
		&recordingSettingsRepository{}, true,
	)
	settings := types.NewSettingsEntity()
	settings.SetEffectsEnabled(false)

	view := settingsUseCases.BuildView(settings, 2)

	want := []types.SettingsRow{
		{Item: types.SettingsItemGraphics, Value: "settings.graphics_classic"},
		{Item: types.SettingsItemFullscreen, Value: "settings.on"},
		{Item: types.SettingsItemVolume, Value: "settings.volume_percent 50"},
		{Item: types.SettingsItemLanguage, Value: "settings.language_auto"},
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
	settingsUseCases := newSettingsUseCases(
		&recordingSettingsRepository{}, false,
	)
	for _, item := range settingsUseCases.Items() {
		if item == types.SettingsItemFullscreen {
			t.Fatal("строки FULLSCREEN быть не должно")
		}
	}
}

// Язык идёт по кругу AUTO и языков конфигурации в обе стороны;
// выбранный язык показывается своим названием
func TestSettings_ChangeLanguage(t *testing.T) {
	repository := &recordingSettingsRepository{}
	settingsUseCases := newSettingsUseCases(repository, true)
	settings := types.NewSettingsEntity()

	for _, want := range []types.Language{"en", "ru", "tr", types.LanguageAuto} {
		_ = settingsUseCases.Change(settings, types.SettingsItemLanguage, 1)
		if settings.GetLanguage() != want {
			t.Errorf("язык %q, ожидался %q", settings.GetLanguage(), want)
		}
	}
	_ = settingsUseCases.Change(settings, types.SettingsItemLanguage, -1)
	if settings.GetLanguage() != "tr" {
		t.Errorf("назад от AUTO — %q, ожидался tr", settings.GetLanguage())
	}
	if repository.saves != 5 {
		t.Errorf("сохранений %d, ожидалось 5", repository.saves)
	}

	view := settingsUseCases.BuildView(settings, 0)
	for _, row := range view.Rows {
		if row.Item == types.SettingsItemLanguage && row.Value != "TR" {
			t.Errorf("значение языка %q, ожидалось TR", row.Value)
		}
	}

	// Выбор, которого больше нет среди языков, — AUTO
	settings.SetLanguage("de")
	_ = settingsUseCases.Change(settings, types.SettingsItemLanguage, 1)
	if settings.GetLanguage() != "en" {
		t.Errorf(
			"после неизвестного языка %q, ожидался en",
			settings.GetLanguage(),
		)
	}
}
