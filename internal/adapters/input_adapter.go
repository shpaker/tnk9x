package adapters

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// KeyboardInputAdapter адаптер для обработки пользовательского ввода с клавиатуры
type KeyboardInputAdapter struct {
	tankUseCases    use_cases.ITankUseCasesRef // Для работы с танком
	tankUseCasesRef use_cases.ITankUseCasesRef // Для управления танком (дубликат для совместимости)
	bulletUseCases  use_cases.IBulletUseCases
	upButton        ebiten.Key
	downButton      ebiten.Key
	leftButton      ebiten.Key
	rightButton     ebiten.Key
	shootButton     ebiten.Key
}

// NewKeyboardInputAdapter создает новый экземпляр KeyboardInputAdapter
func NewKeyboardInputAdapter(
	tankUseCases use_cases.ITankUseCasesRef,
	tankUseCasesRef use_cases.ITankUseCasesRef,
	bulletUseCases use_cases.IBulletUseCases,
	upButton ebiten.Key,
	downButton ebiten.Key,
	leftButton ebiten.Key,
	rightButton ebiten.Key,
	shootButton ebiten.Key,
) *KeyboardInputAdapter {
	return &KeyboardInputAdapter{
		tankUseCases:    tankUseCases,
		tankUseCasesRef: tankUseCasesRef,
		bulletUseCases:  bulletUseCases,
		upButton:        upButton,
		downButton:      downButton,
		leftButton:      leftButton,
		rightButton:     rightButton,
		shootButton:     shootButton,
	}
}

// Update обрабатывает пользовательский ввод
func (a *KeyboardInputAdapter) Update() {
	a.keyPressedEvents()
	a.keyReleasedEvents()
}

// keyPressedEvents обрабатывает события нажатия клавиш
func (a *KeyboardInputAdapter) keyPressedEvents() {
	// Проверяем нажатие клавиши стрельбы
	if inpututil.IsKeyJustPressed(a.shootButton) {
		log.Printf("DEBUG: Shoot button pressed (key: %v)", a.shootButton)
		a.tankShoot()
	}

	// Получаем танк
	tank, err := a.tankUseCases.GetPlayerTank()
	if err != nil {
		return
	}

	// Rotate the tank if the key is pressed
	tankRotated := false
	if ebiten.IsKeyPressed(a.upButton) && !tankRotated {
		a.tankUseCasesRef.RotateTank(tank, types.DirectionUp)
		tankRotated = true
	}
	if ebiten.IsKeyPressed(a.downButton) && !tankRotated {
		a.tankUseCasesRef.RotateTank(tank, types.DirectionDown)
		tankRotated = true
	}
	if ebiten.IsKeyPressed(a.leftButton) && !tankRotated {
		a.tankUseCasesRef.RotateTank(tank, types.DirectionLeft)
		tankRotated = true
	}
	if ebiten.IsKeyPressed(a.rightButton) && !tankRotated {
		a.tankUseCasesRef.RotateTank(tank, types.DirectionRight)
		tankRotated = true
	}
}

// keyReleasedEvents обрабатывает события отпускания клавиш
func (a *KeyboardInputAdapter) keyReleasedEvents() {
	// Stop the tank if the key is released
	tank, err := a.tankUseCases.GetPlayerTank()
	if err != nil {
		return
	}

	if inpututil.IsKeyJustReleased(a.upButton) && tank.Direction == types.DirectionUp {
		a.tankUseCasesRef.StopTank(tank, false)
	}
	if inpututil.IsKeyJustReleased(a.downButton) && tank.Direction == types.DirectionDown {
		a.tankUseCasesRef.StopTank(tank, false)
	}
	if inpututil.IsKeyJustReleased(a.leftButton) && tank.Direction == types.DirectionLeft {
		a.tankUseCasesRef.StopTank(tank, false)
	}
	if inpututil.IsKeyJustReleased(a.rightButton) && tank.Direction == types.DirectionRight {
		a.tankUseCasesRef.StopTank(tank, false)
	}
}

// tankShoot обрабатывает стрельбу танка
func (a *KeyboardInputAdapter) tankShoot() {
	log.Printf("DEBUG: tankShoot called")
	tank, err := a.tankUseCases.GetPlayerTank()
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
