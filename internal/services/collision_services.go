package services

import (
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/utils"
)

// CollisionService предоставляет логику обработки коллизий
type CollisionService struct {
	mapWidthHeight int
	tankSpriteSize int
	tileMinSize    int
	checkColliders func(obj1 types.IMapObject, obj2 types.IMapObject) bool
	checkWithArray func(obj types.IMapObject, objects []types.IMapObject) types.IMapObject
}

// NewCollisionService создает новый сервис коллизий
func NewCollisionService(
	mapWidthHeight int,
	tankSpriteSize int,
	tileMinSize int,
	checkColliders func(obj1 types.IMapObject, obj2 types.IMapObject) bool,
	checkWithArray func(obj types.IMapObject, objects []types.IMapObject) types.IMapObject,
) *CollisionService {
	return &CollisionService{
		mapWidthHeight: mapWidthHeight,
		tankSpriteSize: tankSpriteSize,
		tileMinSize:    tileMinSize,
		checkColliders: checkColliders,
		checkWithArray: checkWithArray,
	}
}

// CheckEnemyBoundaryCollisions проверяет коллизии врага с границами экрана
func (s *CollisionService) CheckEnemyBoundaryCollisions(
	enemy *types.TankEntity,
) {
	// Откатываем позицию при выходе за границы
	if enemy.Position.X < 0 {
		enemy.Position.X = 0
		enemy.Speed = 0
	}
	if enemy.Position.Y < 0 {
		enemy.Position.Y = 0
		enemy.Speed = 0
	}
	if enemy.Position.X > float64(s.mapWidthHeight-s.tankSpriteSize) {
		enemy.Position.X = float64(s.mapWidthHeight - s.tankSpriteSize)
		enemy.Speed = 0
	}
	if enemy.Position.Y > float64(s.mapWidthHeight-s.tankSpriteSize) {
		enemy.Position.Y = float64(s.mapWidthHeight - s.tankSpriteSize)
		enemy.Speed = 0
	}

	// Округляем координаты врага до ближайшего кратного 4
	if enemy.Speed == 0 {
		enemy.Position.X = utils.RoundToNearestMultipleOf4(enemy.Position.X)
		enemy.Position.Y = utils.RoundToNearestMultipleOf4(enemy.Position.Y)
	}
}

// HandleEnemyWallCollision обрабатывает коллизию врага со стеной
func (s *CollisionService) HandleEnemyWallCollision(
	enemy *types.TankEntity,
	block *types.BlockEntity,
) {
	blockPos := block.GetPosition()
	blockSize := block.GetSize()

	// Откатываем позицию врага в зависимости от направления
	switch enemy.Direction {
	case types.DirectionUp:
		enemy.Position.Y = blockPos.Y + float64(blockSize.Height)
	case types.DirectionDown:
		enemy.Position.Y = blockPos.Y - float64(s.tankSpriteSize)
	case types.DirectionLeft:
		enemy.Position.X = blockPos.X + float64(blockSize.Width)
	case types.DirectionRight:
		enemy.Position.X = blockPos.X - float64(s.tankSpriteSize)
	}

	// Останавливаем врага
	enemy.Speed = 0

	// Округляем координаты врага до ближайшего кратного 4
	enemy.Position.X = utils.RoundToNearestMultipleOf4(enemy.Position.X)
	enemy.Position.Y = utils.RoundToNearestMultipleOf4(enemy.Position.Y)
}

// CheckBulletBoundaryCollisions проверяет коллизии пуль с границами экрана
func (s *CollisionService) CheckBulletBoundaryCollisions(
	bullets []types.BulletEntity,
) []int {
	var indicesToRemove []int

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Удаляем пули, которые вышли за границы экрана
		if bullet.Position.X < 0 ||
			bullet.Position.X > float64(s.mapWidthHeight) ||
			bullet.Position.Y < 0 ||
			bullet.Position.Y > float64(s.mapWidthHeight) {
			indicesToRemove = append(indicesToRemove, i)
		}
	}

	return indicesToRemove
}

// CheckTankBoundaryCollisions проверяет коллизии танка с границами экрана
// Возвращает true, если была коллизия (нужно остановить танк)
func (s *CollisionService) CheckTankBoundaryCollisions(
	tank *types.TankEntity,
	stop func(bool),
) bool {
	collision := false

	if tank.Position.X < 0 {
		tank.Position.X = 0
		stop(false)
		collision = true
	}
	if tank.Position.Y < 0 {
		tank.Position.Y = 0
		stop(false)
		collision = true
	}
	if tank.Position.X > float64(s.mapWidthHeight-s.tankSpriteSize) {
		tank.Position.X = float64(s.mapWidthHeight - s.tankSpriteSize)
		stop(false)
		collision = true
	}
	if tank.Position.Y > float64(s.mapWidthHeight-s.tankSpriteSize) {
		tank.Position.Y = float64(s.mapWidthHeight - s.tankSpriteSize)
		stop(false)
		collision = true
	}

	return collision
}

// CreateBlockFromWall создает блок из данных стены для проверки коллизий
func (s *CollisionService) CreateBlockFromWall(
	wall types.BlockEntity,
) types.BlockEntity {
	return types.BlockEntity{
		ImageGetter: wall.ImageGetter,
		Data:        wall.Data,
		Position: types.Position{
			X: wall.Position.X * float64(s.tileMinSize),
			Y: wall.Position.Y * float64(s.tileMinSize),
		},
		Altitude: wall.Altitude,
	}
}

