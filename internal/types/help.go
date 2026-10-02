package types

// HelpRowKind — вид строки страницы «Как играть»
type HelpRowKind int

const (
	// HelpRowAction — действие игрока по его раскладке
	HelpRowAction HelpRowKind = iota
	// HelpRowPause — пауза: зарезервированные Esc и Start
	HelpRowPause
)

// HelpRow — строка таблицы управления: клавиши обоих игроков
// и кнопка геймпада первого игрока
type HelpRow struct {
	Kind   HelpRowKind
	Action InputAction
	Keys   [MaxPlayers]string
	Button string
}

// HelpViewData — страница «Как играть» для отрисовки: на двоих
// у каждого игрока свой столбец клавиатуры, на таче — экранные
// контроллы вместо таблицы
type HelpViewData struct {
	Players     uint
	TouchActive bool
	Rows        []HelpRow
	// Ticks — счётчик кадров для мигания подсказки старта
	Ticks uint
}
