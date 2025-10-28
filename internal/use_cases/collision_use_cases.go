package use_cases

import (
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/utils"
)

// CollisionUseCases реализация для операций с коллизиями
type CollisionUseCases struct {
	bulletUseCases IBulletUseCases
	tankUseCases   ITankUseCasesRef
	mapUseCases    IMapUseCases
	enemyTanks     []*types.TankEntity
}

// NewCollisionUseCasesWithEnemies создает новый экземпляр CollisionUseCases с массивом врагов
func NewCollisionUseCasesWithEnemies(
	bulletUseCases IBulletUseCases,
	tankUseCases ITankUseCasesRef,
	mapUseCases IMapUseCases,
	enemyTanks []*types.TankEntity,
) *CollisionUseCases {
	return &CollisionUseCases{
		bulletUseCases: bulletUseCases,
		tankUseCases:   tankUseCases,
		mapUseCases:    mapUseCases,
		enemyTanks:     enemyTanks,
	}
}

// UpdateCollisions обновляет все коллизии в игре
func (uc *CollisionUseCases) UpdateCollisions() error {
	tank := uc.tankUseCases.GetTank()
	if tank == nil {
		return nil
	}

	uc.checkBulletBoundaryCollisions()
	uc.checkBulletTankCollisions(tank)
	uc.checkBulletEnemyCollisions()
	uc.checkBulletWallCollisions()
	uc.checkTankBoundaryCollisions(tank)
	uc.checkTankWallCollisions(tank)

	// Проверяем коллизии врагов (БЕЗ коллизий с игроком)
	uc.checkEnemyCollisions()

	return nil
}

// checkEnemyCollisions проверяет коллизии врагов с границами и стенами
func (uc *CollisionUseCases) checkEnemyCollisions() {
	for _, enemy := range uc.enemyTanks {
		if enemy == nil || enemy.State == types.TankStateExploding || enemy.State == types.TankStateSpawning || enemy.State == types.TankStateExploded {
			continue
		}

		// Проверяем коллизии с границами
		uc.checkEnemyBoundaryCollisions(enemy)

		// Проверяем коллизии со стенами
		uc.checkEnemyWallCollisions(enemy)
	}
}

// checkEnemyBoundaryCollisions проверяет коллизии врага с границами экрана
func (uc *CollisionUseCases) checkEnemyBoundaryCollisions(enemy *types.TankEntity) {
	// Откатываем позицию при выходе за границы
	if enemy.Position.X < 0 {
		enemy.Position.X = 0
		enemy.Speed = 0
	}
	if enemy.Position.Y < 0 {
		enemy.Position.Y = 0
		enemy.Speed = 0
	}
	if enemy.Position.X > MapWidthHeight-TankSpriteSize {
		enemy.Position.X = MapWidthHeight - TankSpriteSize
		enemy.Speed = 0
	}
	if enemy.Position.Y > MapWidthHeight-TankSpriteSize {
		enemy.Position.Y = MapWidthHeight - TankSpriteSize
		enemy.Speed = 0
	}

	// Округляем координаты врага до ближайшего кратного 4
	if enemy.Speed == 0 {
		enemy.Position.X = utils.RoundToNearestMultipleOf4(enemy.Position.X)
		enemy.Position.Y = utils.RoundToNearestMultipleOf4(enemy.Position.Y)
	}
}

// checkEnemyWallCollisions проверяет коллизии врага со стенами
func (uc *CollisionUseCases) checkEnemyWallCollisions(enemy *types.TankEntity) {
	level := uc.mapUseCases.GetBlocks()
	mapObjects := uc.createMapObjectsFromLevel(level)

	// Проверяем коллизии с использованием внутренних методов
	collidingObject := uc.CheckCollidersWithArrayFirst(enemy, mapObjects)

	if collidingObject != nil {
		uc.handleEnemyWallCollision(enemy, collidingObject.(*types.BlockEntity))
	}
}

