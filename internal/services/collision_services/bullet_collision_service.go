package collision_services

import (
	"github.com/shpaker/tnk25/internal/interfaces"
	"github.com/shpaker/tnk25/internal/types"
)

type BulletCollisionService struct {
	tileMinSize              int
	entitiesCollisionService interfaces.IEntitiesCollisionService
}

func NewBulletCollisionService(
	tileMinSize int,
	entitiesCollisionService interfaces.IEntitiesCollisionService,
) *BulletCollisionService {
	return &BulletCollisionService{
		tileMinSize:              tileMinSize,
		entitiesCollisionService: entitiesCollisionService,
	}
}

func (s *BulletCollisionService) CheckBulletBlockCollision(
	bullet *types.BulletEntity,
	block *types.BlockEntity,
) bool {
	if block.Data == nil {
		return false
	}
	return s.entitiesCollisionService.CheckColliders(bullet, block)
}

func (s *BulletCollisionService) CheckBulletTankCollision(
	bullet *types.BulletEntity,
	tank *types.TankEntity,
) bool {
	owner := bullet.GetOwner()
	if tank == nil || owner == nil ||
		owner == tank {
		return false
	}

	if tank.IsEnemy() && owner.IsEnemy() == tank.IsEnemy() {
		return false
	}
	return s.entitiesCollisionService.CheckColliders(bullet, tank)
}

func (s *BulletCollisionService) CheckBulletHQCollision(
	bullet *types.BulletEntity,
	hq *types.HQEntity,
) bool {
	if hq == nil || bullet.GetOwner() == nil {
		return false
	}
	return s.entitiesCollisionService.CheckColliders(bullet, hq)
}
