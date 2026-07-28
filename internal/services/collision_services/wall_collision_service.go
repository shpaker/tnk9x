package collision_services

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IWallCollisionService = (*WallCollisionService)(nil)

type WallCollisionService struct {
	entitiesCollisionService interfaces.IEntitiesCollisionService
}

func NewWallCollisionService(
	entitiesCollisionService interfaces.IEntitiesCollisionService,
) *WallCollisionService {
	return &WallCollisionService{
		entitiesCollisionService: entitiesCollisionService,
	}
}

func (s *WallCollisionService) CheckTankWallCollision(
	tank *types.TankEntity,
	wall *types.BlockEntity,
) bool {
	if wall.Data == nil {
		return false
	}

	return s.entitiesCollisionService.CheckColliders(tank, wall)
}
