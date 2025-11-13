package input_adapters

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk25/internal/interfaces"
	"github.com/shpaker/tnk25/internal/types"
)

type StageKeyboardInputAdapter struct {
	tankActions   interfaces.ITankActionsUseCases
	stageUseCases interfaces.IStageUseCases

	tank *types.TankEntity

	upButton    ebiten.Key
	downButton  ebiten.Key
	leftButton  ebiten.Key
	rightButton ebiten.Key
	shootButton ebiten.Key
	pauseButton ebiten.Key
}

func (a *StageKeyboardInputAdapter) SetPlayerTank(tank *types.TankEntity) {
	a.tank = tank
}

func NewStageKeyboardInputAdapter(
	tankActions interfaces.ITankActionsUseCases,
	tank *types.TankEntity,
	stageUseCases interfaces.IStageUseCases,
	upButton ebiten.Key,
	downButton ebiten.Key,
	leftButton ebiten.Key,
	rightButton ebiten.Key,
	shootButton ebiten.Key,
	pauseButton ebiten.Key,
) *StageKeyboardInputAdapter {
	return &StageKeyboardInputAdapter{
		tankActions:   tankActions,
		tank:          tank,
		stageUseCases: stageUseCases,
		upButton:      upButton,
		downButton:    downButton,
		leftButton:    leftButton,
		rightButton:   rightButton,
		shootButton:   shootButton,
		pauseButton:   pauseButton,
	}
}

func (a *StageKeyboardInputAdapter) Update(dt float64) {
	a.handlePauseToggle()
	a.keyPressedEvents()
	a.keyReleasedEvents()
}

func (a *StageKeyboardInputAdapter) handlePauseToggle() {
	if a.stageUseCases == nil {
		return
	}
	if inpututil.IsKeyJustPressed(a.pauseButton) {
		a.stageUseCases.TogglePause()
	}
}

func (a *StageKeyboardInputAdapter) keyPressedEvents() {
	if a.stageUseCases != nil && a.stageUseCases.IsPaused() {
		return
	}

	if inpututil.IsKeyJustPressed(a.shootButton) {
		a.tankShoot()
	}

	if a.tank == nil {
		return
	}

	tankRotated := false
	if ebiten.IsKeyPressed(a.upButton) && !tankRotated {
		_ = a.tankActions.Rotate(a.tank, types.DirectionUp)
		_ = a.tankActions.Move(a.tank)
		tankRotated = true
	}
	if ebiten.IsKeyPressed(a.downButton) && !tankRotated {
		_ = a.tankActions.Rotate(a.tank, types.DirectionDown)
		_ = a.tankActions.Move(a.tank)
		tankRotated = true
	}
	if ebiten.IsKeyPressed(a.leftButton) && !tankRotated {
		_ = a.tankActions.Rotate(a.tank, types.DirectionLeft)
		_ = a.tankActions.Move(a.tank)
		tankRotated = true
	}
	if ebiten.IsKeyPressed(a.rightButton) && !tankRotated {
		_ = a.tankActions.Rotate(a.tank, types.DirectionRight)
		_ = a.tankActions.Move(a.tank)

	}
}

func (a *StageKeyboardInputAdapter) keyReleasedEvents() {
	if a.tank == nil {
		return
	}

	if a.stageUseCases != nil && a.stageUseCases.IsPaused() {
		a.tankActions.Stop(a.tank, false)
		return
	}

	if inpututil.IsKeyJustReleased(a.upButton) &&
		a.tank.Direction == types.DirectionUp {
		a.tankActions.Stop(a.tank, false)
	}
	if inpututil.IsKeyJustReleased(a.downButton) &&
		a.tank.Direction == types.DirectionDown {
		a.tankActions.Stop(a.tank, false)
	}
	if inpututil.IsKeyJustReleased(a.leftButton) &&
		a.tank.Direction == types.DirectionLeft {
		a.tankActions.Stop(a.tank, false)
	}
	if inpututil.IsKeyJustReleased(a.rightButton) &&
		a.tank.Direction == types.DirectionRight {
		a.tankActions.Stop(a.tank, false)
	}
}

func (a *StageKeyboardInputAdapter) tankShoot() {
	if a.tank == nil {
		return
	}
	_ = a.tankActions.Shoot(a.tank)
}
