package use_cases

import (
	"github.com/shpaker/gonflict/internal/types"
)

// CollisionUseCases реализация для операций с коллизиями
type CollisionUseCases struct {
	bulletUseCases IBulletUseCases
	tankUseCases   ITankUseCases
	mapUseCases    IMapUseCases
	enemyUseCases  IEnemyUseCases
}

// NewCollisionUseCases создает новый экземпляр CollisionUseCases
func NewCollisionUseCases(
	bulletUseCases IBulletUseCases,
	tankUseCases ITankUseCases,
	mapUseCases IMapUseCases,
	enemyUseCases IEnemyUseCases,
) *CollisionUseCases {
	return &CollisionUseCases{
		bulletUseCases: bulletUseCases,
		tankUseCases:   tankUseCases,
		mapUseCases:    mapUseCases,
		enemyUseCases:  enemyUseCases,
	}
}

// UpdateCollisions обновляет все коллизии в игре
func (uc *CollisionUseCases) UpdateCollisions() error {
	tank, err := uc.tankUseCases.GetTank()
	if err != nil {
		return err
	}

	uc.checkBulletBoundaryCollisions()
	uc.checkBulletTankCollisions(tank)
	uc.checkBulletEnemyCollisions()
	uc.checkBulletWallCollisions()
	uc.checkTankBoundaryCollisions(tank)
	uc.checkTankWallCollisions(tank)

	return nil
}

// checkBulletBoundaryCollisions проверяет коллизии пуль с границами экрана
func (uc *CollisionUseCases) checkBulletBoundaryCollisions() {
	bullets := uc.bulletUseCases.GetBullets()

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Удаляем пули, которые вышли за границы экрана
		if bullet.WorldPosition.X < 0 || bullet.WorldPosition.X > MapWidthHeight ||
			bullet.WorldPosition.Y < 0 || bullet.WorldPosition.Y > MapWidthHeight {
			uc.bulletUseCases.RemoveBullet(i)
		}
	}
}

// checkBulletTankCollisions проверяет коллизии пуль с танком
func (uc *CollisionUseCases) checkBulletTankCollisions(tank *types.TankEntity) {
	bullets := uc.bulletUseCases.GetBullets()

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем коллизию между пулей и танком
		if bullet.Owner != tank && uc.CheckColliders(bullet, tank) {
			// Удаляем пулю при попадании в танк
			uc.bulletUseCases.RemoveBullet(i)
			// Здесь можно добавить логику обработки попадания в танк
			// println("Tank hit by bullet!")
		}
	}
}

// checkBulletEnemyCollisions проверяет коллизии пуль с врагами
func (uc *CollisionUseCases) checkBulletEnemyCollisions() {
	bullets := uc.bulletUseCases.GetBullets()
	enemies := uc.enemyUseCases.GetEnemies()

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем коллизию с каждым врагом
		for j := len(enemies) - 1; j >= 0; j-- {
			enemy := enemies[j]

			// Если враг заспавнен и есть коллизия
			if enemy.IsSpawned && uc.CheckColliders(bullet, enemy) {
				// Удаляем пулю
				uc.bulletUseCases.RemoveBullet(i)
				// Удаляем врага
				uc.enemyUseCases.RemoveEnemy(j)
				// Выходим из цикла врагов, так как пуля уже удалена
				break
			}
		}
	}
}

// checkBulletWallCollisions проверяет коллизии пуль со стенами
func (uc *CollisionUseCases) checkBulletWallCollisions() {
	bullets := uc.bulletUseCases.GetBullets()
	level := uc.mapUseCases.GetBlocks()

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем коллизии пули со стенами
		bulletHit := false
		var blocksToRemove []int

		for j, wall := range level {
			if wall.Data == nil {
				continue
			}

			block := uc.createBlockFromWall(wall)

			// Проверяем коллизию пули с блоком
			if uc.CheckColliders(bullet, &block) {
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
			blocks := uc.mapUseCases.GetBlocks()
			if blockIndex < len(blocks) {
				uc.mapUseCases.RemoveBlock(&blocks[blockIndex])
			}
		}

		// Удаляем пулю только после проверки всех блоков
		if bulletHit {
			uc.bulletUseCases.RemoveBullet(i)
		}
	}
}

