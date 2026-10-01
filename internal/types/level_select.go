package types

// LevelSelectorEntity — курсор экрана выбора уровня: пачка
// и позиция уровня в ней
type LevelSelectorEntity struct {
	PackIndex int
	Position  int
}

// LevelSelectEntry — ячейка уровня в строке пачки
type LevelSelectEntry struct {
	Number   int
	Unlocked bool
	Stars    uint
}

// LevelSelectViewData — состояние экрана выбора уровня для отрисовки:
// плоский DTO, рендер показывает значения как есть
type LevelSelectViewData struct {
	PackName   string
	PackIndex  int
	PacksCount int
	Pack       PackStatus
	// CampaignMaxStars — максимум звёзд кампании для общего счётчика
	CampaignMaxStars uint

	Entries        []LevelSelectEntry
	ActivePosition int

	// Выбранный уровень: превью карты и сценарий врагов
	Level          *LevelEntity
	LevelUnlocked  bool
	LevelStars     uint
	EnemyCounts    [EnemyLevelsCount]uint
	Time3StarTicks uint

	// TouchActive — управление с экрана: подсказки меняются
	// с клавиатурных на тач-варианты
	TouchActive bool
	// PlayerCount — режим игры для строки запуска
	PlayerCount uint
}
