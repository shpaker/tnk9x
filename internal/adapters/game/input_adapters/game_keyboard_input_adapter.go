package input_adapters

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// GameKeyboardInputAdapter адаптер для обработки пользовательского ввода с клавиатуры
type GameKeyboardInputAdapter struct {
	tankActions interfaces.ITankActionsUseCases
	tank        *types.TankEntity
	upButton    ebiten.Key
	downButton  ebiten.Key
	leftButton  ebiten.Key
	rightButton ebiten.Key
	shootButton ebiten.Key
}

// SetPlayerTank назначает танк игрока для управления клавиатурой
func (a *GameKeyboardInputAdapter) SetPlayerTank(tank *types.TankEntity) {
	a.tank = tank
}

// NewGameKeyboardInputAdapter создает новый экземпляр GameKeyboardInputAdapter
func NewGameKeyboardInputAdapter(
	tankActions interfaces.ITankActionsUseCases,
	tank *types.TankEntity,
	upButton ebiten.Key,
	downButton ebiten.Key,
	leftButton ebiten.Key,
	rightButton ebiten.Key,
	shootButton ebiten.Key,
) *GameKeyboardInputAdapter {
	return &GameKeyboardInputAdapter{
		tankActions: tankActions,
		tank:        tank,
		upButton:    upButton,
		downButton:  downButton,
		leftButton:  leftButton,
		rightButton: rightButton,
		shootButton: shootButton,
	}
}

// Update обрабатывает пользовательский ввод
func (a *GameKeyboardInputAdapter) Update(dt float64) {
	a.keyPressedEvents()
	a.keyReleasedEvents()
}

// keyPressedEvents обрабатывает события нажатия клавиш
func (a *GameKeyboardInputAdapter) keyPressedEvents() {
	// Проверяем нажатие клавиши стрельбы
	if inpututil.IsKeyJustPressed(a.shootButton) {
		a.tankShoot()
	}

	// Пропускаем если танка нет
	if a.tank == nil {
		return
	}

	// Rotate the tank if the key is pressed
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
		// tankRotated = true // Не устанавливаем здесь, так как это последняя проверка
	}
}

// keyReleasedEvents обрабатывает события отпускания клавиш
func (a *GameKeyboardInputAdapter) keyReleasedEvents() {
	// Stop the tank if the key is released
	if a.tank == nil {
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

// tankShoot обрабатывает стрельбу танка
func (a *GameKeyboardInputAdapter) tankShoot() {
	if a.tank == nil {
		return
	}
	_ = a.tankActions.Shoot(a.tank)
}
