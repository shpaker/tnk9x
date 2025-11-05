package services

import "github.com/shpaker/gonflict/internal/types"

// BulletCollisionService предоставляет логику обработки коллизий пуль
type BulletCollisionService struct {
	tileMinSize    int
	checkColliders func(obj1 types.IMapObject, obj2 types.IMapObject) bool
}

// NewBulletCollisionService создает новый сервис коллизий пуль
func NewBulletCollisionService(
	tileMinSize int,
	checkColliders func(obj1 types.IMapObject, obj2 types.IMapObject) bool,
) *BulletCollisionService {
	return &BulletCollisionService{
		tileMinSize:    tileMinSize,
		checkColliders: checkColliders,
	}
}

// CheckBulletWallCollisions проверяет коллизии пуль со стенами
// Возвращает список индексов пуль для удаления и список индексов блоков для удаления
func (s *BulletCollisionService) CheckBulletWallCollisions(
	bullets []types.BulletEntity,
	level []types.BlockEntity,
) (bulletIndicesToRemove []int, blockIndicesToRemove []int) {
	bulletSize := 4 // Размер пули (стандартный)

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем коллизии пули со стенами
		bulletHit := false

		for j, wall := range level {
			if wall.Data == nil {
				continue
			}

			// Преобразуем координаты блока из tile координат в пиксели
			blockWorldX := wall.Position.X * float64(s.tileMinSize)
			blockWorldY := wall.Position.Y * float64(s.tileMinSize)
			blockSize := float64(s.tileMinSize)

			// Проверяем коллизию пули с блоком напрямую по координатам
			if s.checkCollisionRectangles(
				bullet.Position.X,
				bullet.Position.Y,
				float64(bulletSize),
				float64(bulletSize),
				blockWorldX,
				blockWorldY,
				blockSize,
				blockSize,
			) && bullet.GetAltitude() == wall.Altitude {
				// Если блок - кирпичная стена, помечаем для удаления
				if wall.Data.Name == types.Brick {
					blockIndicesToRemove = append(blockIndicesToRemove, j)
				}

				bulletHit = true
				// Не прерываем цикл, продолжаем проверять другие блоки
			}
		}

		// Удаляем пулю только после проверки всех блоков
		if bulletHit {
			bulletIndicesToRemove = append(bulletIndicesToRemove, i)
		}
	}

	return bulletIndicesToRemove, blockIndicesToRemove
}

// checkCollisionRectangles проверяет коллизию двух прямоугольников
func (s *BulletCollisionService) checkCollisionRectangles(
	x1, y1, w1, h1 float64,
	x2, y2, w2, h2 float64,
) bool {
	return x1 < x2+w2 &&
		x1+w1 > x2 &&
		y1 < y2+h2 &&
		y1+h1 > y2
}

// CheckBulletTanksCollisions проверяет коллизии пуль со всеми танками (игрок + враги)
// Возвращает карту индекс_пули -> танк для обработки взрыва
// Пули врагов проходят сквозь других врагов
// Пули игрока попадают во врагов
// Пули врагов попадают в игрока
func (s *BulletCollisionService) CheckBulletTanksCollisions(
	bullets []types.BulletEntity,
	allTanks []*types.TankEntity,
	enemyTanks []*types.TankEntity,
) (bulletIndicesToRemove []int, tanksToExplode map[int]*types.TankEntity) {
	tanksToExplode = make(map[int]*types.TankEntity)

	// Создаем карту для быстрой проверки, является ли танк врагом
	enemySet := make(map[*types.TankEntity]bool)
	for _, enemy := range enemyTanks {
		if enemy != nil {
			enemySet[enemy] = true
		}
	}

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]
		if bullet.Owner == nil {
			continue
		}

		// Определяем, является ли пуля вражеской
		isEnemyBullet := enemySet[bullet.Owner]

		// Проверяем коллизию с каждым танком
		for _, tank := range allTanks {
			if tank == nil || !tank.IsActive() {
				continue
			}

			// Проверяем, что пуля не принадлежит этому танку (избегаем самоуничтожения)
			if bullet.Owner == tank {
				continue
			}

			// Пули врагов проходят сквозь других врагов
			if isEnemyBullet && enemySet[tank] {
				continue
			}

			// Проверяем коллизию между пулей и танком
			if s.checkColliders(bullet, tank) {
				bulletIndicesToRemove = append(bulletIndicesToRemove, i)
				tanksToExplode[i] = tank
				// Выходим из цикла танков, так как пуля уже обработана
				break
			}
		}
	}

	return bulletIndicesToRemove, tanksToExplode
}

// CheckBulletHQCollision проверяет коллизию пуль с базой
// Возвращает список индексов пуль для удаления и true если база была уничтожена
// Базу могут разрушать только пули врагов (определяется по списку enemyTanks)
func (s *BulletCollisionService) CheckBulletHQCollision(
	bullets []types.BulletEntity,
	hq *types.HQEntity,
	enemyTanks []*types.TankEntity,
) (bulletIndicesToRemove []int, hqDestroyed bool) {
	if hq == nil || hq.IsDestroyed() {
		return nil, false
	}

	// Создаем карту для быстрой проверки, является ли пуля вражеской
	enemySet := make(map[*types.TankEntity]bool)
	for _, enemy := range enemyTanks {
		if enemy != nil {
			enemySet[enemy] = true
		}
	}

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем, что пуля существует и имеет владельца
		if bullet.Owner == nil {
			continue
		}

		// Базу могут разрушать только пули врагов (не пули игрока)
		if !enemySet[bullet.Owner] {
			continue
		}

		// Проверяем коллизию между пулей и базой
		if s.checkColliders(bullet, hq) {
			bulletIndicesToRemove = append(bulletIndicesToRemove, i)
			hqDestroyed = true
			// Останавливаем после первого попадания (база умирает от одного попадания)
			break
		}
	}

	return bulletIndicesToRemove, hqDestroyed
}
