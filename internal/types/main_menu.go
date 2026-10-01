package types

// MainMenuItem — пункт главного меню
type MainMenuItem int

const (
	// MainMenuItemOnePlayer и MainMenuItemTwoPlayers — режим игры:
	// у каждого своё прохождение кампании
	MainMenuItemOnePlayer MainMenuItem = iota
	MainMenuItemTwoPlayers
	MainMenuItemSettings
	// MainMenuItemQuit — выход из игры, только на десктопе
	MainMenuItemQuit
)

// MainMenuViewData — главное меню для отрисовки: пункты, версия игры
// и затемнение появления (0 — нет, 1 — полностью чёрный)
type MainMenuViewData struct {
	Items       []MainMenuItem
	ActiveIndex int
	Version     string
	Fade        float64
}
