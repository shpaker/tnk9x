package input_adapters

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk9x/internal/adapters/bindings"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IInputAdapterWithTank = (*StageGamepadInputAdapter)(nil)

// StageGamepadInputAdapter управляет танком игрока с его геймпада
// (первый подключённый у P1, второй у P2): назначенные кнопки
// направлений или левый стик, огонь — назначенная кнопка.
// Семантика как у тача: Rotate+Move при удержании, один Stop при
// отпускании; Start — пауза, её обрабатывает стейт уровня
type StageGamepadInputAdapter struct {
	tankActions   interfaces.ITankActionsUseCases
	stageUseCases interfaces.IStageUseCases

	// Entities
	controls *types.ControlsEntity
	player   types.PlayerTankNum

	tank *types.TankEntity

	// Направление было задано в прошлом кадре: отпускание должно
	// один раз остановить танк
	wasSteering bool

	// Переиспользуемый буфер опроса геймпадов
	gamepads []ebiten.GamepadID
}

func (a *StageGamepadInputAdapter) SetPlayerTank(tank *types.TankEntity) {
	a.tank = tank
}

func NewStageGamepadInputAdapter(
	tankActions interfaces.ITankActionsUseCases,
	tank *types.TankEntity,
	stageUseCases interfaces.IStageUseCases,
	controls *types.ControlsEntity,
	player types.PlayerTankNum,
) *StageGamepadInputAdapter {
	return &StageGamepadInputAdapter{
		tankActions:   tankActions,
		tank:          tank,
		stageUseCases: stageUseCases,
		controls:      controls,
		player:        player,
	}
}

func (a *StageGamepadInputAdapter) Update(dt float64) {
	if a.tank == nil {
		return
	}

	id, gamepads, connected := bindings.PlayerGamepad(a.player, a.gamepads)
	a.gamepads = gamepads
	if !connected || a.stageUseCases.IsPaused() {
		a.release()
		return
	}

	if button, ok := a.button(types.InputActionFire); ok &&
		inpututil.IsStandardGamepadButtonJustPressed(id, button) {
		_ = a.tankActions.Shoot(a.tank)
	}

	if direction, ok := a.direction(id); ok {
		_ = a.tankActions.Rotate(a.tank, direction)
		_ = a.tankActions.Move(a.tank)
		a.wasSteering = true
		return
	}
	a.release()
}

// direction — зажатая кнопка направления или отклонение левого стика
func (a *StageGamepadInputAdapter) direction(
	id ebiten.GamepadID,
) (types.Direction, bool) {
	for _, move := range directionActions {
		button, ok := a.button(move.action)
		if ok && ebiten.IsStandardGamepadButtonPressed(id, button) {
			return move.direction, true
		}
	}
	return bindings.StickDirection(
		ebiten.StandardGamepadAxisValue(
			id, ebiten.StandardGamepadAxisLeftStickHorizontal,
		),
		ebiten.StandardGamepadAxisValue(
			id, ebiten.StandardGamepadAxisLeftStickVertical,
		),
	)
}

// button — кнопка действия по раскладке игрока
func (a *StageGamepadInputAdapter) button(
	action types.InputAction,
) (ebiten.StandardGamepadButton, bool) {
	return bindings.ButtonByName(a.controls.GetButton(a.player, action))
}

// release один раз останавливает танк после отпускания направления
func (a *StageGamepadInputAdapter) release() {
	if !a.wasSteering {
		return
	}
	a.tankActions.Stop(a.tank, false)
	a.wasSteering = false
}
