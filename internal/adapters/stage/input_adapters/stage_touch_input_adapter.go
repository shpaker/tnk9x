package input_adapters

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IInputAdapterWithTank = (*StageTouchInputAdapter)(nil)

// StageTouchInputAdapter превращает события экранных контролов в
// команды танку игрока; семантика повторяет клавиатурный адаптер:
// Rotate+Move каждый кадр удержания, один Stop при отпускании
type StageTouchInputAdapter struct {
	tankActions   interfaces.ITankActionsUseCases
	stageUseCases interfaces.IStageUseCases
	touchControls interfaces.ITouchControlsAdapter

	tank *types.TankEntity

	// Крестовина была активна в прошлом кадре: отпускание должно
	// один раз остановить танк
	wasSteering bool
}

func (a *StageTouchInputAdapter) SetPlayerTank(tank *types.TankEntity) {
	a.tank = tank
}

func NewStageTouchInputAdapter(
	tankActions interfaces.ITankActionsUseCases,
	tank *types.TankEntity,
	stageUseCases interfaces.IStageUseCases,
	touchControls interfaces.ITouchControlsAdapter,
) *StageTouchInputAdapter {
	return &StageTouchInputAdapter{
		tankActions:   tankActions,
		tank:          tank,
		stageUseCases: stageUseCases,
		touchControls: touchControls,
	}
}

func (a *StageTouchInputAdapter) Update(dt float64) {
	if a.touchControls.PauseJustPressed() {
		a.stageUseCases.TogglePause()
	}

	if a.tank == nil {
		return
	}

	if a.stageUseCases.IsPaused() {
		a.tankActions.Stop(a.tank, false)
		a.wasSteering = false

		return
	}

	if a.touchControls.FireJustPressed() {
		_ = a.tankActions.Shoot(a.tank)
	}

	if direction, ok := a.touchControls.DPadDirection(); ok {
		_ = a.tankActions.Rotate(a.tank, direction)
		_ = a.tankActions.Move(a.tank)
		a.wasSteering = true

		return
	}
	if a.wasSteering {
		a.tankActions.Stop(a.tank, false)
		a.wasSteering = false
	}
}
