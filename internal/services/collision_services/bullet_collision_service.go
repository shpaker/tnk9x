package collision_services

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// BulletCollisionService предоставляет логику обработки коллизий пуль
type BulletCollisionService struct {
	tileMinSize              int
	entitiesCollisionService interfaces.IEntitiesCollisionService
}

// NewBulletCollisionService создает новый сервис коллизий пуль
func NewBulletCollisionService(
	tileMinSize int,
	entitiesCollisionService interfaces.IEntitiesCollisionService,
) *BulletCollisionService {
	return &BulletCollisionService{
		tileMinSize:              tileMinSize,
		entitiesCollisionService: entitiesCollisionService,
	}
}

// CheckBulletWallCollisions проверяет коллизии пуль со стенами
// Возвращает список индексов пуль для удаления и список индексов блоков для удаления
func (s *BulletCollisionService) CheckBulletWallCollisions(
	bullets []types.BulletEntity,
	level []types.BlockEntity,
) (bulletIndicesToRemove []int, blockIndicesToRemove []int) {
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

			// Создаем временный блок для проверки коллизии с правильными мировыми координатами
			tempBlock := types.BlockEntity{
				Position: types.Position{X: blockWorldX, Y: blockWorldY},
				Size: types.Size{
					Width:  int(blockSize),
					Height: int(blockSize),
				},
				Altitude: wall.Altitude,
			}

			// Проверяем коллизию пули с блоком через EntitiesCollisionService
			if s.entitiesCollisionService.CheckColliders(bullet, &tempBlock) {
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

// CheckTank проверяет коллизии конкретного танка с пулями
// Возвращает true, если была коллизия
func (s *BulletCollisionService) CheckTank(
	tank *types.TankEntity,
	bullet *types.BulletEntity,
) bool {
	if tank == nil || !tank.IsActive() || bullet.Owner == nil ||
		bullet.Owner == tank {
		return false
	}
	return s.entitiesCollisionService.CheckColliders(bullet, tank)
}

// CheckHQ проверяет коллизию пуль с базой
// Возвращает true, если была коллизия
func (s *BulletCollisionService) CheckHQ(
	hq *types.HQEntity,
	bullets []types.BulletEntity,
) bool {
	if hq == nil || hq.IsDestroyed() {
		return false
	}

	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Проверяем, что пуля существует
		if bullet.Owner == nil {
			continue
		}

		// Проверяем коллизию между пулей и базой
		if s.entitiesCollisionService.CheckColliders(bullet, hq) {
			return true
		}
	}

	return false
}
