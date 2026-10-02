package use_cases

import (
	"strings"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ILocalizationUseCases = (*LocalizationUseCases)(nil)

// LocalizationUseCases — выбор языка интерфейса: язык из настроек,
// иначе язык площадки, сведённый к поддерживаемому
type LocalizationUseCases struct {
	// Adapters
	texts interfaces.ITextsAdapter
	// Конфигурация
	config types.LanguageConfig
	// platformLanguage — язык площадки при запуске: определяется
	// один раз, до первого кадра
	platformLanguage types.Language
}

// NewLocalizationUseCases: platformTag — язык площадки как его
// отдаёт площадка ("ru-RU", "ru_RU.UTF-8")
func NewLocalizationUseCases(
	texts interfaces.ITextsAdapter,
	config types.LanguageConfig,
	platformTag string,
) *LocalizationUseCases {
	return &LocalizationUseCases{
		texts:            texts,
		config:           config,
		platformLanguage: supportedLanguage(config, platformTag),
	}
}

// ResolveLanguage реализует ILocalizationUseCases: выбор вручную
// главнее языка площадки; выбор, которого больше нет среди языков,
// считается AUTO
func (uc *LocalizationUseCases) ResolveLanguage(
	settings *types.SettingsEntity,
) types.Language {
	if selected := settings.GetLanguage(); uc.config.IsSupported(selected) {
		return selected
	}
	return uc.platformLanguage
}

// Apply реализует ILocalizationUseCases
func (uc *LocalizationUseCases) Apply(settings *types.SettingsEntity) error {
	return uc.texts.SetLanguage(uc.ResolveLanguage(settings))
}

// supportedLanguage сводит язык площадки к поддерживаемому: основной
// подтег ("ru" из "ru-RU" и "ru_RU.UTF-8"), затем замена из Fallbacks
// (be, kk, uk, uz -> ru), иначе язык по умолчанию
func supportedLanguage(config types.LanguageConfig, tag string) types.Language {
	primary, _, _ := strings.Cut(strings.ToLower(tag), "-")
	primary, _, _ = strings.Cut(primary, "_")
	primary, _, _ = strings.Cut(primary, ".")
	primary, _, _ = strings.Cut(primary, "@")
	language := types.Language(strings.TrimSpace(primary))

	if config.IsSupported(language) {
		return language
	}
	if fallback, ok := config.Fallbacks[language]; ok &&
		config.IsSupported(fallback) {
		return fallback
	}
	return config.Default
}
