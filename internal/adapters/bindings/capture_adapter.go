package bindings

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

var _ interfaces.IBindingCaptureAdapter = (*CaptureAdapter)(nil)

// CaptureAdapter ловит нажатие для назначения раскладки: любую
// клавишу и любую кнопку любого геймпада со стандартной раскладкой
type CaptureAdapter struct {
	// Переиспользуемые буферы опроса
	keys    []ebiten.Key
	pads    []ebiten.GamepadID
	buttons []ebiten.StandardGamepadButton
}

func NewCaptureAdapter() *CaptureAdapter {
	return &CaptureAdapter{}
}

func (a *CaptureAdapter) JustPressedKey() (string, bool) {
	a.keys = inpututil.AppendJustPressedKeys(a.keys[:0])
	for _, key := range a.keys {
		if name := key.String(); name != "" {
			return name, true
		}
	}
	return "", false
}

func (a *CaptureAdapter) JustPressedPadButton() (string, bool) {
	a.pads = StandardGamepads(a.pads[:0])
	for _, id := range a.pads {
		a.buttons = inpututil.AppendJustPressedStandardGamepadButtons(
			id, a.buttons[:0],
		)
		for _, button := range a.buttons {
			if name, ok := ButtonName(button); ok {
				return name, true
			}
		}
	}
	return "", false
}
