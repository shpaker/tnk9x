package app

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

// Языки конфигурации: без списка — английский, язык по умолчанию
// и замены — только из списка
func TestParseLanguageConfig(t *testing.T) {
	config, err := parseLanguageConfig(appConfigSchema{})
	if err != nil || config.Default != "en" || !config.IsSupported("en") {
		t.Errorf("empty config: %+v, %v", config, err)
	}

	config, err = parseLanguageConfig(appConfigSchema{
		Languages:         []string{"en", "ru"},
		LanguageFallbacks: map[string]string{"kk": "ru"},
	})
	if err != nil || config.Default != "en" ||
		config.Fallbacks["kk"] != types.Language("ru") {
		t.Errorf("config: %+v, %v", config, err)
	}

	invalid := []appConfigSchema{
		{Languages: []string{"en"}, DefaultLanguage: "ru"},
		{
			Languages:         []string{"en"},
			LanguageFallbacks: map[string]string{"kk": "ru"},
		},
	}
	for _, schema := range invalid {
		if _, err := parseLanguageConfig(schema); err == nil {
			t.Errorf("%+v must be refused", schema)
		}
	}
}

// Файл локали есть у каждого языка конфигурации игры
func TestLoadConfig_Languages(t *testing.T) {
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	for _, language := range cfg.Languages.Languages {
		if _, err := assetsFS().Open("locales/" + string(language) + ".yml"); err != nil {
			t.Errorf("no locale for %q: %v", language, err)
		}
	}
}
