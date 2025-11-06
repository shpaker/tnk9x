package collision_services

import "github.com/shpaker/gonflict/internal/types"

// EntitiesCollisionService предоставляет логику проверки коллизий между сущностями
type EntitiesCollisionService struct{}

// NewEntitiesCollisionService создает новый сервис проверки коллизий между сущностями
func NewEntitiesCollisionService() *EntitiesCollisionService {
	return &EntitiesCollisionService{}
}

// CheckColliders проверяет коллизию между двумя сущностями
func (s *EntitiesCollisionService) CheckColliders(
	obj1 types.IEntityCollider,
	obj2 types.IEntityCollider,
) bool {
	// Проверяем, что объекты на одном уровне высоты
	if obj1.GetAltitude() != obj2.GetAltitude() {
		return false
	}

	pos1 := obj1.GetPosition()
	size1 := obj1.GetSize()
	pos2 := obj2.GetPosition()
	size2 := obj2.GetSize()

	// Проверяем пересечение прямоугольников
	return pos1.X < pos2.X+float64(size2.Width) &&
		pos1.X+float64(size1.Width) > pos2.X &&
		pos1.Y < pos2.Y+float64(size2.Height) &&
		pos1.Y+float64(size1.Height) > pos2.Y
}

// ResolveCollisionPosition вычисляет скорректированную позицию сущности
// при столкновении с препятствием
// Возвращает новую позицию с учетом направления движения
func (s *EntitiesCollisionService) ResolveCollisionPosition(
	entity types.IEntityCollider,
	obstacle types.IEntityCollider,
	direction types.Direction,
) types.Position {
	entityPos := entity.GetPosition()
	entitySize := entity.GetSize()
	obstaclePos := obstacle.GetPosition()
	obstacleSize := obstacle.GetSize()

	// Инициализируем позицию текущей позицией сущности
	updatedPosition := entityPos

	switch direction {
	case types.DirectionUp:
		// верх сущности упирается в низ препятствия
		updatedPosition.Y = obstaclePos.Y + float64(obstacleSize.Height)
	case types.DirectionDown:
		// низ сущности упирается в верх препятствия
		updatedPosition.Y = obstaclePos.Y - float64(entitySize.Height)
	case types.DirectionLeft:
		// левая сторона сущности упирается в правую сторону препятствия
		updatedPosition.X = obstaclePos.X + float64(obstacleSize.Width)
	case types.DirectionRight:
		// правая сторона сущности упирается в левую сторону препятствия
		updatedPosition.X = obstaclePos.X - float64(entitySize.Width)
	}

	return updatedPosition
}
