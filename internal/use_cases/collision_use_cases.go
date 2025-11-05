package use_cases

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// CollisionUseCases реализация для операций с коллизиями
type CollisionUseCases struct {
	bulletUseCases           interfaces.IBulletUseCases
	tankActions              interfaces.ITankActionsUseCases // Для остановки танка при коллизиях
	mapUseCases              interfaces.IMapUseCases
	hqUseCases               interfaces.IHQUseCases
	tankCommonUseCases       interfaces.ITankCommonUseCases    // Общий use case для всех танков
	tankLifecycleUseCases    interfaces.ITankLifecycleUseCases // Общий lifecycle use case для всех танков
	boundaryCollisionService interfaces.IBoundaryCollisionService
	wallCollisionService     interfaces.IWallCollisionService
	bulletCollisionService   interfaces.IBulletCollisionService
}

// NewCollisionUseCases создает новый экземпляр CollisionUseCases
func NewCollisionUseCases(
	bulletUseCases interfaces.IBulletUseCases,
	tankActions interfaces.ITankActionsUseCases,
	mapUseCases interfaces.IMapUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	boundaryCollisionService interfaces.IBoundaryCollisionService,
	wallCollisionService interfaces.IWallCollisionService,
	bulletCollisionService interfaces.IBulletCollisionService,
	hqUseCases interfaces.IHQUseCases,
) *CollisionUseCases {
	uc := &CollisionUseCases{
		bulletUseCases:           bulletUseCases,
		tankActions:              tankActions,
		mapUseCases:              mapUseCases,
		hqUseCases:               hqUseCases,
		tankCommonUseCases:       tankCommonUseCases,
		tankLifecycleUseCases:    tankLifecycleUseCases,
		boundaryCollisionService: boundaryCollisionService,
		wallCollisionService:     wallCollisionService,
		bulletCollisionService:   bulletCollisionService,
	}

	return uc
}

// UpdateCollisions обновляет все коллизии в игре
func (uc *CollisionUseCases) UpdateCollisions(
	playerTank *types.TankEntity,
	enemyTanks []*types.TankEntity,
	hq *types.HQEntity,
) error {
	if playerTank == nil {
		return nil
	}

	// Проверяем коллизии с HQ ПЕРВЫМИ, чтобы пули не удалялись другими проверками
	uc.checkBulletHQCollisions(hq)
	uc.checkBulletBoundaryCollisions()
	uc.checkBulletTankCollisions(playerTank)
	uc.checkBulletEnemyCollisions(enemyTanks)
	uc.checkBulletWallCollisions()
	uc.checkTankBoundaryCollisions(playerTank)
	uc.checkTankWallCollisions(playerTank)

	// Проверяем коллизии врагов (БЕЗ коллизий с игроком)
	uc.checkEnemyCollisions(enemyTanks)

	return nil
}

// checkEnemyCollisions проверяет коллизии врагов с границами и стенами
func (uc *CollisionUseCases) checkEnemyCollisions(
	enemyTanks []*types.TankEntity,
) {
	for _, enemy := range enemyTanks {
		if enemy == nil || !enemy.IsActive() {
			continue
		}

		// Проверяем коллизии с границами
		hadBoundaryCollision := uc.boundaryCollisionService.CheckEnemyBoundaryCollisions(
			enemy,
		)
		if hadBoundaryCollision {
			uc.tankActions.Stop(enemy, true)
		}

		// Проверяем коллизии со стенами
		level := uc.mapUseCases.GetBlocks()
		collidingBlock := uc.wallCollisionService.CheckEnemyWallCollision(
			enemy,
			level,
		)

		if collidingBlock != nil {
			uc.wallCollisionService.HandleEnemyWallCollision(
				enemy,
				collidingBlock,
			)
			uc.tankActions.Stop(enemy, true)
		}
	}
}

// checkBulletBoundaryCollisions проверяет коллизии пуль с границами экрана
func (uc *CollisionUseCases) checkBulletBoundaryCollisions() {
	bullets := uc.bulletUseCases.GetBullets()
	indicesToRemove := uc.boundaryCollisionService.CheckBulletBoundaryCollisions(
		bullets,
	)

	for _, i := range indicesToRemove {
		_ = uc.bulletUseCases.RemoveBullet(i)
	}
}

