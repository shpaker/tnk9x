package types

// PauseMenuItem — пункт меню паузы
type PauseMenuItem int

const (
	PauseMenuItemContinue PauseMenuItem = iota
	PauseMenuItemGraphics
	PauseMenuItemExitToLevels
)

// PauseMenuViewData — состояние меню паузы для отрисовки:
// плоский DTO, рендер показывает пункты как есть
type PauseMenuViewData struct {
	Items       []PauseMenuItem
	ActiveIndex int
	// EffectsEnabled — подпись пункта графики: RTX или классика
	EffectsEnabled bool
}

// StageResultItem — пункт меню итогов уровня
type StageResultItem int

const (
	StageResultItemNext StageResultItem = iota
	StageResultItemRetry
	StageResultItemLevels
)

// StageResultViewData — экран итогов уровня для отрисовки
type StageResultViewData struct {
	Won          bool
	Stars        uint
	ElapsedTicks uint
	LivesLost    uint
	// NewBest — результат лучше прежнего
	NewBest     bool
	Items       []StageResultItem
	ActiveIndex int
}
