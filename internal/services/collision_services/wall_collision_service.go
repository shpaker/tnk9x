package collision_services

import (
	"github.com/shpaker/tnk25/internal/interfaces"
	"github.com/shpaker/tnk25/internal/types"
)

type WallCollisionService struct {
	tankSpriteSize           int
	tileMinSize              int
	coordinateService        interfaces.ICoordinateService
	entitiesCollisionService interfaces.IEntitiesCollisionService
}

func NewWallCollisionService(
	tankSpriteSize int,
	tileMinSize int,
	coordinateService interfaces.ICoordinateService,
	entitiesCollisionService interfaces.IEntitiesCollisionService,
) *WallCollisionService {
	return &WallCollisionService{
		tankSpriteSize:           tankSpriteSize,
		tileMinSize:              tileMinSize,
		coordinateService:        coordinateService,
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