// checkBulletTankCollisions проверяет коллизии пуль с танком
func (uc *CollisionUseCases) checkBulletTankCollisions(tank *types.TankEntity) {
	bullets := uc.bulletUseCases.GetBullets()
	indicesToRemove := uc.bulletCollisionService.CheckBulletTankCollision(
		bullets,
		tank,
	)

	for _, i := range indicesToRemove {
		_ = uc.bulletUseCases.RemoveBullet(i)
		// Здесь можно добавить логику обработки попадания в танк
		// println("Tank hit by bullet!")
	}
}

// checkBulletEnemyCollisions проверяет коллизии пуль с врагами
func (uc *CollisionUseCases) checkBulletEnemyCollisions(
	enemyTanks []*types.TankEntity,
) {
	bullets := uc.bulletUseCases.GetBullets()
	bulletIndicesToRemove, enemyIndicesToExplode := uc.bulletCollisionService.CheckBulletEnemyCollisions(
		bullets,
		enemyTanks,
	)

	for _, i := range bulletIndicesToRemove {
		_ = uc.bulletUseCases.RemoveBullet(i)

		// Запускаем анимацию взрыва для врага через общий Lifecycle Use Cases
		if enemyIndex, exists := enemyIndicesToExplode[i]; exists {
			if enemyIndex < len(enemyTanks) &&
				enemyTanks[enemyIndex] != nil {
				_ = uc.tankLifecycleUseCases.Explode(
					enemyTanks[enemyIndex],
				)
			}
		}
	}
}

// checkBulletWallCollisions проверяет коллизии пуль со стенами
func (uc *CollisionUseCases) checkBulletWallCollisions() {
	bullets := uc.bulletUseCases.GetBullets()
	level := uc.mapUseCases.GetBlocks()

	bulletIndicesToRemove, blockIndicesToRemove := uc.bulletCollisionService.CheckBulletWallCollisions(
		bullets,
		level,
	)

	// Удаляем блоки в обратном порядке (чтобы индексы не сдвигались)
	for k := len(blockIndicesToRemove) - 1; k >= 0; k-- {
		blockIndex := blockIndicesToRemove[k]
		blocks := uc.mapUseCases.GetBlocks()
		if blockIndex < len(blocks) {
			_ = uc.mapUseCases.RemoveBlock(&blocks[blockIndex])
		}
	}

	// Удаляем пули
	for _, i := range bulletIndicesToRemove {
		_ = uc.bulletUseCases.RemoveBullet(i)
	}
}

// checkBulletHQCollisions проверяет коллизии пуль с базой
func (uc *CollisionUseCases) checkBulletHQCollisions(hq *types.HQEntity) {
	if uc.hqUseCases == nil || hq == nil {
		return
	}

	bulletIndicesToRemove, _ := uc.hqUseCases.HandleBulletHit(hq)

	// Удаляем пули
	for _, i := range bulletIndicesToRemove {
		_ = uc.bulletUseCases.RemoveBullet(i)
	}
}

// checkTankBoundaryCollisions проверяет коллизии танка с границами экрана
func (uc *CollisionUseCases) checkTankBoundaryCollisions(
	tank *types.TankEntity,
) {
	if uc.boundaryCollisionService.CheckTankBoundaryCollisions(tank, false) {
		uc.tankActions.Stop(tank, false)
	}
}

// checkTankWallCollisions проверяет коллизии танка со стенами
func (uc *CollisionUseCases) checkTankWallCollisions(tank *types.TankEntity) {
	level := uc.mapUseCases.GetBlocks()
	collidingBlock := uc.wallCollisionService.CheckTankWallCollision(
		tank,
		level,
	)

	if collidingBlock != nil {
		uc.wallCollisionService.HandleTankWallCollision(tank, collidingBlock)
		uc.tankActions.Stop(tank, true)
	}
}

// CheckColliders проверяет коллизию между двумя объектами карты
func (uc *CollisionUseCases) CheckColliders(
	obj1 types.IMapObject,
	obj2 types.IMapObject,
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
	obj types.IMapObject,
	objects []types.IMapObject,
) []types.IMapObject {
	var collidingObjects []types.IMapObject

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
	obj types.IMapObject,
	objects []types.IMapObject,
) types.IMapObject {
	for _, mapObj := range objects {
		if uc.CheckColliders(obj, mapObj) {
			return mapObj
		}
	}
	return nil
}
