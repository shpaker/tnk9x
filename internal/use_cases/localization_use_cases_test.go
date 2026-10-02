package use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// testLanguageConfig — языки игры: русский для СНГ, иначе английский
var testLanguageConfig = types.LanguageConfig{
	Languages: []types.Language{"en", "ru", "tr"},
	Default:   "en",
	Fallbacks: map[types.Language]types.Language{
		"be": "ru", "kk": "ru", "uk": "ru", "uz": "ru",
	},
}

// Язык площадки сводится к поддерживаемому
func TestLocalization_PlatformLanguage(t *testing.T) {
	cases := map[string]types.Language{
		"ru":          "ru",
		"ru-RU":       "ru",
		"ru_RU.UTF-8": "ru",
		"RU":          "ru",
		"be":          "ru",
		"kk":          "ru",
		"uk-UA":       "ru",
		"uz":          "ru",
		"tr":          "tr",
		"tr-TR":       "tr",
		"en-US":       "en",
		"de":          "en",
		"ja":          "en",
		"C":           "en",
		"":            "en",
	}
	for tag, want := range cases {
		localization := use_cases.NewLocalizationUseCases(
			&testutil.FakeTexts{}, testLanguageConfig, tag,
		)
		got := localization.ResolveLanguage(types.NewSettingsEntity())
		if got != want {
			t.Errorf("%q: язык %q, ожидался %q", tag, got, want)
		}
	}
}

// Выбор в настройках главнее языка площадки; выбор, которого больше
// нет среди языков, — AUTO
func TestLocalization_SelectedLanguage(t *testing.T) {
	texts := &testutil.FakeTexts{Languages: testLanguageConfig.Languages}
	localization := use_cases.NewLocalizationUseCases(
		texts, testLanguageConfig, "ru-RU",
	)
	settings := types.NewSettingsEntity()

	settings.SetLanguage("tr")
	if err := localization.Apply(settings); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if texts.GetLanguage() != "tr" {
		t.Errorf("активный язык %q, ожидался tr", texts.GetLanguage())
	}

	settings.SetLanguage("de")
	if err := localization.Apply(settings); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if texts.GetLanguage() != "ru" {
		t.Errorf(
			"активный язык %q, ожидался язык площадки ru",
			texts.GetLanguage(),
		)
	}
}
