package services

import (
	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/models"
	"github.com/shpaker/gonflict/internal/types"
)

type CollidersService struct {
	bulletsService *BulletsService
	playerService  *PlayerService
	mapService     MapService
}

func NewCollidersService(
	bulletsService *BulletsService,
	playerService *PlayerService,
	mapService MapService,
) *CollidersService {
	return &CollidersService{
		bulletsService: bulletsService,
		playerService:  playerService,
		mapService:     mapService,
	}
}

func (s *CollidersService) Update() {
	player, err := s.playerService.GetPlayer()
	if err != nil {
		return
	}

	s.checkBulletBoundaryCollisions()
	s.checkBulletPlayerCollisions(player)
	s.checkBulletWallCollisions()
	s.checkPlayerBoundaryCollisions(player)
	s.checkPlayerWallCollisions(player)
}

// checkBulletBoundaryCollisions проверяет коллизии пуль с границами экрана
func (s *CollidersService) checkBulletBoundaryCollisions() {
	bullets := s.bulletsService.GetBullets()

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Удаляем пули, которые вышли за границы экрана
		if bullet.WorldPosition.X < 0 || bullet.WorldPosition.X > constants.BattleFieldWidthHeight ||
			bullet.WorldPosition.Y < 0 || bullet.WorldPosition.Y > constants.BattleFieldWidthHeight {
			s.bulletsService.RemoveBullet(i)
		}
	}
}

// checkBulletPlayerCollisions проверяет коллизии пуль с игроком
func (s *CollidersService) checkBulletPlayerCollisions(player *models.Tank) {
	bullets := s.bulletsService.GetBullets()

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем коллизию между пулей и игроком
		if bullet.Owner != player && s.checkColliders(bullet, player) {
			// Удаляем пулю при попадании в игрока
			s.bulletsService.RemoveBullet(i)
			// Здесь можно добавить логику обработки попадания в игрока
			// println("Player hit by bullet!")
		}
	}
}

// checkBulletWallCollisions проверяет коллизии пуль со стенами
func (s *CollidersService) checkBulletWallCollisions() {
	bullets := s.bulletsService.GetBullets()
	level := s.mapService.GetBlocks()

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем коллизии пули со стенами
		bulletHit := false
		var blocksToRemove []int

		for j, wall := range level {
			if wall.Data == nil {
				continue
			}

			block := s.createBlockFromWall(wall)

			// Проверяем, является ли блок коллидируемым
			if block.Properties != nil && !block.Properties.Collidable {
				continue
			}

			// Проверяем коллизию пули с блоком
			if s.checkColliders(bullet, &block) {
				// Если блок - кирпичная стена, помечаем для удаления
				if wall.Data.Name == types.Brick {
					blocksToRemove = append(blocksToRemove, j)
				}

				bulletHit = true
				// Не прерываем цикл, продолжаем проверять другие блоки
			}
		}

		// Удаляем блоки в обратном порядке (чтобы индексы не сдвигались)
		for k := len(blocksToRemove) - 1; k >= 0; k-- {
			blockIndex := blocksToRemove[k]
			s.mapService.RemoveBlock(&s.mapService.Level[blockIndex])
		}

		// Удаляем пулю только после проверки всех блоков
		if bulletHit {
			s.bulletsService.RemoveBullet(i)
		}
	}
}

// checkPlayerBoundaryCollisions проверяет коллизии игрока с границами экрана
func (s *CollidersService) checkPlayerBoundaryCollisions(player *models.Tank) {
	if player.WorldPosition.X < 0 {
		player.WorldPosition.X = 0
		s.playerService.Stop(false)
	}
	if player.WorldPosition.Y < 0 {
		player.WorldPosition.Y = 0
		s.playerService.Stop(false)
	}
	if player.WorldPosition.X > constants.BattleFieldWidthHeight-constants.TankSpriteSize {
		player.WorldPosition.X = constants.BattleFieldWidthHeight - constants.TankSpriteSize
		s.playerService.Stop(false)
	}
	if player.WorldPosition.Y > constants.BattleFieldWidthHeight-constants.TankSpriteSize {
		player.WorldPosition.Y = constants.BattleFieldWidthHeight - constants.TankSpriteSize
		s.playerService.Stop(false)
	}
}

