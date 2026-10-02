package types

// Language — язык интерфейса: код ISO 639-1 ("en", "ru", "tr")
type Language string

// LanguageAuto — язык не выбран вручную: берётся язык площадки
const LanguageAuto Language = ""

// LanguageConfig — поддерживаемые языки и правило выбора языка
// по языку площадки; иммутабельная конфигурация игры
type LanguageConfig struct {
	// Languages — языки с файлами локалей; порядок — для пункта
	// LANGUAGE экрана настроек
	Languages []Language
	// Default — язык, если язык площадки не поддерживается
	Default Language
	// Fallbacks — замена неподдерживаемого языка площадки
	// поддерживаемым (be, kk, uk, uz -> ru)
	Fallbacks map[Language]Language
}

// IsSupported — для языка есть файл локали
func (c LanguageConfig) IsSupported(language Language) bool {
	for _, supported := range c.Languages {
		if supported == language {
			return true
		}
	}
	return false
}
