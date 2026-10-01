package bindings

import (
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IHotkeysAdapter = (*HotkeysAdapter)(nil)

// HotkeysAdapter — нажатые хоткеи по текущей раскладке: клавиши
// читаются каждый кадр, переназначение действует сразу
type HotkeysAdapter struct {
	// Entities
	controls *types.ControlsEntity

	pressed []types.HotkeyAction
}

func NewHotkeysAdapter(controls *types.ControlsEntity) *HotkeysAdapter {
	return &HotkeysAdapter{
		controls: controls,
	}
}

func (a *HotkeysAdapter) JustPressed() []types.HotkeyAction {
	a.pressed = a.pressed[:0]
	for hotkey := range types.HotkeysCount {
		key, ok := KeyByName(a.controls.GetHotkey(hotkey))
		if ok && inpututil.IsKeyJustPressed(key) {
			a.pressed = append(a.pressed, hotkey)
		}
	}
	return a.pressed
}