// CreateMapObjectsFromLevel создает массив объектов карты из уровня
func (s *CollisionService) CreateMapObjectsFromLevel(
	level []types.BlockEntity,
) []types.IMapObject {
	var mapObjects []types.IMapObject
	for _, wall := range level {
		if wall.Data != nil {
			block := s.CreateBlockFromWall(wall)
			mapObjects = append(mapObjects, &block)
		}
	}
	return mapObjects
}

// HandleTankWallCollision обрабатывает коллизию танка со стеной
func (s *CollisionService) HandleTankWallCollision(
	tank *types.TankEntity,
	block *types.BlockEntity,
	stop func(bool),
) {
	blockPos := block.GetPosition()
	blockSize := block.GetSize()

	switch tank.Direction {
	case types.DirectionUp:
		// верх танка упирается в низ блока
		tank.Position.Y = blockPos.Y + float64(blockSize.Height)
	case types.DirectionDown:
		// низ танка упирается в верх блока
		tank.Position.Y = blockPos.Y - float64(s.tankSpriteSize)
	case types.DirectionLeft:
		// левая сторона танка упирается в правую сторону блока
		tank.Position.X = blockPos.X + float64(blockSize.Width)
	case types.DirectionRight:
		// правая сторона танка упирается в левую сторону блока
		tank.Position.X = blockPos.X - float64(s.tankSpriteSize)
	}
	stop(true)
}

// CheckEnemyWallCollision проверяет коллизию врага со стенами
// Возвращает блок, с которым произошла коллизия, или nil
func (s *CollisionService) CheckEnemyWallCollision(
	enemy *types.TankEntity,
	level []types.BlockEntity,
) *types.BlockEntity {
	mapObjects := s.CreateMapObjectsFromLevel(level)
	collidingObject := s.checkWithArray(enemy, mapObjects)

	if collidingObject != nil {
		if block, ok := collidingObject.(*types.BlockEntity); ok {
			return block
		}
	}
	return nil
}

// CheckTankWallCollision проверяет коллизию танка со стенами
// Возвращает блок, с которым произошла коллизия, или nil
func (s *CollisionService) CheckTankWallCollision(
	tank *types.TankEntity,
	level []types.BlockEntity,
) *types.BlockEntity {
	mapObjects := s.CreateMapObjectsFromLevel(level)
	collidingObject := s.checkWithArray(tank, mapObjects)

	if collidingObject != nil {
		if block, ok := collidingObject.(*types.BlockEntity); ok {
			return block
		}
	}
	return nil
}

// CheckBulletWallCollisions проверяет коллизии пуль со стенами
// Возвращает список индексов пуль для удаления и список индексов блоков для удаления
func (s *CollisionService) CheckBulletWallCollisions(
	bullets []types.BulletEntity,
	level []types.BlockEntity,
	checkColliders func(obj1 types.IMapObject, obj2 types.IMapObject) bool,
) (bulletIndicesToRemove []int, blockIndicesToRemove []int) {
	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем коллизии пули со стенами
		bulletHit := false

		for j, wall := range level {
			if wall.Data == nil {
				continue
			}

			block := s.CreateBlockFromWall(wall)

			// Проверяем коллизию пули с блоком
			if checkColliders(bullet, &block) {
				// Если блок - кирпичная стена, помечаем для удаления
				if wall.Data.Name == types.Brick {
					blockIndicesToRemove = append(blockIndicesToRemove, j)
				}

				bulletHit = true
				// Не прерываем цикл, продолжаем проверять другие блоки
			}
		}

		// Удаляем пулю только после проверки всех блоков
		if bulletHit {
			bulletIndicesToRemove = append(bulletIndicesToRemove, i)
		}
	}

	return bulletIndicesToRemove, blockIndicesToRemove
}

// CheckBulletTankCollision проверяет коллизию пуль с танком
// Возвращает список индексов пуль для удаления
func (s *CollisionService) CheckBulletTankCollision(
	bullets []types.BulletEntity,
	tank *types.TankEntity,
	checkColliders func(obj1 types.IMapObject, obj2 types.IMapObject) bool,
) []int {
	var indicesToRemove []int

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем коллизию между пулей и танком
		if bullet.Owner != tank && checkColliders(bullet, tank) {
			indicesToRemove = append(indicesToRemove, i)
		}
	}

	return indicesToRemove
}

// CheckBulletEnemyCollisions проверяет коллизии пуль с врагами
// Возвращает карту индекс_пули -> индекс_врага для обработки взрыва
func (s *CollisionService) CheckBulletEnemyCollisions(
	bullets []types.BulletEntity,
	enemies []*types.TankEntity,
	checkColliders func(obj1 types.IMapObject, obj2 types.IMapObject) bool,
) (bulletIndicesToRemove []int, enemyIndicesToExplode map[int]int) {
	enemyIndicesToExplode = make(map[int]int)

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем коллизию с каждым врагом
		for enemyIndex, enemy := range enemies {
			// Пропускаем если врага нет
			if enemy == nil {
				continue
			}

			// Проверяем, что пуля не принадлежит этому танку (избегаем самоуничтожения)
			if bullet.Owner == enemy {
				continue
			}

			// Если враг активен и есть коллизия
			if enemy.IsActive() && checkColliders(bullet, enemy) {
				// Удаляем пулю
				bulletIndicesToRemove = append(bulletIndicesToRemove, i)
				// Запоминаем врага для взрыва
				enemyIndicesToExplode[i] = enemyIndex
				// Выходим из цикла врагов, так как пуля уже обработана
				break
			}
		}
	}

	return bulletIndicesToRemove, enemyIndicesToExplode
}
