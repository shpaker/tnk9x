package texts

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/shpaker/tnk9x/internal/types"
)

// memoryTexts — локали в памяти: язык -> содержимое файла
type memoryTexts map[types.Language]string

func (m memoryTexts) GetLocale(lang types.Language) ([]byte, error) {
	data, ok := m[lang]
	if !ok {
		return nil, fmt.Errorf("no locale %q", lang)
	}
	return []byte(data), nil
}

const testEnglish = `
language:
  name: "ENGLISH"
menu:
  play: "PLAY"
  only_english: "ONLY ENGLISH"
  press: "PRESS {{.Button}}"
  stars:
    one: "{{.Count}} STAR"
    other: "{{.Count}} STARS"
`

const testRussian = `
language:
  name: "РУССКИЙ"
menu:
  play: "ИГРАТЬ"
  press: "НАЖМИ {{.Button}}"
  stars:
    one: "{{.Count}} ЗВЕЗДА"
    few: "{{.Count}} ЗВЕЗДЫ"
    many: "{{.Count}} ЗВЁЗД"
    other: "{{.Count}} ЗВЕЗДЫ"
`

func newTestAdapter(t *testing.T) *TextsAdapter {
	t.Helper()
	adapter, err := NewTextsAdapter(
		memoryTexts{"en": testEnglish, "ru": testRussian},
		types.LanguageConfig{
			Languages: []types.Language{"en", "ru"},
			Default:   "en",
		},
	)
	if err != nil {
		t.Fatalf("NewTextsAdapter: %v", err)
	}
	return adapter
}

func TestTextsAdapter_Languages(t *testing.T) {
	adapter := newTestAdapter(t)

	if adapter.GetLanguage() != "en" {
		t.Errorf("initial language %q, want en", adapter.GetLanguage())
	}
	if err := adapter.SetLanguage("tr"); err == nil {
		t.Error("unsupported language must be refused")
	}
	if err := adapter.SetLanguage("ru"); err != nil {
		t.Fatalf("SetLanguage: %v", err)
	}
	if got := adapter.Get("menu.play"); got != "ИГРАТЬ" {
		t.Errorf("Get %q, want ИГРАТЬ", got)
	}
	if got := adapter.GetLanguageName("en"); got != "ENGLISH" {
		t.Errorf("name of en %q, want ENGLISH", got)
	}
	if got := adapter.GetLanguageName("ru"); got != "РУССКИЙ" {
		t.Errorf("name of ru %q, want РУССКИЙ", got)
	}
}

// Нет перевода — текст языка по умолчанию, нет нигде — сам ключ
func TestTextsAdapter_Fallbacks(t *testing.T) {
	adapter := newTestAdapter(t)
	if err := adapter.SetLanguage("ru"); err != nil {
		t.Fatalf("SetLanguage: %v", err)
	}

	if got := adapter.Get("menu.only_english"); got != "ONLY ENGLISH" {
		t.Errorf("default language fallback %q", got)
	}
	if got := adapter.Get("menu.missing"); got != "menu.missing" {
		t.Errorf("missing key %q, want the key itself", got)
	}
	if got := adapter.GetOr("menu.missing", "FROM MAP"); got != "FROM MAP" {
		t.Errorf("GetOr %q, want FROM MAP", got)
	}
	if got := adapter.GetOr("menu.play", "FROM MAP"); got != "ИГРАТЬ" {
		t.Errorf("GetOr %q, want ИГРАТЬ", got)
	}
}

func TestTextsAdapter_FormatAndPlural(t *testing.T) {
	adapter := newTestAdapter(t)

	press := adapter.Format("menu.press", types.TextArgs{"Button": "ENTER"})
	if press != "PRESS ENTER" {
		t.Errorf("Format %q", press)
	}

	english := map[int]string{1: "1 STAR", 2: "2 STARS", 5: "5 STARS"}
	for count, want := range english {
		if got := adapter.Plural("menu.stars", count, nil); got != want {
			t.Errorf("en Plural(%d) %q, want %q", count, got, want)
		}
	}

	if err := adapter.SetLanguage("ru"); err != nil {
		t.Fatalf("SetLanguage: %v", err)
	}
	russian := map[int]string{
		1: "1 ЗВЕЗДА", 2: "2 ЗВЕЗДЫ", 5: "5 ЗВЁЗД", 21: "21 ЗВЕЗДА",
	}
	for count, want := range russian {
		if got := adapter.Plural("menu.stars", count, nil); got != want {
			t.Errorf("ru Plural(%d) %q, want %q", count, got, want)
		}
	}
}

// localesDir — файлы локалей игры
const localesDir = "../../../assets/locales"

// placeholderPattern — параметр шаблона {{.Name}}
var placeholderPattern = regexp.MustCompile(`{{\s*\.(\w+)\s*}}`)

// Каждая локаль игры: все ключи интерфейса переведены в самом языке,
// набор ключей и параметров шаблонов совпадает с английским
func TestLocales_Complete(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(localesDir, "*.yml"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no locales in %s: %v", localesDir, err)
	}

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	locales := make(map[string]map[string]string)
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		file, err := bundle.ParseMessageFileBytes(data, filepath.Base(path))
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		placeholders := make(map[string]string)
		for _, message := range file.Messages {
			placeholders[message.ID] = messagePlaceholders(message)
		}
		locales[strings.TrimSuffix(filepath.Base(path), ".yml")] = placeholders
	}

	english, ok := locales["en"]
	if !ok {
		t.Fatal("no en.yml")
	}
	for _, key := range types.TextKeys {
		if _, ok := english[string(key)]; !ok {
			t.Errorf("en: no %s", key)
		}
	}
	for lang, messages := range locales {
		for id, want := range english {
			got, ok := messages[id]
			if !ok {
				t.Errorf("%s: no %s", lang, id)
				continue
			}
			if got != want {
				t.Errorf("%s: %s placeholders %q, want %q", lang, id, got, want)
			}
		}
		for id := range messages {
			if _, ok := english[id]; !ok {
				t.Errorf("%s: unknown key %s", lang, id)
			}
		}
	}
}

// Каждой карте и пачке кампании — перевод названия
func TestLocales_CampaignNames(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(localesDir, "en.yml"))
	if err != nil {
		t.Fatalf("read en.yml: %v", err)
	}
	maps, err := filepath.Glob("../../../assets/levels/*.tnkmap")
	if err != nil || len(maps) == 0 {
		t.Fatalf("no maps: %v", err)
	}
	var locale struct {
		Levels map[string]string `yaml:"levels"`
		Packs  map[string]string `yaml:"packs"`
	}
	if err := yaml.Unmarshal(data, &locale); err != nil {
		t.Fatalf("parse en.yml: %v", err)
	}
	for number := 1; number <= len(maps); number++ {
		key := strings.TrimPrefix(
			string(types.LevelNameTextKey(number)),
			"levels.",
		)
		if _, ok := locale.Levels[key]; !ok {
			t.Errorf("no name of level %d", number)
		}
	}
	if len(locale.Packs) == 0 {
		t.Error("no pack names")
	}
}

// messagePlaceholders — параметры всех форм сообщения через запятую
func messagePlaceholders(message *i18n.Message) string {
	set := make(map[string]bool)
	forms := []string{
		message.Zero, message.One, message.Two,
		message.Few, message.Many, message.Other,
	}
	for _, form := range forms {
		for _, match := range placeholderPattern.FindAllStringSubmatch(form, -1) {
			set[match[1]] = true
		}
	}
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ",")
}
