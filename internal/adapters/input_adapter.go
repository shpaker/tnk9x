package adapters

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// InputAdapter адаптер для обработки пользовательского ввода
type InputAdapter struct {
	tankUseCases   use_cases.ITankUseCases
	bulletUseCases use_cases.IBulletUseCases
	upButton       ebiten.Key
	downButton     ebiten.Key
	leftButton     ebiten.Key
	rightButton    ebiten.Key
	shootButton    ebiten.Key
}

// NewInputAdapter создает новый экземпляр InputAdapter
func NewInputAdapter(
	tankUseCases use_cases.ITankUseCases,
	bulletUseCases use_cases.IBulletUseCases,
	upButton ebiten.Key,
	downButton ebiten.Key,
	leftButton ebiten.Key,
	rightButton ebiten.Key,
	shootButton ebiten.Key,
) *InputAdapter {
	return &InputAdapter{
		tankUseCases:   tankUseCases,
		bulletUseCases: bulletUseCases,
		upButton:       upButton,
		downButton:     downButton,
		leftButton:     leftButton,
		rightButton:    rightButton,
		shootButton:    shootButton,
	}
}

// Update обрабатывает пользовательский ввод
func (a *InputAdapter) Update() {
	a.keyPressedEvents()
	a.keyReleasedEvents()
}

// keyPressedEvents обрабатывает события нажатия клавиш
func (a *InputAdapter) keyPressedEvents() {
	// Проверяем нажатие клавиши стрельбы
	if inpututil.IsKeyJustPressed(a.shootButton) {
		log.Printf("DEBUG: Shoot button pressed (key: %v)", a.shootButton)
		a.tankShoot()
	}

	// Rotate the tank if the key is pressed
	tankRotated := false
	if ebiten.IsKeyPressed(a.upButton) && !tankRotated {
		a.tankUseCases.RotateTank(types.DirectionUp)
		tankRotated = true
	}
	if ebiten.IsKeyPressed(a.downButton) && !tankRotated {
		a.tankUseCases.RotateTank(types.DirectionDown)
		tankRotated = true
	}
	if ebiten.IsKeyPressed(a.leftButton) && !tankRotated {
		a.tankUseCases.RotateTank(types.DirectionLeft)
		tankRotated = true
	}
	if ebiten.IsKeyPressed(a.rightButton) && !tankRotated {
		a.tankUseCases.RotateTank(types.DirectionRight)
		tankRotated = true
	}
}

// keyReleasedEvents обрабатывает события отпускания клавиш
func (a *InputAdapter) keyReleasedEvents() {
	// Stop the tank if the key is released
	if inpututil.IsKeyJustReleased(a.upButton) && a.tankUseCases.GetDirection() == types.DirectionUp {
		a.tankUseCases.StopTank(false)
	}
	if inpututil.IsKeyJustReleased(a.downButton) && a.tankUseCases.GetDirection() == types.DirectionDown {
		a.tankUseCases.StopTank(false)
	}
	if inpututil.IsKeyJustReleased(a.leftButton) && a.tankUseCases.GetDirection() == types.DirectionLeft {
		a.tankUseCases.StopTank(false)
	}
	if inpututil.IsKeyJustReleased(a.rightButton) && a.tankUseCases.GetDirection() == types.DirectionRight {
		a.tankUseCases.StopTank(false)
	}
}

// tankShoot обрабатывает стрельбу танка
func (a *InputAdapter) tankShoot() {
	log.Printf("DEBUG: tankShoot called")
	tank, err := a.tankUseCases.GetTank()
	if err != nil {
		log.Printf("ERROR: Failed to get tank: %v", err)
		return
	}
	log.Printf("DEBUG: Got tank, calling ShootBullet")
	err = a.bulletUseCases.ShootBullet(tank)
	if err != nil {
		log.Printf("ERROR: Failed to shoot bullet: %v", err)
	}
}
