package collision_services

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ISpawnCollisionService = (*SpawnCollisionService)(nil)

type SpawnCollisionService struct{}

func NewSpawnCollisionService() *SpawnCollisionService {
	return &SpawnCollisionService{}
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

	spawner := types.Position{
		X: position.X * float64(size.Width),
		Y: position.Y * float64(size.Height),
	}
	for _, tank := range tanks {
		if tank == nil || tank.IsDestroyed() {
			continue
		}
		if types.RectsOverlap(spawner, size, tank.Position, tank.GetSize()) {
			return true
		}
	}
	return false
}
