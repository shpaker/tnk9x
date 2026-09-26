package types

// PauseMenuItem — пункт меню паузы
type PauseMenuItem int

const (
	PauseMenuItemContinue PauseMenuItem = iota
	PauseMenuItemGraphics
	PauseMenuItemExitToSelect
)

// PauseMenuViewData — состояние меню паузы для отрисовки:
// плоский DTO, рендер показывает пункты как есть
type PauseMenuViewData struct {
	Items       []PauseMenuItem
	ActiveIndex int
	// EffectsEnabled — подпись пункта графики: RTX или классика
	EffectsEnabled bool
}
