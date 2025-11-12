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
func (uc *CollisionUseCases) UpdateCollisions() {
	// Получаем все танки из репозитория через use cases
	allTanks := uc.tankCommonUseCases.GetAllTanks()

	// Получаем все пули из репозитория
	bullets := uc.bulletUseCases.GetBullets()

	// Получаем HQ из use cases
	hq := uc.hqUseCases.GetHQ()

	// Получаем блоки карты из use cases
	mapBlocks := uc.mapUseCases.GetBlocks()

	// Проверяем коллизии с HQ ПЕРВЫМИ, чтобы пули не удалялись другими проверками
	uc.checkBulletsCollisions(bullets, hq, mapBlocks)
	// После удаления пуль обновляем срез
	bullets = uc.bulletUseCases.GetBullets()
	uc.checkTanksCollisions(allTanks, bullets, mapBlocks)
}

// checkTanksCollisions проверяет коллизии всех танков с границами и стенами унифицированно
func (uc *CollisionUseCases) checkTanksCollisions(
	allTanks []*types.TankEntity,
	bullets []*types.BulletEntity,
	mapBlocks types.MapBlocks,
) {
	for _, tank := range allTanks {
		// Проверяем коллизии с границами карты
		if !tank.IsActive() {
			continue
		}
		uc.checkTankBoundaryCollision(tank)
		uc.checkTankBlockCollisions(tank, mapBlocks)
		uc.checkTankTankCollision(tank, allTanks)
		uc.checkTankBulletCollisions(tank, bullets)
	}
}

func (uc *CollisionUseCases) checkBulletsCollisions(
	bullets []*types.BulletEntity,
	hq *types.HQEntity,
	mapBlocks types.MapBlocks,
) {
	for index := 0; index < len(bullets); index++ {
		bullet := bullets[index]
		if bullet == nil {
			continue
		}
		if uc.checkBulletBoundaryCollision(bullet) ||
			uc.checkBulletHQCollision(bullet, hq) ||
			uc.checkBulletWallCollision(bullet, mapBlocks) {
			_ = uc.bulletUseCases.RemoveBullet(index)
			bullets = uc.bulletUseCases.GetBullets()
			index--
			continue
		}

		for j := index + 1; j < len(bullets); j++ {
			other := bullets[j]
			if uc.checkBulletBulletCollision(bullet, other) {
				_ = uc.bulletUseCases.RemoveBullet(j)
				_ = uc.bulletUseCases.RemoveBullet(index)
				bullets = uc.bulletUseCases.GetBullets()
				index--
				break
			}
		}
	}
}

func (uc *CollisionUseCases) checkBulletBulletCollision(
	first *types.BulletEntity,
	second *types.BulletEntity,
) bool {
	if first == nil || second == nil {
		return false
	}

	return uc.entitiesCollisionService.CheckColliders(first, second)
}

// checkTankBulletCollisions проверяет коллизии танка с пулями
func (uc *CollisionUseCases) checkTankBulletCollisions(
	tank *types.TankEntity,
	bullets []*types.BulletEntity,
) {
	for index, bullet := range bullets {
		if bullet == nil {
			continue
		}
		if uc.bulletCollisionService.CheckBulletTankCollision(
			bullet,
			tank,
		) {
			_ = uc.bulletUseCases.RemoveBullet(index)
			if tank.IsActive() {
				_ = uc.tankLifecycleUseCases.Explode(tank)
			}
			return
		}
	}
}

// HasTankCollision проверяет, пересекается ли переданный танк с любым другим танком
func (uc *CollisionUseCases) HasTankCollision(
	candidate *types.TankEntity,
) bool {
	if candidate == nil {
		return false
	}

	allTanks := uc.tankCommonUseCases.GetAllTanks()
	for _, otherTank := range allTanks {
		if otherTank == nil {
			continue
		}

		if uc.entitiesCollisionService.CheckColliders(candidate, otherTank) {
			return true
		}
	}

	return false
}

// IsSpawnerBlocked проверяет, занята ли позиция спавнера другим танком
func (uc *CollisionUseCases) IsSpawnerBlocked(
	position types.Position,
	size types.Size,
) bool {
	if uc.entitiesCollisionService == nil ||
		uc.tankCommonUseCases == nil ||
		size.Width == 0 ||
		size.Height == 0 {
		return false
	}

	candidate := types.NewDefaultTankEntity(
		types.TankRoleEnemy,
		types.DirectionUp,
	)
	candidate.Size = size
	candidate.Position = types.Position{
		X: position.X * float64(size.Width),
		Y: position.Y * float64(size.Height),
	}

	for _, otherTank := range uc.tankCommonUseCases.GetAllTanks() {
		if otherTank == nil {
			continue
		}

		if uc.entitiesCollisionService.CheckColliders(&candidate, otherTank) {
			return true
		}
	}

	return false
}

