package services

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/models"
	"github.com/shpaker/gonflict/internal/types"
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
	spawnPosition := models.Position{X: 4 * constants.TankSpriteSize, Y: 12 * constants.TankSpriteSize}

	player := models.Tank{
		Image:         tankSprite,
		SpawnPosition: spawnPosition, // Начальная позиция спавна
		WorldPosition: models.Position{
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

func (s *PlayerService) Move(dt float64, wallsColliders []*models.Block) {
	delta := s.tank.Speed * dt

	newX := s.tank.WorldPosition.X
	newY := s.tank.WorldPosition.Y
	switch s.tank.Direction {
	case types.DirectionUp:
		newY = s.tank.WorldPosition.Y - delta
	case types.DirectionDown:
		newY = s.tank.WorldPosition.Y + delta
	case types.DirectionLeft:
		newX = s.tank.WorldPosition.X - delta
	case types.DirectionRight:
		newX = s.tank.WorldPosition.X + delta
	}

	// Проверка пересечения танка с коллайдерами стен
	for _, wall := range wallsColliders {
		if wall == nil || wall.Data == nil {
			continue
		}
		// Прямоугольники танка и блока
		tankLeft := newX
		tankRight := newX + constants.TankSpriteSize
		tankTop := newY
		tankBottom := newY + constants.TankSpriteSize

		blockLeft := wall.WorldPosition.X * constants.TileMinSize
		blockRight := blockLeft + constants.TileMinSize
		blockTop := wall.WorldPosition.Y * constants.TileMinSize
		blockBottom := blockTop + constants.TileMinSize

		if tankRight > blockLeft && tankLeft < blockRight &&
			tankBottom > blockTop && tankTop < blockBottom {
			// Разрешаем коллизию: ставим танк вплотную к препятствию в сторону движения
			switch s.tank.Direction {
			case types.DirectionUp:
				// верх танка упирается в низ блока
				newY = blockBottom
			case types.DirectionDown:
				// низ танка упирается в верх блока
				newY = blockTop - constants.TankSpriteSize
			case types.DirectionLeft:
				// левая сторона танка упирается в правую сторону блока
				newX = blockRight
			case types.DirectionRight:
				// правая сторона танка упирается в левую сторону блока
				newX = blockLeft - constants.TankSpriteSize
			}
			s.tank.Speed = 0
			break
		}
	}
	s.tank.WorldPosition.X = newX
	s.tank.WorldPosition.Y = newY
}

func (s *PlayerService) Update(dt float64, wallsColliders []*models.Block) {
	s.Move(dt, wallsColliders)
}

func (s *PlayerService) Draw(screen *ebiten.Image) {
	if s.tank.Image != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			constants.BattleFieldOffset+s.tank.WorldPosition.X,
			constants.BattleFieldOffset+s.tank.WorldPosition.Y,
		)
		screen.DrawImage(s.tank.Image, op)
	}
}

func (s *PlayerService) KeyPressed() {
	// Rotate the tank if the key is pressed
	playerRotated := false
	if ebiten.IsKeyPressed(ebiten.KeyW) && !playerRotated {
		playerRotated = s.Rotate(types.DirectionUp)
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) && !playerRotated {
		playerRotated = s.Rotate(types.DirectionDown)
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) && !playerRotated {
		playerRotated = s.Rotate(types.DirectionLeft)
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) && !playerRotated {
		playerRotated = s.Rotate(types.DirectionRight)
	}
	// Stop the tank if the key is released
	if inpututil.IsKeyJustReleased(ebiten.KeyW) && s.tank.Direction == types.DirectionUp {
		s.tank.Speed = 0
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyS) && s.tank.Direction == types.DirectionDown {
		s.tank.Speed = 0
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyA) && s.tank.Direction == types.DirectionLeft {
		s.tank.Speed = 0
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyD) && s.tank.Direction == types.DirectionRight {
		s.tank.Speed = 0
	}
}
