// Package menu_input собирает ввод меню и паузы из всех источников:
// фиксированные клавиши, любой геймпад, тач-контроллы любого игрока
// и мышь.
package menu_input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk9x/internal/adapters/bindings"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IMenuInputAdapter = (*MenuInputAdapter)(nil)

// Фиксированные клавиши меню: не зависят от раскладки игроков,
// чтобы неудачное переназначение не ломало навигацию
var (
	upKeys      = []ebiten.Key{ebiten.KeyArrowUp, ebiten.KeyW}
	downKeys    = []ebiten.Key{ebiten.KeyArrowDown, ebiten.KeyS}
	leftKeys    = []ebiten.Key{ebiten.KeyArrowLeft, ebiten.KeyA}
	rightKeys   = []ebiten.Key{ebiten.KeyArrowRight, ebiten.KeyD}
	confirmKeys = []ebiten.Key{ebiten.KeyEnter, ebiten.KeySpace}
)

// menuPlayers — тач-контроллы обоих игроков управляют меню
var menuPlayers = []types.PlayerTankNum{
	types.PlayerTankNumPlayer1, types.PlayerTankNumPlayer2,
}

// wheelStepThreshold — накопленная прокрутка колеса на один шаг:
// тачпад даёт много мелких сдвигов за жест
const wheelStepThreshold = 1.0

// stickState — направление левого стика геймпада в прошлом кадре
type stickState struct {
	direction types.Direction
	active    bool
}

// MenuInputAdapter — события меню кадра; Update вызывается раз в кадр
// после опроса тач-контролов
type MenuInputAdapter struct {
	// Adapters
	touchControls interfaces.ITouchControlsAdapter

	// Состояние кадра
	up, down        bool
	side            int
	confirmed, back bool
	pause           bool
	tapped          bool
	tapPosition     types.Position
	pointed         bool
	pointPosition   types.Position

	// Мышь: последняя позиция курсора, защёлка «мышь замечена»
	// и накопленная прокрутка колеса
	cursorX, cursorY int
	cursorKnown      bool
	pointerSeen      bool
	wheelAccumulated float64

	// Стики геймпадов: шаг меню — только при отклонении или смене
	// направления, удержание не листает
	sticks map[ebiten.GamepadID]stickState

	// Ebiten-функции инжектируются для headless-тестов
	isKeyJustPressed    func(ebiten.Key) bool
	appendGamepads      func([]ebiten.GamepadID) []ebiten.GamepadID
	isButtonJustPressed func(ebiten.GamepadID, ebiten.StandardGamepadButton) bool
	axisValue           func(ebiten.GamepadID, ebiten.StandardGamepadAxis) float64
	cursorPosition      func() (int, int)
	isMouseJustPressed  func(ebiten.MouseButton) bool
	wheel               func() (float64, float64)

	// Переиспользуемый буфер опроса
	gamepads []ebiten.GamepadID
}

func NewMenuInputAdapter(
	touchControls interfaces.ITouchControlsAdapter,
) *MenuInputAdapter {
	return &MenuInputAdapter{
		touchControls:       touchControls,
		sticks:              make(map[ebiten.GamepadID]stickState),
		isKeyJustPressed:    inpututil.IsKeyJustPressed,
		appendGamepads:      bindings.StandardGamepads,
		isButtonJustPressed: inpututil.IsStandardGamepadButtonJustPressed,
		axisValue:           ebiten.StandardGamepadAxisValue,
		cursorPosition:      ebiten.CursorPosition,
		isMouseJustPressed:  inpututil.IsMouseButtonJustPressed,
		wheel:               ebiten.Wheel,
	}
}

// Update собирает события кадра из клавиатуры, геймпадов, тача и мыши
func (a *MenuInputAdapter) Update() {
	a.up = a.anyKey(upKeys)
	a.down = a.anyKey(downKeys)
	a.side = 0
	if a.anyKey(leftKeys) {
		a.side = -1
	}
	if a.anyKey(rightKeys) {
		a.side = 1
	}
	a.confirmed = a.anyKey(confirmKeys)
	a.pause = a.isKeyJustPressed(ebiten.KeyEscape)
	a.back = a.pause

	a.updateGamepads()
	a.updateTouch()
	a.updateMouse()
}

func (a *MenuInputAdapter) IsTouchActive() bool {
	return a.touchControls.IsTouchActive()
}

func (a *MenuInputAdapter) Steps() (bool, bool) {
	return a.up, a.down
}

func (a *MenuInputAdapter) SideStep() int {
	return a.side
}

func (a *MenuInputAdapter) Confirmed() bool {
	return a.confirmed
}

func (a *MenuInputAdapter) Back() bool {
	return a.back
}

func (a *MenuInputAdapter) PauseJustPressed() bool {
	return a.pause
}

func (a *MenuInputAdapter) Tapped() (types.Position, bool) {
	return a.tapPosition, a.tapped
}

