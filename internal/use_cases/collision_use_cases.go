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
	entitiesCollisionService interfaces.IEntitiesCollisionService
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
	entitiesCollisionService interfaces.IEntitiesCollisionService,
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
		entitiesCollisionService: entitiesCollisionService,
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

	// Проверяем коллизии пуль с границами экрана
	bullets := uc.bulletUseCases.GetBullets()

	uc.checkBulletWallCollisions()

	uc.checkBulletsCollisions(bullets)
	uc.checkTanksCollisions(allTanks, bullets)

	return nil
}

// checkTanksCollisions проверяет коллизии всех танков с границами и стенами унифицированно
func (uc *CollisionUseCases) checkTanksCollisions(
	allTanks []*types.TankEntity,
	bullets []types.BulletEntity,
) {
	level := uc.mapUseCases.GetBlocks()

	for _, tank := range allTanks {
		if tank == nil || !tank.IsActive() {
			continue
		}

		// Проверяем коллизии с границами карты
		uc.checkTankBoundaryCollisions(tank)

		for index, bullet := range bullets {
			if uc.bulletCollisionService.CheckTank(tank, &bullet) {
				_ = uc.bulletUseCases.RemoveBullet(index)
				_ = uc.tankLifecycleUseCases.Explode(tank)
				continue
			}
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

func (uc *CollisionUseCases) checkBulletsCollisions(
	bullets []types.BulletEntity,
) {
	for index, bullet := range bullets {
		uc.checkBulletBoundaryCollisions(&bullet, index)
	}
}

// checkTankBoundaryCollisions проверяет коллизию танка с краями карты
func (uc *CollisionUseCases) checkTankBoundaryCollisions(
	tank *types.TankEntity,
) {
	if uc.boundaryCollisionService.CheckLeftBoundaryCollision(tank) {
		uc.tankActions.SetMinXPosition(tank)
	}
	if uc.boundaryCollisionService.CheckRightBoundaryCollision(tank) {
		uc.tankActions.SetMaxXPosition(tank)
	}
	if uc.boundaryCollisionService.CheckTopBoundaryCollision(tank) {
		uc.tankActions.SetMinYPosition(tank)
	}
	if uc.boundaryCollisionService.CheckBottomBoundaryCollision(tank) {
		uc.tankActions.SetMaxYPosition(tank)
	}
}

// checkBulletBoundaryCollisions проверяет коллизии пуль с границами экрана
// nolint:unused // Метод оставлен для возможного использования в будущем
func (uc *CollisionUseCases) checkBulletBoundaryCollisions(
	bullet *types.BulletEntity,
	index int,
) {
	if uc.boundaryCollisionService.CheckLeftBoundaryCollision(bullet) ||
		uc.boundaryCollisionService.CheckRightBoundaryCollision(bullet) ||
		uc.boundaryCollisionService.CheckTopBoundaryCollision(bullet) ||
		uc.boundaryCollisionService.CheckBottomBoundaryCollision(bullet) {
		_ = uc.bulletUseCases.RemoveBullet(index)
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

	// Создаем карту для быстрой проверки, является ли пуля вражеской
	enemySet := make(map[*types.TankEntity]bool)
	for _, enemy := range enemyTanks {
		if enemy != nil {
			enemySet[enemy] = true
		}
	}

	// Проверяем коллизию пуль с базой
	// Базу могут разрушать только пули врагов (не пули игрока)
	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем, что пуля существует и является вражеской
		if bullet.Owner == nil || !enemySet[bullet.Owner] {
			continue
		}

		// Проверяем коллизию конкретной пули с базой
		singleBulletList := []types.BulletEntity{bullets[i]}
		if uc.bulletCollisionService.CheckHQ(hq, singleBulletList) {
			// Удаляем пулю
			_ = uc.bulletUseCases.RemoveBullet(i)

			// Запускаем анимацию взрыва базы
			if uc.hqUseCases != nil {
				_ = uc.hqUseCases.Explode(hq)
			}

			// Останавливаем после первого попадания (база умирает от одного попадания)
			break
		}
	}
}
