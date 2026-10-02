// Package texts отдаёт тексты интерфейса из файлов локалей
// assets/locales/<язык>.yml; go-i18n живёт только здесь.
package texts

import (
	"errors"
	"fmt"
	"log"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ITextsAdapter = (*TextsAdapter)(nil)

// TextsAdapter — тексты активного языка; один на приложение.
// Все языки конфигурации грузятся при создании: смена языка
// в настройках не читает файлы
type TextsAdapter struct {
	bundle *i18n.Bundle
	// messages — ключи каждого языка: перевод есть в самом языке
	messages map[types.Language]map[types.TextKey]bool
	// names — название каждого языка на нём самом
	names map[types.Language]string

	// Активный язык
	language  types.Language
	localizer *i18n.Localizer
	// cache — тексты без параметров активного языка: рендеры
	// запрашивают их каждый кадр
	cache map[types.TextKey]string
}

// NewTextsAdapter читает файлы локалей всех языков конфигурации
// и делает активным язык по умолчанию; файл с ошибкой останавливает
// запуск
func NewTextsAdapter(
	textsRepository interfaces.ITextsRepository,
	config types.LanguageConfig,
) (*TextsAdapter, error) {
	defaultTag, err := language.Parse(string(config.Default))
	if err != nil {
		return nil, fmt.Errorf("default language %q: %w", config.Default, err)
	}
	bundle := i18n.NewBundle(defaultTag)
	bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)

	adapter := &TextsAdapter{
		bundle:   bundle,
		messages: make(map[types.Language]map[types.TextKey]bool),
		names:    make(map[types.Language]string),
	}
	for _, lang := range config.Languages {
		data, err := textsRepository.GetLocale(lang)
		if err != nil {
			return nil, err
		}
		file, err := bundle.ParseMessageFileBytes(data, string(lang)+".yml")
		if err != nil {
			return nil, fmt.Errorf("locale %q: %w", lang, err)
		}
		keys := make(map[types.TextKey]bool, len(file.Messages))
		for _, message := range file.Messages {
			keys[types.TextKey(message.ID)] = true
		}
		adapter.messages[lang] = keys
	}
	for _, lang := range config.Languages {
		name, err := i18n.NewLocalizer(bundle, string(lang)).
			Localize(&i18n.LocalizeConfig{
				MessageID: string(types.TextLanguageName),
			})
		if err != nil {
			return nil, fmt.Errorf("locale %q: %w", lang, err)
		}
		adapter.names[lang] = name
	}

	if err := adapter.SetLanguage(config.Default); err != nil {
		return nil, err
	}
	return adapter, nil
}

// GetLanguage реализует ITextsAdapter
func (a *TextsAdapter) GetLanguage() types.Language {
	return a.language
}

// SetLanguage реализует ITextsAdapter
func (a *TextsAdapter) SetLanguage(lang types.Language) error {
	if _, ok := a.messages[lang]; !ok {
		return fmt.Errorf("unsupported language %q", lang)
	}
	a.language = lang
	a.localizer = i18n.NewLocalizer(a.bundle, string(lang))
	a.cache = make(map[types.TextKey]string)
	return nil
}

// Get реализует ITextsAdapter
func (a *TextsAdapter) Get(key types.TextKey) string {
	if text, ok := a.cache[key]; ok {
		return text
	}
	text := a.localize(&i18n.LocalizeConfig{MessageID: string(key)})
	a.cache[key] = text
	return text
}

// Format реализует ITextsAdapter
func (a *TextsAdapter) Format(key types.TextKey, args types.TextArgs) string {
	return a.localize(&i18n.LocalizeConfig{
		MessageID:    string(key),
		TemplateData: map[string]any(args),
	})
}

// Plural реализует ITextsAdapter
func (a *TextsAdapter) Plural(
	key types.TextKey,
	count int,
	args types.TextArgs,
) string {
	data := map[string]any{"Count": count}
	for name, value := range args {
		data[name] = value
	}
	return a.localize(&i18n.LocalizeConfig{
		MessageID:    string(key),
		TemplateData: data,
		PluralCount:  count,
	})
}

// GetOr реализует ITextsAdapter: перевод только из активного языка
// или языка по умолчанию, иначе fallback
func (a *TextsAdapter) GetOr(key types.TextKey, fallback string) string {
	text, _ := a.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: string(key),
	})
	if text == "" {
		return fallback
	}
	return text
}

// GetLanguageName реализует ITextsAdapter
func (a *TextsAdapter) GetLanguageName(lang types.Language) string {
	if name, ok := a.names[lang]; ok {
		return name
	}
	return string(lang)
}

// localize — текст активного языка; перевод из языка по умолчанию
// годится молча, а пропуск ключа везде виден на экране самим ключом
func (a *TextsAdapter) localize(config *i18n.LocalizeConfig) string {
	text, err := a.localizer.Localize(config)
	if err != nil && !isNotFound(err) {
		log.Printf("text %q: %v", config.MessageID, err)
	}
	if text == "" {
		return config.MessageID
	}
	return text
}

func isNotFound(err error) bool {
	var notFound *i18n.MessageNotFoundErr
	return errors.As(err, &notFound)
}
