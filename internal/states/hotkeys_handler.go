package states

import (
	"log"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// captureStatus — идёт ли назначение клавиши; контракт определён
// у потребителя
type captureStatus interface {
	IsCapturing() bool
}

// HotkeysHandler — реакция на глобальные хоткеи на любом экране:
// графика NORMAL/CLASSIC и полный экран. Пока на экране раскладки
// ждут нажатия, хоткеи не срабатывают — их клавишу назначают
type HotkeysHandler struct {
	// Use Cases
	settingsUseCases interfaces.ISettingsUseCases
	// Adapters
	hotkeys interfaces.IHotkeysAdapter
	window  interfaces.IWindowAdapter
	// Presentation
	capture captureStatus
	// Entities
	settings *types.SettingsEntity
}

func NewHotkeysHandler(
	settingsUseCases interfaces.ISettingsUseCases,
	hotkeys interfaces.IHotkeysAdapter,
	window interfaces.IWindowAdapter,
	capture captureStatus,
	settings *types.SettingsEntity,
) *HotkeysHandler {
	return &HotkeysHandler{
		settingsUseCases: settingsUseCases,
		hotkeys:          hotkeys,
		window:           window,
		capture:          capture,
		settings:         settings,
	}
}

// Update применяет нажатые в этом кадре хоткеи; изменения
// сохраняются, ошибка сохранения не мешает игре
func (h *HotkeysHandler) Update() {
	if h.capture.IsCapturing() {
		return
	}

	for _, hotkey := range h.hotkeys.JustPressed() {
		switch hotkey {
		case types.HotkeyGraphics:
			h.change(types.SettingsItemGraphics)
		case types.HotkeyFullscreen:
			// Переключаем относительно окна: из полного экрана могли
			// выйти средствами ОС или браузера
			h.settings.SetFullscreen(h.window.IsFullscreen())
			h.change(types.SettingsItemFullscreen)
			h.window.SetFullscreen(h.settings.IsFullscreen())
		}
	}
}

func (h *HotkeysHandler) change(item types.SettingsItem) {
	if err := h.settingsUseCases.Change(h.settings, item, 1); err != nil {
		log.Printf("save settings: %v", err)
	}
}
