package services

import "github.com/shpaker/gonflict/internal/types"

// BoundaryCollisionService предоставляет логику обработки коллизий с границами экрана
type BoundaryCollisionService struct {
	mapWidthHeight    int
	tankSpriteSize    int
	coordinateService *CoordinateService
}

// NewBoundaryCollisionService создает новый сервис коллизий с границами
func NewBoundaryCollisionService(
	mapWidthHeight int,
	tankSpriteSize int,
) *BoundaryCollisionService {
	return &BoundaryCollisionService{
		mapWidthHeight:    mapWidthHeight,
		tankSpriteSize:    tankSpriteSize,
		coordinateService: NewCoordinateService(),
	}
}

// CheckTankBoundaryCollisions проверяет коллизии танка/врага с границами экрана
// Корректирует позицию при выходе за границы
// stopAndRound: если true, останавливает (Speed = 0) и округляет координаты до кратного 4
// Возвращает true, если была коллизия
func (s *BoundaryCollisionService) CheckTankBoundaryCollisions(
	tank *types.TankEntity,
	stopAndRound bool,
) bool {
	collision := false

	// Откатываем позицию при выходе за границы
	if tank.Position.X < 0 {
		tank.Position.X = 0
		collision = true
		if stopAndRound {
			tank.Speed = 0
		}
	}
	if tank.Position.Y < 0 {
		tank.Position.Y = 0
		collision = true
		if stopAndRound {
			tank.Speed = 0
		}
	}
	if tank.Position.X > float64(s.mapWidthHeight-s.tankSpriteSize) {
		tank.Position.X = float64(s.mapWidthHeight - s.tankSpriteSize)
		collision = true
		if stopAndRound {
			tank.Speed = 0
		}
	}
	if tank.Position.Y > float64(s.mapWidthHeight-s.tankSpriteSize) {
		tank.Position.Y = float64(s.mapWidthHeight - s.tankSpriteSize)
		collision = true
		if stopAndRound {
			tank.Speed = 0
		}
	}

	// Округляем координаты до ближайшего кратного 4 (для врагов или если остановлен)
	if stopAndRound && tank.Speed == 0 {
		tank.Position.X = s.coordinateService.RoundToNearestMultipleOf4(
			tank.Position.X,
		)
		tank.Position.Y = s.coordinateService.RoundToNearestMultipleOf4(
			tank.Position.Y,
		)
	}

	return collision
}

// CheckEnemyBoundaryCollisions проверяет коллизии врага с границами экрана
func (s *BoundaryCollisionService) CheckEnemyBoundaryCollisions(
	enemy *types.TankEntity,
) {
	s.CheckTankBoundaryCollisions(enemy, true)
}

// CheckBulletBoundaryCollisions проверяет коллизии пуль с границами экрана
func (s *BoundaryCollisionService) CheckBulletBoundaryCollisions(
	bullets []types.BulletEntity,
) []int {
	var indicesToRemove []int

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Удаляем пули, которые вышли за границы экрана
		if bullet.Position.X < 0 ||
			bullet.Position.X > float64(s.mapWidthHeight) ||
			bullet.Position.Y < 0 ||
			bullet.Position.Y > float64(s.mapWidthHeight) {
			indicesToRemove = append(indicesToRemove, i)
		}
	}

	return indicesToRemove
}
