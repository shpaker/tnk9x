package types

import "errors"

// InputAction — действие игрока, назначаемое на клавишу и кнопку
// геймпада
type InputAction int

const (
	InputActionUp InputAction = iota
	InputActionDown
	InputActionLeft
	InputActionRight
	InputActionFire
	InputActionsCount
)

// HotkeyAction — глобальный хоткей, работает на всех экранах
type HotkeyAction int

const (
	HotkeyGraphics HotkeyAction = iota
	HotkeyFullscreen
	HotkeysCount
)

// ControlsPage — страница экрана раскладки
type ControlsPage int

const (
	ControlsPagePlayer1 ControlsPage = iota
	ControlsPagePlayer2
	ControlsPageHotkeys
	ControlsPagesCount
)

// Player — игрок страницы; для страницы хоткеев не определён
func (p ControlsPage) Player() (PlayerTankNum, bool) {
	switch p {
	case ControlsPagePlayer1:
		return PlayerTankNumPlayer1, true
	case ControlsPagePlayer2:
		return PlayerTankNumPlayer2, true
	default:
		return 0, false
	}
}

// ControlsDevice — устройство ячейки раскладки
type ControlsDevice int

const (
	ControlsDeviceKeyboard ControlsDevice = iota
	ControlsDeviceGamepad
)

// ControlsSlot — назначаемая ячейка раскладки: действие игрока
// страницы (клавиша или кнопка) или хоткей (только клавиша)
type ControlsSlot struct {
	Page   ControlsPage
	Action InputAction
	Hotkey HotkeyAction
	Device ControlsDevice
}

// ErrBindingTaken — клавиша или кнопка уже занята другим действием
// или зарезервирована
var ErrBindingTaken = errors.New("binding is taken")

// Зарезервированные имена: Esc и Start — пауза и выход из меню,
// их нельзя назначить
const (
	ReservedKey    = "Escape"
	ReservedButton = "START"
)

// ControlsRowKind — вид строки экрана раскладки
type ControlsRowKind int

const (
	// ControlsRowPage — заголовок: листание страниц
	ControlsRowPage ControlsRowKind = iota
	// ControlsRowAction — действие игрока: клавиша и кнопка геймпада
	ControlsRowAction
	// ControlsRowHotkey — хоткей: только клавиша
	ControlsRowHotkey
	// ControlsRowReset — сброс всей раскладки к умолчаниям
	ControlsRowReset
	ControlsRowBack
)

// ControlsRow — строка экрана раскладки; Key и Button — текущие
// имена назначенных клавиши и кнопки
type ControlsRow struct {
	Kind   ControlsRowKind
	Action InputAction
	Hotkey HotkeyAction
	Key    string
	Button string
}

// ControlsCursor — положение на экране раскладки: страница, строка,
// колонка и режим ожидания нажатия
type ControlsCursor struct {
	Page   ControlsPage
	Row    int
	Column ControlsDevice
	// Capturing — ждём клавишу или кнопку для выбранной ячейки
	Capturing bool
	// Rejected — последнее нажатие отклонено: имя занято
	Rejected bool
	// Ticks — счётчик кадров для мигания ожидания
	Ticks uint
}

// ControlsViewData — экран раскладки для отрисовки
type ControlsViewData struct {
	Cursor ControlsCursor
	Rows   []ControlsRow
}
