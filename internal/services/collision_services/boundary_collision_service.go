package collision_services

import "github.com/shpaker/gonflict/internal/types"

// BoundaryCollisionService предоставляет логику обработки коллизий с границами экрана
type BoundaryCollisionService struct {
	sizePx types.Size
}

// NewBoundaryCollisionService создает новый сервис коллизий с границами
func NewBoundaryCollisionService(
	mapWidthHeight int,
	tankSpriteSize int,
) *BoundaryCollisionService {
	return &BoundaryCollisionService{
		sizePx: types.Size{
			Width:  mapWidthHeight,
			Height: mapWidthHeight,
		},
	}
}

// CheckLeftBoundaryCollision проверяет коллизию сущности с левой границей экрана
// Возвращает true, если была коллизия
func (s *BoundaryCollisionService) CheckLeftBoundaryCollision(
	entity types.IEntityCollider,
) bool {
	position := entity.GetPosition()
	return position.X < 0
}

// CheckRightBoundaryCollision проверяет коллизию сущности с правой границей экрана
// Возвращает true, если была коллизия
func (s *BoundaryCollisionService) CheckRightBoundaryCollision(
	entity types.IEntityCollider,
) bool {
	position := entity.GetPosition()
	size := entity.GetSize()
	maxX := float64(s.sizePx.Width - size.Width)
	return position.X > maxX
}

// CheckTopBoundaryCollision проверяет коллизию сущности с верхней границей экрана
// Возвращает true, если была коллизия
func (s *BoundaryCollisionService) CheckTopBoundaryCollision(
	entity types.IEntityCollider,
) bool {
	position := entity.GetPosition()
	return position.Y < 0
}

// CheckBottomBoundaryCollision проверяет коллизию сущности с нижней границей экрана
// Возвращает true, если была коллизия
func (s *BoundaryCollisionService) CheckBottomBoundaryCollision(
	entity types.IEntityCollider,
) bool {
	position := entity.GetPosition()
	size := entity.GetSize()
	maxY := float64(s.sizePx.Height - size.Height)
	return position.Y > maxY
}
