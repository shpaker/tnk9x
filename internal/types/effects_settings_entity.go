package types

// EffectsSettingsEntity — включённость графических эффектов (свет, bloom,
// CRT); одна на приложение, переключается горячей клавишей
type EffectsSettingsEntity struct {
	enabled bool
}

func NewEffectsSettingsEntity(enabled bool) *EffectsSettingsEntity {
	return &EffectsSettingsEntity{enabled: enabled}
}

func (s *EffectsSettingsEntity) IsEnabled() bool {
	return s.enabled
}

func (s *EffectsSettingsEntity) Toggle() {
	s.enabled = !s.enabled
}

// GraphicsLabel — подпись строки режима графики в меню: общая для меню
// выбора уровня и паузы
func GraphicsLabel(effectsEnabled bool) string {
	if effectsEnabled {
		return "GRAPHICS RTX"
	}
	return "GRAPHICS CLASSIC"
}
