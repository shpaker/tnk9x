package types

// SettingsItem — строка экрана настроек
type SettingsItem int

const (
	SettingsItemGraphics SettingsItem = iota
	SettingsItemFullscreen
	SettingsItemVolume
	// SettingsItemControls — переход к экрану раскладки
	SettingsItemControls
	SettingsItemBack
)

// SettingsRow — строка экрана настроек: пункт и текущее значение;
// у CONTROLS и BACK значения нет
type SettingsRow struct {
	Item  SettingsItem
	Value string
}

// SettingsViewData — экран настроек для отрисовки: плоский DTO,
// рендер показывает значения как есть
type SettingsViewData struct {
	Rows        []SettingsRow
	ActiveIndex int
}