// checkPlayerWallCollisions проверяет коллизии игрока со стенами
func (s *CollidersService) checkPlayerWallCollisions(
	player *models.Tank,
) {
	level := s.mapService.GetBlocks()
	mapObjects := s.createMapObjectsFromLevel(level)

	// Проверяем коллизии с использованием внутренних методов
	collidingObject := s.checkCollidersWithArrayFirst(player, mapObjects)

	if collidingObject != nil {
		s.handlePlayerWallCollision(player, collidingObject.(*models.Block))
	}
}

// createBlockFromWall создает блок из данных стены для проверки коллизий
func (s *CollidersService) createBlockFromWall(wall models.Block) models.Block {
	return models.Block{
		Image:      wall.Image,
		Data:       wall.Data,
		Properties: wall.Properties,
		WorldPosition: types.Position{
			X: wall.WorldPosition.X * constants.TileMinSize,
			Y: wall.WorldPosition.Y * constants.TileMinSize,
		},
	}
}

// createMapObjectsFromLevel создает массив объектов карты из уровня
func (s *CollidersService) createMapObjectsFromLevel(level models.Level) []interfaces.IMapObject {
	var mapObjects []interfaces.IMapObject
	for _, wall := range level {
		if wall.Data != nil {
			block := s.createBlockFromWall(wall)
			mapObjects = append(mapObjects, &block)
		}
	}
	return mapObjects
}

// handlePlayerWallCollision обрабатывает коллизию игрока со стеной
func (s *CollidersService) handlePlayerWallCollision(player *models.Tank, block *models.Block) {
	blockPos := block.GetWorldPosition()
	blockSize := block.GetSize()

	switch player.Direction {
	case types.DirectionUp:
		// верх танка упирается в низ блока
		player.WorldPosition.Y = blockPos.Y + float64(blockSize.Height)
	case types.DirectionDown:
		// низ танка упирается в верх блока
		player.WorldPosition.Y = blockPos.Y - float64(constants.TankSpriteSize)
	case types.DirectionLeft:
		// левая сторона танка упирается в правую сторону блока
		player.WorldPosition.X = blockPos.X + float64(blockSize.Width)
	case types.DirectionRight:
		// правая сторона танка упирается в левую сторону блока
		player.WorldPosition.X = blockPos.X - float64(constants.TankSpriteSize)
	}
	s.playerService.Stop(true)
}

// checkColliders проверяет коллизию между двумя объектами карты
func (s *CollidersService) checkColliders(
	obj1 interfaces.IMapObject,
	obj2 interfaces.IMapObject,
) bool {
	pos1 := obj1.GetWorldPosition()
	size1 := obj1.GetSize()
	pos2 := obj2.GetWorldPosition()
	size2 := obj2.GetSize()

	// Проверяем пересечение прямоугольников
	return pos1.X < pos2.X+float64(size2.Width) &&
		pos1.X+float64(size1.Width) > pos2.X &&
		pos1.Y < pos2.Y+float64(size2.Height) &&
		pos1.Y+float64(size1.Height) > pos2.Y
}

// checkCollidersWithArray проверяет коллизии между объектом и массивом объектов карты
func (s *CollidersService) checkCollidersWithArray(
	obj interfaces.IMapObject,
	objects []interfaces.IMapObject,
) []interfaces.IMapObject {
	var collidingObjects []interfaces.IMapObject

	for _, mapObj := range objects {
		if s.checkColliders(obj, mapObj) {
			collidingObjects = append(collidingObjects, mapObj)
		}
	}

	return collidingObjects
}

// checkCollidersWithArrayFirst проверяет коллизии между объектом и массивом объектов карты
// Возвращает первый коллидирующий объект или nil, если коллизий нет
func (s *CollidersService) checkCollidersWithArrayFirst(
	obj interfaces.IMapObject,
	objects []interfaces.IMapObject,
) interfaces.IMapObject {
	for _, mapObj := range objects {
		if s.checkColliders(obj, mapObj) {
			return mapObj
		}
	}
	return nil
}
