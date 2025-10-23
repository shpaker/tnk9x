package services

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/shpaker/gonflict/internal/types"
)

type ControllerService struct {
	playerService  *PlayerService
	bulletsService *BulletsService
	upButton       ebiten.Key
	downButton     ebiten.Key
	leftButton     ebiten.Key
	rightButton    ebiten.Key
	shootButton    ebiten.Key
}

func NewControllerService(
	playerService *PlayerService,
	bulletsService *BulletsService,
	upButton ebiten.Key,
	downButton ebiten.Key,
	leftButton ebiten.Key,
	rightButton ebiten.Key,
	shootButton ebiten.Key,
) *ControllerService {
	return &ControllerService{
		playerService:  playerService,
		bulletsService: bulletsService,
		upButton:       upButton,
		downButton:     downButton,
		leftButton:     leftButton,
		rightButton:    rightButton,
		shootButton:    shootButton,
	}
}

func (s *ControllerService) Update() {
	s.keyPressedEvents()
	s.keyReleasedEvents()
}

func (s *ControllerService) keyPressedEvents() {
	// Rotate the tank if the key is pressed
	playerRotated := false
	if ebiten.IsKeyPressed(s.upButton) && !playerRotated {
		playerRotated = s.playerService.Rotate(types.DirectionUp)
	}
	if ebiten.IsKeyPressed(s.downButton) && !playerRotated {
		playerRotated = s.playerService.Rotate(types.DirectionDown)
	}
	if ebiten.IsKeyPressed(s.leftButton) && !playerRotated {
		playerRotated = s.playerService.Rotate(types.DirectionLeft)
	}
	if ebiten.IsKeyPressed(s.rightButton) && !playerRotated {
		playerRotated = s.playerService.Rotate(types.DirectionRight)
	}

	// Shoot if the shoot key is pressed
	if inpututil.IsKeyJustPressed(s.shootButton) {
		s.playerShoot()
	}
}

func (s *ControllerService) keyReleasedEvents() {
	// Stop the tank if the key is released
	if inpututil.IsKeyJustReleased(s.upButton) && s.playerService.GetDirection() == types.DirectionUp {
		s.playerService.Stop(false)
	}
	if inpututil.IsKeyJustReleased(s.downButton) && s.playerService.GetDirection() == types.DirectionDown {
		s.playerService.Stop(false)
	}
	if inpututil.IsKeyJustReleased(s.leftButton) && s.playerService.GetDirection() == types.DirectionLeft {
		s.playerService.Stop(false)
	}
	if inpututil.IsKeyJustReleased(s.rightButton) && s.playerService.GetDirection() == types.DirectionRight {
		s.playerService.Stop(false)
	}
}

func (s *ControllerService) playerShoot() {
	player, err := s.playerService.GetPlayer()
	if err == nil {
		s.bulletsService.AddBullet(player)
	}
}
