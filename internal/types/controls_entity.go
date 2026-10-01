package types

// playersCount — игроков с собственной раскладкой
const playersCount = 2

// ControlsEntity — раскладка управления: клавиши и кнопки геймпада
// игроков и клавиши хоткеев; одна на приложение. Имена текстовые —
// формат хранения: клавиши в виде имён ebiten.Key ("W", "Quote"),
// кнопки — короткие имена стандартной раскладки ("A", "D-UP");
// разбирает их только адаптер ввода
type ControlsEntity struct {
	keys    [playersCount][InputActionsCount]string
	buttons [playersCount][InputActionsCount]string
	hotkeys [HotkeysCount]string
}

// NewControlsEntity — раскладка по умолчанию: P1 — WASD и G,
// P2 — IJKL и апостроф, геймпады — крестовина и A, хоткеи — F2
// (графика) и F11 (полный экран)
func NewControlsEntity() *ControlsEntity {
	controls := &ControlsEntity{}
	controls.Reset()
	return controls
}

// Reset возвращает всю раскладку к умолчаниям
func (c *ControlsEntity) Reset() {
	c.keys = [playersCount][InputActionsCount]string{
		{"W", "S", "A", "D", "G"},
		{"I", "K", "J", "L", "Quote"},
	}
	padDefaults := [InputActionsCount]string{
		"D-UP", "D-DOWN", "D-LEFT", "D-RIGHT", "A",
	}
	c.buttons = [playersCount][InputActionsCount]string{
		padDefaults, padDefaults,
	}
	c.hotkeys = [HotkeysCount]string{"F2", "F11"}
}

// GetKey — клавиша действия игрока
func (c *ControlsEntity) GetKey(
	player PlayerTankNum,
	action InputAction,
) string {
	if !validPlayer(player) {
		return ""
	}
	return c.keys[player][action]
}

// GetButton — кнопка геймпада действия игрока
func (c *ControlsEntity) GetButton(
	player PlayerTankNum,
	action InputAction,
) string {
	if !validPlayer(player) {
		return ""
	}
	return c.buttons[player][action]
}

// GetHotkey — клавиша хоткея
func (c *ControlsEntity) GetHotkey(hotkey HotkeyAction) string {
	return c.hotkeys[hotkey]
}

// GetName — имя, назначенное ячейке
func (c *ControlsEntity) GetName(slot ControlsSlot) string {
	player, isPlayer := slot.Page.Player()
	switch {
	case !isPlayer:
		return c.hotkeys[slot.Hotkey]
	case slot.Device == ControlsDeviceGamepad:
		return c.buttons[player][slot.Action]
	default:
		return c.keys[player][slot.Action]
	}
}

// SetName назначает имя ячейке; проверки занятости — в use cases
func (c *ControlsEntity) SetName(slot ControlsSlot, name string) {
	player, isPlayer := slot.Page.Player()
	switch {
	case !isPlayer:
		c.hotkeys[slot.Hotkey] = name
	case slot.Device == ControlsDeviceGamepad:
		c.buttons[player][slot.Action] = name
	default:
		c.keys[player][slot.Action] = name
	}
}

// KeySlots — все ячейки клавиатуры: действия обоих игроков и хоткеи
func (c *ControlsEntity) KeySlots() []ControlsSlot {
	slots := make([]ControlsSlot, 0, playersCount*int(InputActionsCount)+
		int(HotkeysCount))
	for _, page := range []ControlsPage{
		ControlsPagePlayer1, ControlsPagePlayer2,
	} {
		for action := range InputActionsCount {
			slots = append(slots, ControlsSlot{Page: page, Action: action})
		}
	}
	for hotkey := range HotkeysCount {
		slots = append(slots, ControlsSlot{
			Page: ControlsPageHotkeys, Hotkey: hotkey,
		})
	}
	return slots
}

func validPlayer(player PlayerTankNum) bool {
	return player >= 0 && int(player) < playersCount
}