// checkTankBlockCollisions проверяет коллизии танка с блоками карты
func (uc *CollisionUseCases) checkTankBlockCollisions(
	tank *types.TankEntity,
	mapBlocks types.MapBlocks,
) {
	for _, block := range mapBlocks {
		if uc.wallCollisionService.CheckTankWallCollision(tank, block) {
			// Вычисляем скорректированную позицию через EntitiesCollisionService
			correctedPos, err := uc.entitiesCollisionService.ResolveCollisionPosition(
				tank,
				block,
				tank.Direction,
			)
			// Если препятствие не по направлению движения, пропускаем резолв
			if err != nil {
				continue
			}
			// Применяем скорректированную позицию
			tank.Position = correctedPos
			uc.tankActions.Stop(tank, true)
			return
		}
	}
}

// checkTankTankCollision проверяет коллизию танка с другими танками
func (uc *CollisionUseCases) checkTankTankCollision(
	tank *types.TankEntity,
	allTanks []*types.TankEntity,
) {
	for _, otherTank := range allTanks {
		// Пропускаем проверку с самим собой и неактивными танками
		if tank == otherTank || otherTank.IsDestroyed() {
			continue
		}

		// Проверяем коллизию
		if uc.entitiesCollisionService.CheckColliders(tank, otherTank) {
			// Вычисляем скорректированную позицию
			correctedPos, err := uc.entitiesCollisionService.ResolveCollisionPosition(
				tank,
				otherTank,
				tank.Direction,
			)
			// Если препятствие не по направлению движения, пропускаем резолв
			if err != nil {
				continue
			}
			// Применяем скорректированную позицию
			tank.Position = correctedPos
			uc.tankActions.Stop(tank, true)
			uc.tankActions.Stop(otherTank, true)
			return
		}
	}
}

// checkTankBoundaryCollision проверяет коллизию танка с краями карты
func (uc *CollisionUseCases) checkTankBoundaryCollision(
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

// checkBulletBoundaryCollision проверяет коллизию пули с границами экрана
func (uc *CollisionUseCases) checkBulletBoundaryCollision(
	bullet *types.BulletEntity,
) bool {
	if uc.boundaryCollisionService.CheckLeftBoundaryCollision(bullet) ||
		uc.boundaryCollisionService.CheckRightBoundaryCollision(bullet) ||
		uc.boundaryCollisionService.CheckTopBoundaryCollision(bullet) ||
		uc.boundaryCollisionService.CheckBottomBoundaryCollision(bullet) {
		return true
	}
	return false
}

// checkBulletWallCollision проверяет коллизию пули со стенами
// Возвращает true, если пуля попала в блок и должна быть удалена
func (uc *CollisionUseCases) checkBulletWallCollision(
	bullet *types.BulletEntity,
	mapBlocks types.MapBlocks,
) bool {
	// Проверяем коллизию пули с каждым блоком
	for _, block := range mapBlocks {
		if uc.bulletCollisionService.CheckBulletBlockCollision(
			bullet,
			block,
		) {
			// Если блок - кирпичная стена, удаляем его
			if block.Data != nil && block.Data.Name == types.Brick {
				_ = uc.mapUseCases.RemoveBlock(block)
			}

			// Пуля попала в блок, нужно удалить пулю
			return true
		}
	}

	// Коллизии не было
	return false
}

// checkBulletHQCollision проверяет коллизию пули с базой
// Возвращает true, если пуля попала в базу и должна быть удалена
func (uc *CollisionUseCases) checkBulletHQCollision(
	bullet *types.BulletEntity,
	hq *types.HQEntity,
) bool {
	// Проверяем коллизию пули с базой
	// Базу могут разрушать только пули врагов (не пули игрока)
	if uc.bulletCollisionService.CheckBulletHQCollision(bullet, hq) {
		// Запускаем анимацию взрыва базы
		if uc.hqUseCases != nil && !hq.IsDestroyed() {
			_ = uc.hqUseCases.Explode(hq)
		}
		return true
	}
	return false
}
