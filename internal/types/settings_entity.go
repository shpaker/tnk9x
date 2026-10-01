package types

// MaxVolumeLevel — число шагов громкости: шкала 0..100% шагом 10%
const MaxVolumeLevel = 10

// defaultVolumeLevel — громкость первого запуска, 50%
const defaultVolumeLevel = 5

// MaxPlayers — игроков в режиме на двоих
const MaxPlayers = 2

// SettingsEntity — пользовательские настройки: режим графики,
// полный экран, громкость и число игроков; одна на приложение,
// сохраняется между запусками
type SettingsEntity struct {
	effects     bool
	fullscreen  bool
	volumeLevel int
	players     uint
}

// NewSettingsEntity — настройки первого запуска: обычная графика
// с эффектами, полный экран, громкость 50%, один игрок
func NewSettingsEntity() *SettingsEntity {
	return &SettingsEntity{
		effects:     true,
		fullscreen:  true,
		volumeLevel: defaultVolumeLevel,
		players:     1,
	}
}

// GetPlayers — число игроков, 1 или 2
func (s *SettingsEntity) GetPlayers() uint {
	return s.players
}

// SetPlayers задаёт число игроков, ограничивая его 1..MaxPlayers
func (s *SettingsEntity) SetPlayers(players uint) {
	s.players = max(1, min(players, MaxPlayers))
}

// IsEffectsEnabled — обычная графика с эффектами (свет, bloom, CRT);
// иначе классическая
func (s *SettingsEntity) IsEffectsEnabled() bool {
	return s.effects
}

func (s *SettingsEntity) SetEffectsEnabled(enabled bool) {
	s.effects = enabled
}

// IsFullscreen — запуск на весь экран; при старте действует
// только на десктопе
func (s *SettingsEntity) IsFullscreen() bool {
	return s.fullscreen
}

func (s *SettingsEntity) SetFullscreen(fullscreen bool) {
	s.fullscreen = fullscreen
}

// GetVolumeLevel — шаг громкости 0..MaxVolumeLevel
func (s *SettingsEntity) GetVolumeLevel() int {
	return s.volumeLevel
}

// SetVolumeLevel задаёт шаг громкости, ограничивая его шкалой
func (s *SettingsEntity) SetVolumeLevel(level int) {
	s.volumeLevel = max(0, min(level, MaxVolumeLevel))
}

// GetVolume — громкость для звукового движка, 0..1
func (s *SettingsEntity) GetVolume() float64 {
	return float64(s.volumeLevel) / MaxVolumeLevel
}
