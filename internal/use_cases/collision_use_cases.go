package use_cases

import (
	"github.com/shpaker/gonflict/internal/types"
)

// CollisionUseCases реализация для операций с коллизиями
type CollisionUseCases struct {
	bulletUseCases IBulletUseCases
	playerUseCases IPlayerUseCases
	mapUseCases    IMapUseCases
}

// NewCollisionUseCases создает новый экземпляр CollisionUseCases
func NewCollisionUseCases(
	bulletUseCases IBulletUseCases,
	playerUseCases IPlayerUseCases,
	mapUseCases IMapUseCases,
) *CollisionUseCases {
	return &CollisionUseCases{
		bulletUseCases: bulletUseCases,
		playerUseCases: playerUseCases,
		mapUseCases:    mapUseCases,
	}
}

// UpdateCollisions обновляет все коллизии в игре
func (uc *CollisionUseCases) UpdateCollisions() error {
	player, err := uc.playerUseCases.GetPlayer()
	if err != nil {
		return err
	}

	uc.checkBulletBoundaryCollisions()
	uc.checkBulletPlayerCollisions(player)
	uc.checkBulletWallCollisions()
	uc.checkPlayerBoundaryCollisions(player)
	uc.checkPlayerWallCollisions(player)

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

// checkBulletPlayerCollisions проверяет коллизии пуль с игроком
func (uc *CollisionUseCases) checkBulletPlayerCollisions(player *types.TankEntity) {
	bullets := uc.bulletUseCases.GetBullets()

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем коллизию между пулей и игроком
		if bullet.Owner != player && uc.CheckColliders(bullet, player) {
			// Удаляем пулю при попадании в игрока
			uc.bulletUseCases.RemoveBullet(i)
			// Здесь можно добавить логику обработки попадания в игрока
			// println("Player hit by bullet!")
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

			// Проверяем, является ли блок коллидируемым
			if block.Properties != nil && !block.Properties.Collidable {
				continue
			}

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

// checkPlayerBoundaryCollisions проверяет коллизии игрока с границами экрана
func (uc *CollisionUseCases) checkPlayerBoundaryCollisions(player *types.TankEntity) {
	if player.WorldPosition.X < 0 {
		player.WorldPosition.X = 0
		uc.playerUseCases.StopPlayer(false)
	}
	if player.WorldPosition.Y < 0 {
		player.WorldPosition.Y = 0
		uc.playerUseCases.StopPlayer(false)
	}
	if player.WorldPosition.X > MapWidthHeight-TankSpriteSize {
		player.WorldPosition.X = MapWidthHeight - TankSpriteSize
		uc.playerUseCases.StopPlayer(false)
	}
	if player.WorldPosition.Y > MapWidthHeight-TankSpriteSize {
		player.WorldPosition.Y = MapWidthHeight - TankSpriteSize
		uc.playerUseCases.StopPlayer(false)
	}
}

// checkPlayerWallCollisions проверяет коллизии игрока со стенами
func (uc *CollisionUseCases) checkPlayerWallCollisions(player *types.TankEntity) {
	level := uc.mapUseCases.GetBlocks()
	mapObjects := uc.createMapObjectsFromLevel(level)

	// Проверяем коллизии с использованием внутренних методов
	collidingObject := uc.CheckCollidersWithArrayFirst(player, mapObjects)

	if collidingObject != nil {
		uc.handlePlayerWallCollision(player, collidingObject.(*types.BlockEntity))
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

// handlePlayerWallCollision обрабатывает коллизию игрока со стеной
func (uc *CollisionUseCases) handlePlayerWallCollision(player *types.TankEntity, block *types.BlockEntity) {
	blockPos := block.GetWorldPosition()
	blockSize := block.GetSize()

	switch player.Direction {
	case types.DirectionUp:
		// верх танка упирается в низ блока
		player.WorldPosition.Y = blockPos.Y + float64(blockSize.Height)
	case types.DirectionDown:
		// низ танка упирается в верх блока
		player.WorldPosition.Y = blockPos.Y - float64(TankSpriteSize)
	case types.DirectionLeft:
		// левая сторона танка упирается в правую сторону блока
		player.WorldPosition.X = blockPos.X + float64(blockSize.Width)
	case types.DirectionRight:
		// правая сторона танка упирается в левую сторону блока
		player.WorldPosition.X = blockPos.X - float64(TankSpriteSize)
	}
	uc.playerUseCases.StopPlayer(true)
}

// CheckColliders проверяет коллизию между двумя объектами карты
func (uc *CollisionUseCases) CheckColliders(
	obj1 IMapObject,
	obj2 IMapObject,
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
