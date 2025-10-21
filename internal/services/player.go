package services

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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

func (s *PlayerService) Move(dt float64, level models.Level) {
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

	if newX < 0 {
		s.tank.WorldPosition.X = 0
		s.Stop(false)
		return
	}

	if newY < 0 {
		s.tank.WorldPosition.Y = 0
		s.Stop(false)
		return
	}

	if newX > constants.BattleFieldWidthHeight-constants.TankSpriteSize {
		s.tank.WorldPosition.X = constants.BattleFieldWidthHeight - constants.TankSpriteSize
		s.Stop(false)
		return
	}

	if newY > constants.BattleFieldWidthHeight-constants.TankSpriteSize {
		s.tank.WorldPosition.Y = constants.BattleFieldWidthHeight - constants.TankSpriteSize
		s.Stop(false)
		return
	}

	// Создаем временный объект танка с новой позицией для проверки коллизий
	tempTank := models.Tank{
		Image:         s.tank.Image,
		WorldPosition: types.Position{X: newX, Y: newY},
		Speed:         s.tank.Speed,
		Direction:     s.tank.Direction,
	}

	// Преобразуем блоки уровня в массив IMapObject
	var mapObjects []interfaces.IMapObject
	for _, wall := range level {
		if wall.Data != nil {
			// Создаем блок с правильной позицией в мире
			block := models.Block{
				Image:      wall.Image,
				Data:       wall.Data,
				Properties: wall.Properties,
				WorldPosition: types.Position{
					X: wall.WorldPosition.X * constants.TileMinSize,
					Y: wall.WorldPosition.Y * constants.TileMinSize,
				},
			}
			mapObjects = append(mapObjects, &block)
		}
	}

	// Проверяем коллизии с использованием утилит
	collidingObject := utils.CheckCollidersWithArrayFirst(&tempTank, mapObjects)

	if collidingObject != nil {
		// Есть коллизия - ставим танк вплотную к препятствию в сторону движения
		block := collidingObject.(*models.Block)
		blockPos := block.GetWorldPosition()
		blockSize := block.GetSize()

		switch s.tank.Direction {
		case types.DirectionUp:
			// верх танка упирается в низ блока
			newY = blockPos.Y + float64(blockSize.Height)
		case types.DirectionDown:
			// низ танка упирается в верх блока
			newY = blockPos.Y - float64(constants.TankSpriteSize)
		case types.DirectionLeft:
			// левая сторона танка упирается в правую сторону блока
			newX = blockPos.X + float64(blockSize.Width)
		case types.DirectionRight:
			// правая сторона танка упирается в левую сторону блока
			newX = blockPos.X - float64(constants.TankSpriteSize)
		}
		s.Stop(true)
	}
	s.tank.WorldPosition.X = newX
	s.tank.WorldPosition.Y = newY
}

func (s *PlayerService) Update(dt float64, level models.Level) {
	// println(s.tank.WorldPosition.X, s.tank.WorldPosition.Y)
	s.Move(dt, level)
}

func (s *PlayerService) Draw(screen *ebiten.Image) {
	if s.tank.Image != nil {
		// Определяем угол поворота в зависимости от направления
		var rotationAngle float64
		switch s.tank.Direction {
		case types.DirectionUp:
			rotationAngle = 0
		case types.DirectionRight:
			rotationAngle = math.Pi / 2
		case types.DirectionDown:
			rotationAngle = math.Pi
		case types.DirectionLeft:
			rotationAngle = 3 * math.Pi / 2
		}

		// Вычисляем позицию на экране
		screenX := constants.BattleFieldOffset + s.tank.WorldPosition.X
		screenY := constants.BattleFieldOffset + s.tank.WorldPosition.Y

		// Используем функцию для поворота изображения вокруг центра
		op := utils.RotateImage(s.tank.Image, rotationAngle, screenX, screenY)

		screen.DrawImage(s.tank.Image, op)
	}
}

func (s *PlayerService) KeyPressed() (isShoot bool) {
	// Rotate the tank if the key is pressed
	isShoot = false
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
		s.Stop(false)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyS) && s.tank.Direction == types.DirectionDown {
		s.Stop(false)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyA) && s.tank.Direction == types.DirectionLeft {
		s.Stop(false)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyD) && s.tank.Direction == types.DirectionRight {
		s.Stop(false)
	}

	// Shoot if the space key is pressed
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		isShoot = true
	}

	return isShoot
}
