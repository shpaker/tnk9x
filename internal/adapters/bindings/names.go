// Package bindings переводит текстовые имена раскладки в клавиши
// и кнопки ebiten и обратно, находит геймпады игроков и ловит
// нажатия для назначения раскладки.
package bindings

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/types"
)

// keysByName — клавиши ebiten по их текстовым именам ("W", "Quote")
var keysByName = func() map[string]ebiten.Key {
	keys := make(map[string]ebiten.Key)
	for key := ebiten.Key(0); key <= ebiten.KeyMax; key++ {
		name := key.String()
		if _, exists := keys[name]; name != "" && !exists {
			keys[name] = key
		}
	}
	return keys
}()

// buttonsByName — кнопки стандартной раскладки геймпада по коротким
// именам, привычным по Xbox-геймпадам
var buttonsByName = map[string]ebiten.StandardGamepadButton{
	"A":       ebiten.StandardGamepadButtonRightBottom,
	"B":       ebiten.StandardGamepadButtonRightRight,
	"X":       ebiten.StandardGamepadButtonRightLeft,
	"Y":       ebiten.StandardGamepadButtonRightTop,
	"LB":      ebiten.StandardGamepadButtonFrontTopLeft,
	"RB":      ebiten.StandardGamepadButtonFrontTopRight,
	"LT":      ebiten.StandardGamepadButtonFrontBottomLeft,
	"RT":      ebiten.StandardGamepadButtonFrontBottomRight,
	"SELECT":  ebiten.StandardGamepadButtonCenterLeft,
	"START":   ebiten.StandardGamepadButtonCenterRight,
	"HOME":    ebiten.StandardGamepadButtonCenterCenter,
	"L3":      ebiten.StandardGamepadButtonLeftStick,
	"R3":      ebiten.StandardGamepadButtonRightStick,
	"D-UP":    ebiten.StandardGamepadButtonLeftTop,
	"D-DOWN":  ebiten.StandardGamepadButtonLeftBottom,
	"D-LEFT":  ebiten.StandardGamepadButtonLeftLeft,
	"D-RIGHT": ebiten.StandardGamepadButtonLeftRight,
}

// namesByButton — обратная таблица кнопок
var namesByButton = func() map[ebiten.StandardGamepadButton]string {
	names := make(map[ebiten.StandardGamepadButton]string)
	for name, button := range buttonsByName {
		names[button] = name
	}
	return names
}()

// keyLabels — подписи клавиш, чьи имена длинные или нечитаемые
// на экране раскладки
var keyLabels = map[string]string{
	"Quote":        "'",
	"Backquote":    "`",
	"Semicolon":    ";",
	"Comma":        ",",
	"Period":       ".",
	"Slash":        "/",
	"Backslash":    "\\",
	"BracketLeft":  "[",
	"BracketRight": "]",
	"Minus":        "-",
	"Equal":        "=",
	"ArrowUp":      "UP",
	"ArrowDown":    "DOWN",
	"ArrowLeft":    "LEFT",
	"ArrowRight":   "RIGHT",
	"ShiftLeft":    "LSHIFT",
	"ShiftRight":   "RSHIFT",
	"ControlLeft":  "LCTRL",
	"ControlRight": "RCTRL",
	"AltLeft":      "LALT",
	"AltRight":     "RALT",
	"MetaLeft":     "LMETA",
	"MetaRight":    "RMETA",
	"Backspace":    "BKSP",
	"CapsLock":     "CAPS",
	"PageUp":       "PGUP",
	"PageDown":     "PGDN",
	"Insert":       "INS",
	"Delete":       "DEL",
	"ContextMenu":  "MENU",
	"NumpadEnter":  "NUMENT",
}

// KeyByName — клавиша по имени раскладки
func KeyByName(name string) (ebiten.Key, bool) {
	key, ok := keysByName[name]
	return key, ok
}

// ButtonByName — кнопка геймпада по имени раскладки
func ButtonByName(name string) (ebiten.StandardGamepadButton, bool) {
	button, ok := buttonsByName[name]
	return button, ok
}

// ButtonName — имя раскладки для кнопки геймпада
func ButtonName(button ebiten.StandardGamepadButton) (string, bool) {
	name, ok := namesByButton[button]
	return name, ok
}

// KeyLabel — короткая подпись клавиши для экрана раскладки
func KeyLabel(name string) string {
	if label, ok := keyLabels[name]; ok {
		return label
	}
	if digit, ok := strings.CutPrefix(name, "Digit"); ok {
		return digit
	}
	if numpad, ok := strings.CutPrefix(name, "Numpad"); ok {
		return "NUM" + strings.ToUpper(numpad)
	}
	return strings.ToUpper(name)
}

// RepairUnknown заменяет имена, которых нет у движка (ручная правка
// сохранения, другая версия игры), значениями по умолчанию
func RepairUnknown(controls *types.ControlsEntity) {
	defaults := types.NewControlsEntity()
	for _, slot := range controls.KeySlots() {
		if _, ok := KeyByName(controls.GetName(slot)); !ok {
			controls.SetName(slot, defaults.GetName(slot))
		}
	}
	for _, page := range []types.ControlsPage{
		types.ControlsPagePlayer1, types.ControlsPagePlayer2,
	} {
		for action := range types.InputActionsCount {
			slot := types.ControlsSlot{
				Page:   page,
				Action: action,
				Device: types.ControlsDeviceGamepad,
			}
			if _, ok := ButtonByName(controls.GetName(slot)); !ok {
				controls.SetName(slot, defaults.GetName(slot))
			}
		}
	}
}
