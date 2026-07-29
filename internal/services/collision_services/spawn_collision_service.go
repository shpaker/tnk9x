package collision_services

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ISpawnCollisionService = (*SpawnCollisionService)(nil)

type SpawnCollisionService struct {
	entitiesCollisionService interfaces.IEntitiesCollisionService
}

func NewSpawnCollisionService(
	entitiesCollisionService interfaces.IEntitiesCollisionService,
) *SpawnCollisionService {
	return &SpawnCollisionService{
		entitiesCollisionService: entitiesCollisionService,
	}
}

// IsSpawnerBlocked проверяет, перекрыт ли спавнер одним из живых танков.
// Позиция спавнера задаётся в клетках и масштабируется размером танка.
func (s *SpawnCollisionService) IsSpawnerBlocked(
	position types.Position,
	size types.Size,
	tanks []*types.TankEntity,
) bool {
	if size.Width == 0 || size.Height == 0 {
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

	for _, otherTank := range tanks {
		if otherTank == nil || otherTank.IsDestroyed() {
			continue
		}

		if s.entitiesCollisionService.CheckColliders(&candidate, otherTank) {
			return true
		}
	}

	return false
}
