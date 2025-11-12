package collision_services

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// BulletCollisionService предоставляет логику обработки коллизий пуль
type BulletCollisionService struct {
	tileMinSize              int
	entitiesCollisionService interfaces.IEntitiesCollisionService
}

// NewBulletCollisionService создает новый сервис коллизий пуль
func NewBulletCollisionService(
	tileMinSize int,
	entitiesCollisionService interfaces.IEntitiesCollisionService,
) *BulletCollisionService {
	return &BulletCollisionService{
		tileMinSize:              tileMinSize,
		entitiesCollisionService: entitiesCollisionService,
	}
}

// CheckBulletBlockCollision проверяет коллизию пули с блоком
// Возвращает true, если была коллизия
func (s *BulletCollisionService) CheckBulletBlockCollision(
	bullet *types.BulletEntity,
	block *types.BlockEntity,
) bool {
	if block.Data == nil {
		return false
	}
	return s.entitiesCollisionService.CheckColliders(bullet, block)
}

// CheckBulletTankCollision проверяет коллизию пули с танком
// Возвращает true, если была коллизия
func (s *BulletCollisionService) CheckBulletTankCollision(
	bullet *types.BulletEntity,
	tank *types.TankEntity,
) bool {
	owner := bullet.GetOwner()
	if tank == nil || owner == nil ||
		owner == tank {
		return false
	}
	// Пули врагов проходят сквозь врагов, пули игрока проходят сквозь игрока
	if tank.IsEnemy() && owner.IsEnemy() == tank.IsEnemy() {
		return false
	}
	return s.entitiesCollisionService.CheckColliders(bullet, tank)
}

// CheckBulletHQCollision проверяет коллизию пули с базой
// Возвращает true, если была коллизия
func (s *BulletCollisionService) CheckBulletHQCollision(
	bullet *types.BulletEntity,
	hq *types.HQEntity,
) bool {
	if hq == nil || bullet.GetOwner() == nil {
		return false
	}
	return s.entitiesCollisionService.CheckColliders(bullet, hq)
}
