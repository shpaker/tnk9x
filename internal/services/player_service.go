package services

import (
	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/models"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/utils"
)

type PlayerService struct {
	spritesRepository interfaces.ISpritesRepository
	tank              models.Tank
}

func NewPlayerService(spritesRepository interfaces.ISpritesRepository) *PlayerService {

	service := &PlayerService{
		spritesRepository: spritesRepository,
	}
	firstplayers, _ := service.makePlayer()
	service.tank = firstplayers
	return service
}

func (s *PlayerService) makePlayer() (models.Tank, error) {
	tankSprite, err := s.spritesRepository.GetSprite("enemies", "enemy_basic")
	if err != nil {
		return models.Tank{}, err
	}

	// Создаем игрока с начальными параметрами
	spawnPosition := types.Position{X: 4 * constants.TankSpriteSize, Y: 12 * constants.TankSpriteSize}

	player := models.Tank{
		Image:         tankSprite,
		SpawnPosition: spawnPosition, // Начальная позиция спавна
		WorldPosition: types.Position{
			X: spawnPosition.X,
			Y: spawnPosition.Y,
		}, // Текущая позиция в мире
		Speed:     0,
		Direction: types.DirectionUp,
	}

	return player, nil
}

func (s *PlayerService) GetPlayer() (*models.Tank, error) {
	return &s.tank, nil
}

func (s *PlayerService) GetDirection() types.Direction {
	return s.tank.Direction
}

func (s *PlayerService) Rotate(
	direction types.Direction,
) bool {
	if s.tank.Speed != 0 {
		return false
	}
	s.tank.Speed = 32.0
	if s.tank.Direction == direction {
		return false
	}
	s.tank.Direction = direction
	return true
}

func (s *PlayerService) Stop(byCollision bool) {
	s.tank.Speed = 0
	if byCollision {
		s.tank.WorldPosition.X = float64(int(s.tank.WorldPosition.X))
		s.tank.WorldPosition.Y = float64(int(s.tank.WorldPosition.Y))
		return
	}
	if s.tank.Direction == types.DirectionUp {
		s.tank.WorldPosition.Y = float64(utils.RoundToEven(s.tank.WorldPosition.Y, false))
	}
	if s.tank.Direction == types.DirectionDown {
		s.tank.WorldPosition.Y = float64(utils.RoundToEven(s.tank.WorldPosition.Y, true))
	}
	if s.tank.Direction == types.DirectionLeft {
		s.tank.WorldPosition.X = float64(utils.RoundToEven(s.tank.WorldPosition.X, false))
	}
	if s.tank.Direction == types.DirectionRight {
		s.tank.WorldPosition.X = float64(utils.RoundToEven(s.tank.WorldPosition.X, true))
	}
}

func (s *PlayerService) Move(dt float64) {
	delta := s.tank.Speed * dt

	switch s.tank.Direction {
	case types.DirectionUp:
		s.tank.WorldPosition.Y -= delta
	case types.DirectionDown:
		s.tank.WorldPosition.Y += delta
	case types.DirectionLeft:
		s.tank.WorldPosition.X -= delta
	case types.DirectionRight:
		s.tank.WorldPosition.X += delta
	}
}

func (s *PlayerService) Update(dt float64) {
	// println(s.tank.WorldPosition.X, s.tank.WorldPosition.Y)
	s.Move(dt)
}
