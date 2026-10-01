package input_adapters

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk9x/internal/adapters/bindings"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IInputAdapterWithTank = (*StageKeyboardInputAdapter)(nil)

// directionActions — действия движения в порядке приоритета
// при нескольких зажатых клавишах
var directionActions = []struct {
	action    types.InputAction
	direction types.Direction
}{
	{types.InputActionUp, types.DirectionUp},
	{types.InputActionDown, types.DirectionDown},
	{types.InputActionLeft, types.DirectionLeft},
	{types.InputActionRight, types.DirectionRight},
}

// StageKeyboardInputAdapter управляет танком игрока с клавиатуры
// по его раскладке; клавиши читаются каждый кадр, так что
// переназначение из меню паузы действует сразу
type StageKeyboardInputAdapter struct {
	tankActions   interfaces.ITankActionsUseCases
	stageUseCases interfaces.IStageUseCases

	// Entities
	controls *types.ControlsEntity
	player   types.PlayerTankNum

	tank *types.TankEntity
}

func (a *StageKeyboardInputAdapter) SetPlayerTank(tank *types.TankEntity) {
	a.tank = tank
}

func NewStageKeyboardInputAdapter(
	tankActions interfaces.ITankActionsUseCases,
	tank *types.TankEntity,
	stageUseCases interfaces.IStageUseCases,
	controls *types.ControlsEntity,
	player types.PlayerTankNum,
) *StageKeyboardInputAdapter {
	return &StageKeyboardInputAdapter{
		tankActions:   tankActions,
		tank:          tank,
		stageUseCases: stageUseCases,
		controls:      controls,
		player:        player,
	}
}

func (a *StageKeyboardInputAdapter) Update(dt float64) {
	a.keyPressedEvents()
	a.keyReleasedEvents()
}

// key — клавиша действия по раскладке игрока
func (a *StageKeyboardInputAdapter) key(
	action types.InputAction,
) (ebiten.Key, bool) {
	return bindings.KeyByName(a.controls.GetKey(a.player, action))
}

func (a *StageKeyboardInputAdapter) keyPressedEvents() {
	if a.tank == nil || a.stageUseCases.IsPaused() {
		return
	}

	if key, ok := a.key(types.InputActionFire); ok &&
		inpututil.IsKeyJustPressed(key) {
		_ = a.tankActions.Shoot(a.tank)
	}

	for _, move := range directionActions {
		key, ok := a.key(move.action)
		if ok && ebiten.IsKeyPressed(key) {
			_ = a.tankActions.Rotate(a.tank, move.direction)
			_ = a.tankActions.Move(a.tank)
			return
		}
	}
}

func (a *StageKeyboardInputAdapter) keyReleasedEvents() {
	if a.tank == nil {
		return
	}

	if a.stageUseCases.IsPaused() {
		a.tankActions.Stop(a.tank, false)
		return
	}

	for _, move := range directionActions {
		key, ok := a.key(move.action)
		if ok && inpututil.IsKeyJustReleased(key) &&
			a.tank.Direction == move.direction {
			a.tankActions.Stop(a.tank, false)
		}
	}
}
