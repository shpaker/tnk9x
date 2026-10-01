package states

import (
	"errors"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// controlsRejectTicks — сколько кадров ячейка подсвечена красным
// после попытки назначить занятое имя
const controlsRejectTicks = 30

// ControlsRenderer — контракт рендера экрана раскладки,
// определён у потребителя
type ControlsRenderer interface {
	Draw(screen *ebiten.Image, view types.ControlsViewData)
}

// ControlsOverlay — экран раскладки поверх настроек: страницы
// игроков и хоткеев, назначение клавиш и кнопок геймпада.
// Один на приложение
type ControlsOverlay struct {
	// Use Cases
	controlsUseCases interfaces.IControlsUseCases
	// Adapters
	renderer  ControlsRenderer
	menuInput interfaces.IMenuInputAdapter
	capture   interfaces.IBindingCaptureAdapter
	// Entities
	controls *types.ControlsEntity

	open        bool
	cursor      types.ControlsCursor
	rejectTicks int
}

func NewControlsOverlay(
	controlsUseCases interfaces.IControlsUseCases,
	renderer ControlsRenderer,
	menuInput interfaces.IMenuInputAdapter,
	capture interfaces.IBindingCaptureAdapter,
	controls *types.ControlsEntity,
) *ControlsOverlay {
	return &ControlsOverlay{
		controlsUseCases: controlsUseCases,
		renderer:         renderer,
		menuInput:        menuInput,
		capture:          capture,
		controls:         controls,
	}
}

// Open показывает страницу первого игрока с курсором на заголовке
func (o *ControlsOverlay) Open() {
	o.open = true
	o.cursor = types.ControlsCursor{}
	o.rejectTicks = 0
}

func (o *ControlsOverlay) IsOpen() bool {
	return o.open
}

// IsCapturing — ждём нажатия для назначения: хоткеи в это время
// не срабатывают
func (o *ControlsOverlay) IsCapturing() bool {
	return o.open && o.cursor.Capturing
}

// Update: вверх-вниз — строки, на заголовке влево-вправо — страницы,
// на действии — колонки клавиатуры и геймпада; выбор ячейки включает
// ожидание нажатия
func (o *ControlsOverlay) Update() {
	o.cursor.Ticks++
	if o.rejectTicks > 0 {
		o.rejectTicks--
	}
	o.cursor.Rejected = o.rejectTicks > 0

	if o.cursor.Capturing {
		o.updateCapture()
		return
	}
	if o.menuInput.Back() {
		o.open = false
		return
	}

	rows := o.controlsUseCases.Rows(o.cursor.Page)
	moveUp, moveDown := o.menuInput.Steps()
	o.cursor.Row = stepIndex(o.cursor.Row, len(rows), moveUp, moveDown)
	row := rows[o.cursor.Row]
	step := o.menuInput.SideStep()
	confirmed := o.menuInput.Confirmed()

	switch row.Kind {
	case types.ControlsRowPage:
		o.switchPage(step)
	case types.ControlsRowAction:
		if step < 0 {
			o.cursor.Column = types.ControlsDeviceKeyboard
		} else if step > 0 {
			o.cursor.Column = types.ControlsDeviceGamepad
		}
		o.cursor.Capturing = confirmed
	case types.ControlsRowHotkey:
		o.cursor.Column = types.ControlsDeviceKeyboard
		o.cursor.Capturing = confirmed
	case types.ControlsRowReset:
		if confirmed {
			if err := o.controlsUseCases.ResetAll(o.controls); err != nil {
				log.Printf("save controls: %v", err)
			}
		}
	case types.ControlsRowBack:
		if confirmed {
			o.open = false
		}
	}
}

func (o *ControlsOverlay) Draw(screen *ebiten.Image) {
	o.renderer.Draw(
		screen,
		o.controlsUseCases.BuildView(o.controls, o.cursor),
	)
}

// switchPage листает страницы по кругу; строки страниц разной длины,
// курсор остаётся на заголовке
func (o *ControlsOverlay) switchPage(step int) {
	if step == 0 {
		return
	}
	count := int(types.ControlsPagesCount)
	o.cursor.Page = types.ControlsPage(
		(int(o.cursor.Page) + step + count) % count,
	)
	o.cursor.Column = types.ControlsDeviceKeyboard
}

// updateCapture ждёт клавишу или кнопку для выбранной ячейки:
// Esc, Start или тач-пауза отменяют, занятое имя отклоняется
// с подсветкой, ожидание продолжается
func (o *ControlsOverlay) updateCapture() {
	if o.menuInput.PauseJustPressed() {
		o.cursor.Capturing = false
		o.rejectTicks = 0
		o.cursor.Rejected = false
		return
	}

	capture := o.capture.JustPressedKey
	if o.cursor.Column == types.ControlsDeviceGamepad {
		capture = o.capture.JustPressedPadButton
	}
	name, pressed := capture()
	if !pressed {
		return
	}

	err := o.controlsUseCases.Bind(o.controls, o.slot(), name)
	switch {
	case errors.Is(err, types.ErrBindingTaken):
		o.rejectTicks = controlsRejectTicks
		o.cursor.Rejected = true
		return
	case err != nil:
		log.Printf("save controls: %v", err)
	}
	o.cursor.Capturing = false
	o.rejectTicks = 0
	o.cursor.Rejected = false
}

// slot — ячейка под курсором
func (o *ControlsOverlay) slot() types.ControlsSlot {
	row := o.controlsUseCases.Rows(o.cursor.Page)[o.cursor.Row]
	return types.ControlsSlot{
		Page:   o.cursor.Page,
		Action: row.Action,
		Hotkey: row.Hotkey,
		Device: o.cursor.Column,
	}
}
