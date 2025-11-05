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
// allTanks - все танки (игрок + враги) без разделения
// enemyTanks - список врагов, используется только для определения принадлежности пули (пули врагов проходят сквозь врагов)
func (uc *CollisionUseCases) UpdateCollisions(
	allTanks []*types.TankEntity,
	enemyTanks []*types.TankEntity,
	hq *types.HQEntity,
) error {
	if len(allTanks) == 0 {
		return nil
	}

	// Проверяем коллизии с HQ ПЕРВЫМИ, чтобы пули не удалялись другими проверками
	uc.checkBulletHQCollisions(hq, enemyTanks)
	uc.checkBulletBoundaryCollisions()

	// Проверяем коллизии пуль со всеми танками унифицированно
	uc.checkBulletTanksCollisions(allTanks, enemyTanks)

	uc.checkBulletWallCollisions()

	// Проверяем коллизии всех танков с границами и стенами унифицированно
	uc.checkTanksCollisions(allTanks)

	return nil
}

// checkTanksCollisions проверяет коллизии всех танков с границами и стенами унифицированно
func (uc *CollisionUseCases) checkTanksCollisions(
	allTanks []*types.TankEntity,
) {
	level := uc.mapUseCases.GetBlocks()

	for _, tank := range allTanks {
		if tank == nil || !tank.IsActive() {
			continue
		}

		// Проверяем коллизии с границами
		hadBoundaryCollision := uc.boundaryCollisionService.CheckTankBoundaryCollisions(
			tank,
			false,
		)
		if hadBoundaryCollision {
			uc.tankActions.Stop(tank, false)
		}

		// Проверяем коллизии со стенами
		collidingBlock := uc.wallCollisionService.CheckTankWallCollision(
			tank,
			level,
		)

		if collidingBlock != nil {
			uc.wallCollisionService.HandleTankWallCollision(
				tank,
				collidingBlock,
			)
			uc.tankActions.Stop(tank, true)
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

// checkBulletTanksCollisions проверяет коллизии пуль со всеми танками унифицированно
func (uc *CollisionUseCases) checkBulletTanksCollisions(
	allTanks []*types.TankEntity,
	enemyTanks []*types.TankEntity,
) {
	bullets := uc.bulletUseCases.GetBullets()
	bulletIndicesToRemove, tanksToExplode := uc.bulletCollisionService.CheckBulletTanksCollisions(
		bullets,
		allTanks,
		enemyTanks,
	)

	for _, i := range bulletIndicesToRemove {
		_ = uc.bulletUseCases.RemoveBullet(i)

		// Взрываем танк, в который попала пуля
		if tank, exists := tanksToExplode[i]; exists && tank != nil {
			_ = uc.tankLifecycleUseCases.Explode(tank)
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
func (uc *CollisionUseCases) checkBulletHQCollisions(
	hq *types.HQEntity,
	enemyTanks []*types.TankEntity,
) {
	if hq == nil || hq.IsDestroyed() || hq.State == types.HQStateExploding {
		return
	}

	bullets := uc.bulletUseCases.GetBullets()
	bulletIndicesToRemove, destroyed := uc.bulletCollisionService.CheckBulletHQCollision(
		bullets,
		hq,
		enemyTanks,
	)

	// Запускаем анимацию взрыва если нужно
	if destroyed && uc.hqUseCases != nil {
		_ = uc.hqUseCases.Explode(hq)
	}

	// Удаляем пули
	for _, i := range bulletIndicesToRemove {
		_ = uc.bulletUseCases.RemoveBullet(i)
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
