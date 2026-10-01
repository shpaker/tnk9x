package types

// PauseMenuItem — пункт меню паузы
type PauseMenuItem int

const (
	PauseMenuItemContinue PauseMenuItem = iota
	// PauseMenuItemRestart — уровень заново с начальными жизнями
	PauseMenuItemRestart
	PauseMenuItemSettings
	PauseMenuItemExitToLevels
)

// PauseMenuViewData — состояние меню паузы для отрисовки:
// плоский DTO, рендер показывает пункты как есть
type PauseMenuViewData struct {
	Items       []PauseMenuItem
	ActiveIndex int
}

// StageResultItem — пункт меню итогов уровня
type StageResultItem int

const (
	StageResultItemNext StageResultItem = iota
	// StageResultItemContinue — следующий уровень с переносом жизней
	// и прокачки танков
	StageResultItemContinue
	StageResultItemRetry
	StageResultItemLevels
)

// StageResultViewData — экран итогов уровня для отрисовки
type StageResultViewData struct {
	Won          bool
	Stars        uint
	ElapsedTicks uint
	LivesLost    uint
	// CarriedOver — уровень начат с переносом: звёзд не больше двух
	CarriedOver bool
	// NewBest — результат лучше прежнего
	NewBest     bool
	Items       []StageResultItem
	ActiveIndex int
	// Reveal — насколько экран уже появился
	Reveal StageResultReveal
}

// StageResultReveal — поэтапное появление экрана итогов: доли 0..1
// проявления подложки, заголовка и статистики, число загоревшихся
// звёзд и показано ли меню
type StageResultReveal struct {
	Backdrop float64
	Title    float64
	Stars    uint
	Stats    float64
	Menu     bool
}
