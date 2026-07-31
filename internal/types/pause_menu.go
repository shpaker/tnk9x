package types

// PauseMenuItem — пункт меню паузы
type PauseMenuItem int

const (
	PauseMenuItemContinue PauseMenuItem = iota
	PauseMenuItemExitToSelect
)

// PauseMenuViewData — состояние меню паузы для отрисовки:
// плоский DTO, рендер показывает пункты как есть
type PauseMenuViewData struct {
	Items       []PauseMenuItem
	ActiveIndex int
}