func (a *MenuInputAdapter) Pointed() (types.Position, bool) {
	return a.pointPosition, a.pointed
}

func (a *MenuInputAdapter) IsPointerActive() bool {
	return a.pointerSeen
}

// updateGamepads: крестовина и стик — шаги, A — выбор, B — назад,
// Start — назад и пауза
func (a *MenuInputAdapter) updateGamepads() {
	a.gamepads = a.appendGamepads(a.gamepads[:0])
	connected := make(map[ebiten.GamepadID]bool, len(a.gamepads))
	for _, id := range a.gamepads {
		connected[id] = true
		pressed := func(button ebiten.StandardGamepadButton) bool {
			return a.isButtonJustPressed(id, button)
		}

		a.applyDirection(pressed(ebiten.StandardGamepadButtonLeftTop),
			pressed(ebiten.StandardGamepadButtonLeftBottom),
			pressed(ebiten.StandardGamepadButtonLeftLeft),
			pressed(ebiten.StandardGamepadButtonLeftRight))
		a.applyStick(id)

		a.confirmed = a.confirmed ||
			pressed(ebiten.StandardGamepadButtonRightBottom)
		start := pressed(ebiten.StandardGamepadButtonCenterRight)
		a.pause = a.pause || start
		a.back = a.back || start ||
			pressed(ebiten.StandardGamepadButtonRightRight)
	}

	// Отключённые геймпады забываются
	for id := range a.sticks {
		if !connected[id] {
			delete(a.sticks, id)
		}
	}
}

// applyStick даёт шаг при отклонении стика или смене его направления
func (a *MenuInputAdapter) applyStick(id ebiten.GamepadID) {
	x := a.axisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal)
	y := a.axisValue(id, ebiten.StandardGamepadAxisLeftStickVertical)
	direction, active := bindings.StickDirection(x, y)
	previous := a.sticks[id]
	a.sticks[id] = stickState{direction: direction, active: active}
	if !active || (previous.active && previous.direction == direction) {
		return
	}
	a.applyDirection(
		direction == types.DirectionUp,
		direction == types.DirectionDown,
		direction == types.DirectionLeft,
		direction == types.DirectionRight,
	)
}

// updateTouch: крестовины и огонь обоих игроков, общая пауза, тапы
// по экранным кнопкам
func (a *MenuInputAdapter) updateTouch() {
	a.tapPosition, a.tapped = a.touchControls.TapJustPressed()
	for _, player := range menuPlayers {
		if direction, ok := a.touchControls.DPadJustPressed(player); ok {
			a.applyDirection(
				direction == types.DirectionUp,
				direction == types.DirectionDown,
				direction == types.DirectionLeft,
				direction == types.DirectionRight,
			)
		}
		a.confirmed = a.confirmed || a.touchControls.FireJustPressed(player)
	}
	if a.touchControls.PauseJustPressed() {
		a.pause = true
		a.back = true
	}
}

// updateMouse: сдвиг курсора — наведение, левая кнопка — тап,
// правая — назад (не пауза: клик в бою игру не останавливает),
// колесо — шаг влево-вправо
func (a *MenuInputAdapter) updateMouse() {
	x, y := a.cursorPosition()
	moved := a.cursorKnown && (x != a.cursorX || y != a.cursorY)
	a.cursorX, a.cursorY, a.cursorKnown = x, y, true

	a.pointed = false
	if moved {
		a.pointerSeen = true
		a.pointPosition, a.pointed = a.touchControls.GamePosition(x, y)
	}

	if a.isMouseJustPressed(ebiten.MouseButtonLeft) {
		a.pointerSeen = true
		// Тап этого кадра важнее: клик в той же точке его не дублирует
		if !a.tapped {
			a.tapPosition, a.tapped = a.touchControls.GamePosition(x, y)
		}
	}
	if a.isMouseJustPressed(ebiten.MouseButtonRight) {
		a.pointerSeen = true
		a.back = true
	}

	a.updateWheel()
}

// updateWheel копит прокрутку и даёт не больше шага за кадр;
// смена направления сбрасывает накопленное
func (a *MenuInputAdapter) updateWheel() {
	_, offset := a.wheel()
	if offset == 0 {
		return
	}
	if (offset > 0) != (a.wheelAccumulated > 0) {
		a.wheelAccumulated = 0
	}
	a.wheelAccumulated += offset
	switch {
	case a.wheelAccumulated >= wheelStepThreshold:
		a.side = 1
	case a.wheelAccumulated <= -wheelStepThreshold:
		a.side = -1
	default:
		return
	}
	a.wheelAccumulated = 0
}

func (a *MenuInputAdapter) applyDirection(up, down, left, right bool) {
	a.up = a.up || up
	a.down = a.down || down
	if left {
		a.side = -1
	}
	if right {
		a.side = 1
	}
}

func (a *MenuInputAdapter) anyKey(keys []ebiten.Key) bool {
	for _, key := range keys {
		if a.isKeyJustPressed(key) {
			return true
		}
	}
	return false
}
