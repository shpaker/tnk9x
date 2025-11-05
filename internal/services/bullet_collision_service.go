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

// CheckBulletTankCollision проверяет коллизию пуль с танком
// Возвращает список индексов пуль для удаления
func (s *BulletCollisionService) CheckBulletTankCollision(
	bullets []types.BulletEntity,
	tank *types.TankEntity,
) []int {
	var indicesToRemove []int

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем коллизию между пулей и танком
		if bullet.Owner != tank && s.checkColliders(bullet, tank) {
			indicesToRemove = append(indicesToRemove, i)
		}
	}

	return indicesToRemove
}

// CheckBulletEnemyCollisions проверяет коллизии пуль с врагами
// Возвращает карту индекс_пули -> индекс_врага для обработки взрыва
func (s *BulletCollisionService) CheckBulletEnemyCollisions(
	bullets []types.BulletEntity,
	enemies []*types.TankEntity,
) (bulletIndicesToRemove []int, enemyIndicesToExplode map[int]int) {
	enemyIndicesToExplode = make(map[int]int)

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем коллизию с каждым врагом
		for enemyIndex, enemy := range enemies {
			// Пропускаем если врага нет
			if enemy == nil {
				continue
			}

			// Проверяем, что пуля не принадлежит этому танку (избегаем самоуничтожения)
			if bullet.Owner == enemy {
				continue
			}

			// Если враг активен и есть коллизия
			if enemy.IsActive() && s.checkColliders(bullet, enemy) {
				// Удаляем пулю
				bulletIndicesToRemove = append(bulletIndicesToRemove, i)
				// Запоминаем врага для взрыва
				enemyIndicesToExplode[i] = enemyIndex
				// Выходим из цикла врагов, так как пуля уже обработана
				break
			}
		}
	}

	return bulletIndicesToRemove, enemyIndicesToExplode
}

// CheckBulletHQCollision проверяет коллизию пуль с базой
// Возвращает список индексов пуль для удаления и true если база была уничтожена
func (s *BulletCollisionService) CheckBulletHQCollision(
	bullets []types.BulletEntity,
	hq *types.HQEntity,
) (bulletIndicesToRemove []int, hqDestroyed bool) {
	if hq == nil || hq.IsDestroyed() {
		return nil, false
	}

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем, что пуля существует и имеет владельца
		if bullet.Owner == nil {
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
