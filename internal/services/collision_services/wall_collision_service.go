package collision_services

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// WallCollisionService предоставляет логику обработки коллизий со стенами
type WallCollisionService struct {
	tankSpriteSize           int
	tileMinSize              int
	coordinateService        interfaces.ICoordinateService
	entitiesCollisionService interfaces.IEntitiesCollisionService
}

// NewWallCollisionService создает новый сервис коллизий со стенами
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

// CheckTankWallCollision проверяет коллизию танка со стенами
// Возвращает true, если была коллизия
func (s *WallCollisionService) CheckTankWallCollision(
	tank *types.TankEntity,
	wall *types.BlockEntity,
) bool {
	if wall.Data == nil {
		return false
	}
	// Блоки уже хранят позиции и размеры в пикселях, проверяем коллизию напрямую
	return s.entitiesCollisionService.CheckColliders(tank, wall)
}
