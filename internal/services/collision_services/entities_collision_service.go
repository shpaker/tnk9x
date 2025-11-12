package collision_services

import (
	"errors"

	"github.com/shpaker/gonflict/internal/types"
)

type EntitiesCollisionService struct{}

func NewEntitiesCollisionService() *EntitiesCollisionService {
	return &EntitiesCollisionService{}
}

func (s *EntitiesCollisionService) CheckColliders(
	obj1 types.IEntityCollider,
	obj2 types.IEntityCollider,
) bool {
	if obj1.GetAltitude() != obj2.GetAltitude() {
		return false
	}

	pos1 := obj1.GetPosition()
	size1 := obj1.GetSize()
	pos2 := obj2.GetPosition()
	size2 := obj2.GetSize()

	return pos1.X < pos2.X+float64(size2.Width) &&
		pos1.X+float64(size1.Width) > pos2.X &&
		pos1.Y < pos2.Y+float64(size2.Height) &&
		pos1.Y+float64(size1.Height) > pos2.Y
}

func (s *EntitiesCollisionService) isObstacleInDirection(
	entity types.IEntityCollider,
	obstacle types.IEntityCollider,
	direction types.Direction,
) bool {
	entityPos := entity.GetPosition()
	obstaclePos := obstacle.GetPosition()

	switch direction {
	case types.DirectionUp:

		return obstaclePos.Y <= entityPos.Y
	case types.DirectionDown:

		return obstaclePos.Y >= entityPos.Y
	case types.DirectionLeft:

		return obstaclePos.X <= entityPos.X
	case types.DirectionRight:

		return obstaclePos.X >= entityPos.X
	}

	return false
}

func (s *EntitiesCollisionService) calculateCorrectedPosition(
	entity types.IEntityCollider,
	obstacle types.IEntityCollider,
	direction types.Direction,
) types.Position {
	entityPos := entity.GetPosition()
	entitySize := entity.GetSize()
	obstaclePos := obstacle.GetPosition()
	obstacleSize := obstacle.GetSize()

	updatedPosition := entityPos

	switch direction {
	case types.DirectionUp:

		updatedPosition.Y = obstaclePos.Y + float64(obstacleSize.Height)
	case types.DirectionDown:

		updatedPosition.Y = obstaclePos.Y - float64(entitySize.Height)
	case types.DirectionLeft:

		updatedPosition.X = obstaclePos.X + float64(obstacleSize.Width)
	case types.DirectionRight:

		updatedPosition.X = obstaclePos.X - float64(entitySize.Width)
	}

	return updatedPosition
}

func (s *EntitiesCollisionService) ResolveCollisionPosition(
	entity types.IEntityCollider,
	obstacle types.IEntityCollider,
	direction types.Direction,
) (types.Position, error) {
	if !s.isObstacleInDirection(entity, obstacle, direction) {
		return entity.GetPosition(), errors.New(
			"препятствие не по направлению движения",
		)
	}

	return s.calculateCorrectedPosition(entity, obstacle, direction), nil
}
