package services

import "github.com/shpaker/gonflict/internal/types"

// WallCollisionService предоставляет логику обработки коллизий со стенами
type WallCollisionService struct {
	tankSpriteSize    int
	tileMinSize       int
	coordinateService *CoordinateService
}

// NewWallCollisionService создает новый сервис коллизий со стенами
func NewWallCollisionService(
	tankSpriteSize int,
	tileMinSize int,
) *WallCollisionService {
	return &WallCollisionService{
		tankSpriteSize:    tankSpriteSize,
		tileMinSize:       tileMinSize,
		coordinateService: NewCoordinateService(),
	}
}

// CheckEnemyWallCollision проверяет коллизию врага со стенами
// Возвращает блок, с которым произошла коллизия, или nil
func (s *WallCollisionService) CheckEnemyWallCollision(
	enemy *types.TankEntity,
	level []types.BlockEntity,
) *types.BlockEntity {
	for i := range level {
		wall := &level[i]
		if wall.Data == nil {
			continue
		}

		// Преобразуем координаты блока из tile координат в пиксели
		blockWorldX := wall.Position.X * float64(s.tileMinSize)
		blockWorldY := wall.Position.Y * float64(s.tileMinSize)
		blockSize := s.tileMinSize

		// Проверяем коллизию напрямую по координатам
		if s.checkCollisionRectangles(
			enemy.Position.X,
			enemy.Position.Y,
			float64(s.tankSpriteSize),
			float64(s.tankSpriteSize),
			blockWorldX,
			blockWorldY,
			float64(blockSize),
			float64(blockSize),
		) && enemy.GetAltitude() == wall.Altitude {
			return wall
		}
	}
	return nil
}

// HandleEnemyWallCollision обрабатывает коллизию врага со стеной
func (s *WallCollisionService) HandleEnemyWallCollision(
	enemy *types.TankEntity,
	block *types.BlockEntity,
) {
	// Преобразуем координаты блока из tile координат в пиксели
	blockWorldX := block.Position.X * float64(s.tileMinSize)
	blockWorldY := block.Position.Y * float64(s.tileMinSize)
	blockSize := float64(s.tileMinSize)

	// Откатываем позицию врага в зависимости от направления
	switch enemy.Direction {
	case types.DirectionUp:
		enemy.Position.Y = blockWorldY + blockSize
	case types.DirectionDown:
		enemy.Position.Y = blockWorldY - float64(s.tankSpriteSize)
	case types.DirectionLeft:
		enemy.Position.X = blockWorldX + blockSize
	case types.DirectionRight:
		enemy.Position.X = blockWorldX - float64(s.tankSpriteSize)
	}

	// Останавливаем врага
	enemy.Speed = 0

	// Округляем координаты врага до ближайшего кратного 4
	enemy.Position.X = s.coordinateService.RoundToNearestMultipleOf4(
		enemy.Position.X,
	)
	enemy.Position.Y = s.coordinateService.RoundToNearestMultipleOf4(
		enemy.Position.Y,
	)
}

// CheckTankWallCollision проверяет коллизию танка со стенами
// Возвращает блок, с которым произошла коллизия, или nil
func (s *WallCollisionService) CheckTankWallCollision(
	tank *types.TankEntity,
	level []types.BlockEntity,
) *types.BlockEntity {
	for i := range level {
		wall := &level[i]
		if wall.Data == nil {
			continue
		}

		// Преобразуем координаты блока из tile координат в пиксели
		blockWorldX := wall.Position.X * float64(s.tileMinSize)
		blockWorldY := wall.Position.Y * float64(s.tileMinSize)
		blockSize := s.tileMinSize

		// Проверяем коллизию напрямую по координатам
		if s.checkCollisionRectangles(
			tank.Position.X,
			tank.Position.Y,
			float64(s.tankSpriteSize),
			float64(s.tankSpriteSize),
			blockWorldX,
			blockWorldY,
			float64(blockSize),
			float64(blockSize),
		) && tank.GetAltitude() == wall.Altitude {
			return wall
		}
	}
	return nil
}

// HandleTankWallCollision обрабатывает коллизию танка со стеной
// Корректирует позицию танка при столкновении со стеной
func (s *WallCollisionService) HandleTankWallCollision(
	tank *types.TankEntity,
	block *types.BlockEntity,
) {
	// Преобразуем координаты блока из tile координат в пиксели
	blockWorldX := block.Position.X * float64(s.tileMinSize)
	blockWorldY := block.Position.Y * float64(s.tileMinSize)
	blockSize := float64(s.tileMinSize)

	switch tank.Direction {
	case types.DirectionUp:
		// верх танка упирается в низ блока
		tank.Position.Y = blockWorldY + blockSize
	case types.DirectionDown:
		// низ танка упирается в верх блока
		tank.Position.Y = blockWorldY - float64(s.tankSpriteSize)
	case types.DirectionLeft:
		// левая сторона танка упирается в правую сторону блока
		tank.Position.X = blockWorldX + blockSize
	case types.DirectionRight:
		// правая сторона танка упирается в левую сторону блока
		tank.Position.X = blockWorldX - float64(s.tankSpriteSize)
	}
}

// checkCollisionRectangles проверяет коллизию двух прямоугольников
func (s *WallCollisionService) checkCollisionRectangles(
	x1, y1, w1, h1 float64,
	x2, y2, w2, h2 float64,
) bool {
	return x1 < x2+w2 &&
		x1+w1 > x2 &&
		y1 < y2+h2 &&
		y1+h1 > y2
}
