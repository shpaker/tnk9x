package adapters

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// InputAdapter адаптер для обработки пользовательского ввода
type InputAdapter struct {
	playerUseCases use_cases.IPlayerUseCases
	bulletUseCases use_cases.IBulletUseCases
	upButton       ebiten.Key
	downButton     ebiten.Key
	leftButton     ebiten.Key
	rightButton    ebiten.Key
	shootButton    ebiten.Key
}

// NewInputAdapter создает новый экземпляр InputAdapter
func NewInputAdapter(
	playerUseCases use_cases.IPlayerUseCases,
	bulletUseCases use_cases.IBulletUseCases,
	upButton ebiten.Key,
	downButton ebiten.Key,
	leftButton ebiten.Key,
	rightButton ebiten.Key,
	shootButton ebiten.Key,
) *InputAdapter {
	return &InputAdapter{
		playerUseCases: playerUseCases,
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
	// Rotate the tank if the key is pressed
	playerRotated := false
	if ebiten.IsKeyPressed(a.upButton) && !playerRotated {
		a.playerUseCases.RotatePlayer(types.DirectionUp)
		playerRotated = true
	}
	if ebiten.IsKeyPressed(a.downButton) && !playerRotated {
		a.playerUseCases.RotatePlayer(types.DirectionDown)
		playerRotated = true
	}
	if ebiten.IsKeyPressed(a.leftButton) && !playerRotated {
		a.playerUseCases.RotatePlayer(types.DirectionLeft)
		playerRotated = true
	}
	if ebiten.IsKeyPressed(a.rightButton) && !playerRotated {
		a.playerUseCases.RotatePlayer(types.DirectionRight)
		playerRotated = true
	}

	// Shoot if the shoot key is pressed
	if inpututil.IsKeyJustPressed(a.shootButton) {
		a.playerShoot()
	}
}

// keyReleasedEvents обрабатывает события отпускания клавиш
func (a *InputAdapter) keyReleasedEvents() {
	// Stop the tank if the key is released
	if inpututil.IsKeyJustReleased(a.upButton) && a.playerUseCases.GetDirection() == types.DirectionUp {
		a.playerUseCases.StopPlayer(false)
	}
	if inpututil.IsKeyJustReleased(a.downButton) && a.playerUseCases.GetDirection() == types.DirectionDown {
		a.playerUseCases.StopPlayer(false)
	}
	if inpututil.IsKeyJustReleased(a.leftButton) && a.playerUseCases.GetDirection() == types.DirectionLeft {
		a.playerUseCases.StopPlayer(false)
	}
	if inpututil.IsKeyJustReleased(a.rightButton) && a.playerUseCases.GetDirection() == types.DirectionRight {
		a.playerUseCases.StopPlayer(false)
	}
}

// playerShoot обрабатывает стрельбу игрока
func (a *InputAdapter) playerShoot() {
	player, err := a.playerUseCases.GetPlayer()
	if err == nil {
		a.bulletUseCases.ShootBullet(player)
	}
}