// handleEnemyWallCollision обрабатывает коллизию врага со стеной
func (uc *CollisionUseCases) handleEnemyWallCollision(enemy *types.TankEntity, block *types.BlockEntity) {
	blockPos := block.GetPosition()
	blockSize := block.GetSize()

	// Откатываем позицию врага в зависимости от направления
	switch enemy.Direction {
	case types.DirectionUp:
		enemy.Position.Y = blockPos.Y + float64(blockSize.Height)
	case types.DirectionDown:
		enemy.Position.Y = blockPos.Y - float64(TankSpriteSize)
	case types.DirectionLeft:
		enemy.Position.X = blockPos.X + float64(blockSize.Width)
	case types.DirectionRight:
		enemy.Position.X = blockPos.X - float64(TankSpriteSize)
	}

	// Останавливаем врага
	enemy.Speed = 0

	// Округляем координаты врага до ближайшего кратного 4
	enemy.Position.X = utils.RoundToNearestMultipleOf4(enemy.Position.X)
	enemy.Position.Y = utils.RoundToNearestMultipleOf4(enemy.Position.Y)
}

// checkBulletBoundaryCollisions проверяет коллизии пуль с границами экрана
func (uc *CollisionUseCases) checkBulletBoundaryCollisions() {
	bullets := uc.bulletUseCases.GetBullets()

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Удаляем пули, которые вышли за границы экрана
		if bullet.Position.X < 0 || bullet.Position.X > MapWidthHeight ||
			bullet.Position.Y < 0 || bullet.Position.Y > MapWidthHeight {
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

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем коллизию с каждым врагом
		for _, enemy := range uc.enemyTanks {
			// Пропускаем если врага нет
			if enemy == nil {
				continue
			}

			// Проверяем, что пуля не принадлежит этому танку (избегаем самоуничтожения)
			if bullet.Owner == enemy {
				continue
			}

			// Если враг заспавнен, не взрывается и есть коллизия
			if enemy.State != types.TankStateSpawning && enemy.State != types.TankStateExploding && enemy.State != types.TankStateExploded && uc.CheckColliders(bullet, enemy) {
				// Удаляем пулю
				uc.bulletUseCases.RemoveBullet(i)
				// Удаляем врага
				// Запускаем анимацию взрыва через TankUseCases
				uc.tankUseCases.StartExplosion(enemy)
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
	if tank.Position.X < 0 {
		tank.Position.X = 0
		uc.tankUseCases.StopTank(false)
	}
	if tank.Position.Y < 0 {
		tank.Position.Y = 0
		uc.tankUseCases.StopTank(false)
	}
	if tank.Position.X > MapWidthHeight-TankSpriteSize {
		tank.Position.X = MapWidthHeight - TankSpriteSize
		uc.tankUseCases.StopTank(false)
	}
	if tank.Position.Y > MapWidthHeight-TankSpriteSize {
		tank.Position.Y = MapWidthHeight - TankSpriteSize
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
		Position: types.Position{
			X: wall.Position.X * TileMinSize,
			Y: wall.Position.Y * TileMinSize,
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
	blockPos := block.GetPosition()
	blockSize := block.GetSize()

	switch tank.Direction {
	case types.DirectionUp:
		// верх танка упирается в низ блока
		tank.Position.Y = blockPos.Y + float64(blockSize.Height)
	case types.DirectionDown:
		// низ танка упирается в верх блока
		tank.Position.Y = blockPos.Y - float64(TankSpriteSize)
	case types.DirectionLeft:
		// левая сторона танка упирается в правую сторону блока
		tank.Position.X = blockPos.X + float64(blockSize.Width)
	case types.DirectionRight:
		// правая сторона танка упирается в левую сторону блока
		tank.Position.X = blockPos.X - float64(TankSpriteSize)
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

	pos1 := obj1.GetPosition()
	size1 := obj1.GetSize()
	pos2 := obj2.GetPosition()
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
