package collision_services

import "github.com/shpaker/gonflict/internal/types"

type BoundaryCollisionService struct {
	sizePx types.Size
}

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

func (s *BoundaryCollisionService) CheckLeftBoundaryCollision(
	entity types.IEntityCollider,
) bool {
	position := entity.GetPosition()
	return position.X < 0
}

func (s *BoundaryCollisionService) CheckRightBoundaryCollision(
	entity types.IEntityCollider,
) bool {
	position := entity.GetPosition()
	size := entity.GetSize()
	maxX := float64(s.sizePx.Width - size.Width)
	return position.X > maxX
}

func (s *BoundaryCollisionService) CheckTopBoundaryCollision(
	entity types.IEntityCollider,
) bool {
	position := entity.GetPosition()
	return position.Y < 0
}

func (s *BoundaryCollisionService) CheckBottomBoundaryCollision(
	entity types.IEntityCollider,
) bool {
	position := entity.GetPosition()
	size := entity.GetSize()
	maxY := float64(s.sizePx.Height - size.Height)
	return position.Y > maxY
}