// checkTankBoundaryCollisions проверяет коллизии танка с границами экрана
func (uc *CollisionUseCases) checkTankBoundaryCollisions(tank *types.TankEntity) {
	if tank.WorldPosition.X < 0 {
		tank.WorldPosition.X = 0
		uc.tankUseCases.StopTank(false)
	}
	if tank.WorldPosition.Y < 0 {
		tank.WorldPosition.Y = 0
		uc.tankUseCases.StopTank(false)
	}
	if tank.WorldPosition.X > MapWidthHeight-TankSpriteSize {
		tank.WorldPosition.X = MapWidthHeight - TankSpriteSize
		uc.tankUseCases.StopTank(false)
	}
	if tank.WorldPosition.Y > MapWidthHeight-TankSpriteSize {
		tank.WorldPosition.Y = MapWidthHeight - TankSpriteSize
		uc.tankUseCases.StopTank(false)
	}
}

// checkTankWallCollisions проверяет коллизии танка со стенами
func (uc *CollisionUseCases) checkTankWallCollisions(tank *types.TankEntity) {
	level := uc.mapUseCases.GetBlocks()
	mapObjects := uc.createMapObjectsFromLevel(level)

	// Проверяем коллизии с использованием внутренних методов
	collidingObject := uc.CheckCollidersWithArrayFirst(tank, mapObjects)

	if collidingObject != nil {
		uc.handleTankWallCollision(tank, collidingObject.(*types.BlockEntity))
	}
}

// createBlockFromWall создает блок из данных стены для проверки коллизий
func (uc *CollisionUseCases) createBlockFromWall(wall types.BlockEntity) types.BlockEntity {
	return types.BlockEntity{
		ImageGetter: wall.ImageGetter,
		Data:        wall.Data,
		Properties:  wall.Properties,
		WorldPosition: types.Position{
			X: wall.WorldPosition.X * TileMinSize,
			Y: wall.WorldPosition.Y * TileMinSize,
		},
		Altitude: wall.Altitude,
	}
}

// createMapObjectsFromLevel создает массив объектов карты из уровня
func (uc *CollisionUseCases) createMapObjectsFromLevel(level []types.BlockEntity) []IMapObject {
	var mapObjects []IMapObject
	for _, wall := range level {
		if wall.Data != nil {
			block := uc.createBlockFromWall(wall)
			mapObjects = append(mapObjects, &block)
		}
	}
	return mapObjects
}

// handleTankWallCollision обрабатывает коллизию танка со стеной
func (uc *CollisionUseCases) handleTankWallCollision(tank *types.TankEntity, block *types.BlockEntity) {
	blockPos := block.GetWorldPosition()
	blockSize := block.GetSize()

	switch tank.Direction {
	case types.DirectionUp:
		// верх танка упирается в низ блока
		tank.WorldPosition.Y = blockPos.Y + float64(blockSize.Height)
	case types.DirectionDown:
		// низ танка упирается в верх блока
		tank.WorldPosition.Y = blockPos.Y - float64(TankSpriteSize)
	case types.DirectionLeft:
		// левая сторона танка упирается в правую сторону блока
		tank.WorldPosition.X = blockPos.X + float64(blockSize.Width)
	case types.DirectionRight:
		// правая сторона танка упирается в левую сторону блока
		tank.WorldPosition.X = blockPos.X - float64(TankSpriteSize)
	}
	uc.tankUseCases.StopTank(true)
}

// CheckColliders проверяет коллизию между двумя объектами карты
func (uc *CollisionUseCases) CheckColliders(
	obj1 IMapObject,
	obj2 IMapObject,
) bool {
	// Проверяем, что объекты на одном уровне высоты
	if obj1.GetAltitude() != obj2.GetAltitude() {
		return false
	}

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

// CheckCollidersWithArray проверяет коллизии между объектом и массивом объектов карты
func (uc *CollisionUseCases) CheckCollidersWithArray(
	obj IMapObject,
	objects []IMapObject,
) []IMapObject {
	var collidingObjects []IMapObject

	for _, mapObj := range objects {
		if uc.CheckColliders(obj, mapObj) {
			collidingObjects = append(collidingObjects, mapObj)
		}
	}

	return collidingObjects
}

// CheckCollidersWithArrayFirst проверяет коллизии между объектом и массивом объектов карты
// Возвращает первый коллидирующий объект или nil, если коллизий нет
func (uc *CollisionUseCases) CheckCollidersWithArrayFirst(
	obj IMapObject,
	objects []IMapObject,
) IMapObject {
	for _, mapObj := range objects {
		if uc.CheckColliders(obj, mapObj) {
			return mapObj
		}
	}
	return nil
}
