// Package menu_input собирает ввод меню и паузы из всех источников:
// фиксированные клавиши, любой геймпад и тач-контроллы любого игрока.
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

	// Стики геймпадов: шаг меню — только при отклонении или смене
	// направления, удержание не листает
	sticks map[ebiten.GamepadID]stickState

	// Ebiten-функции инжектируются для headless-тестов
	isKeyJustPressed    func(ebiten.Key) bool
	appendGamepads      func([]ebiten.GamepadID) []ebiten.GamepadID
	isButtonJustPressed func(ebiten.GamepadID, ebiten.StandardGamepadButton) bool
	axisValue           func(ebiten.GamepadID, ebiten.StandardGamepadAxis) float64

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
	}
}

// Update собирает события кадра из клавиатуры, геймпадов и тача
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
