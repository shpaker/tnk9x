package window

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

var _ interfaces.IWindowAdapter = (*WindowAdapter)(nil)

// WindowAdapter — окно приложения поверх ebiten; в браузере
// полный экран включается только по жесту пользователя
// (нажатие клавиши или тап)
type WindowAdapter struct {
	// cursorHidden — курсор спрятан: режим меняется только при смене
	cursorHidden bool
}

func NewWindowAdapter() *WindowAdapter {
	return &WindowAdapter{}
}

func (a *WindowAdapter) IsFullscreen() bool {
	return ebiten.IsFullscreen()
}

func (a *WindowAdapter) SetFullscreen(fullscreen bool) {
	ebiten.SetFullscreen(fullscreen)
}

func (a *WindowAdapter) SetCursorVisible(visible bool) {
	if a.cursorHidden == !visible {
		return
	}
	a.cursorHidden = !visible
	mode := ebiten.CursorModeVisible
	if a.cursorHidden {
		mode = ebiten.CursorModeHidden
	}
	ebiten.SetCursorMode(mode)
}
