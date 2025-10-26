package use_cases

import (
	"log"

	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/types"
)

// BulletUseCases реализация интерфейса BulletUseCases
type BulletUseCases struct {
	bulletsRepo   game.IBulletsRepository
	tilesUseCases ITilesUseCases
}

// NewBulletUseCases создает новый экземпляр BulletUseCases
func NewBulletUseCases(bulletsRepo game.IBulletsRepository, tilesUseCases ITilesUseCases) *BulletUseCases {
	return &BulletUseCases{
		bulletsRepo:   bulletsRepo,
		tilesUseCases: tilesUseCases,
	}
}

// ShootBullet создает новую пулю от указанного танка
func (uc *BulletUseCases) ShootBullet(tank *types.TankEntity) error {
	// Проверяем, заспавнен ли танк
	if !tank.IsSpawned {
		log.Printf("DEBUG: Cannot shoot - tank is not spawned yet")
		return nil // Просто игнорируем выстрел
	}

	log.Printf("DEBUG: ShootBullet called for tank at position (%.2f, %.2f) direction %s",
		tank.WorldPosition.X, tank.WorldPosition.Y, tank.Direction)

	// Создаем тайл для пули с ID "bullet"
	bulletImageGetter, err := uc.tilesUseCases.CreateStaticTile("bullet")
	if err != nil {
		log.Printf("ERROR: Failed to create bullet tile: %v", err)
		return err
	}
	log.Printf("DEBUG: Created bullet tile successfully")

	// Вычисляем позицию пули в зависимости от направления танка
	bulletX := tank.WorldPosition.X + TankSpriteSize/2 - 2
	bulletY := tank.WorldPosition.Y + TankSpriteSize/2 - 2

	// Корректируем позицию в зависимости от направления
	switch tank.Direction {
	case types.DirectionUp:
		bulletY = tank.WorldPosition.Y - 4
	case types.DirectionDown:
		bulletY = tank.WorldPosition.Y + TankSpriteSize/2
	case types.DirectionLeft:
		bulletX = tank.WorldPosition.X - 4
	case types.DirectionRight:
		bulletX = tank.WorldPosition.X + TankSpriteSize/2
	}

	bullet := types.BulletEntity{
		ImageGetter: bulletImageGetter,
		WorldPosition: types.Position{
			X: bulletX,
			Y: bulletY,
		},
		Speed:     120.0, // Скорость пули
		Direction: tank.Direction,
		Owner:     tank,
		Altitude:  types.SURFACE, // Пули на уровне поверхности
	}

	log.Printf("DEBUG: Created bullet at position (%.2f, %.2f) direction %s",
		bullet.WorldPosition.X, bullet.WorldPosition.Y, bullet.Direction)

	uc.bulletsRepo.AddBullet(bullet)
	log.Printf("DEBUG: Bullet added to repository")
	return nil
}

// UpdateBullets обновляет позиции всех пуль
func (uc *BulletUseCases) UpdateBullets(dt float64) error {
	bullets := uc.bulletsRepo.GetAllBullets()
	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := &bullets[i]

		// Вычисляем новую позицию
		delta := bullet.Speed * dt
		switch bullet.Direction {
		case types.DirectionUp:
			bullet.WorldPosition.Y -= delta
		case types.DirectionDown:
			bullet.WorldPosition.Y += delta
		case types.DirectionLeft:
			bullet.WorldPosition.X -= delta
		case types.DirectionRight:
			bullet.WorldPosition.X += delta
		}
	}
	return nil
}

// GetBullets возвращает все активные пули
func (uc *BulletUseCases) GetBullets() []types.BulletEntity {
	return uc.bulletsRepo.GetAllBullets()
}

// RemoveBullet удаляет пулю по индексу
func (uc *BulletUseCases) RemoveBullet(index int) error {
	uc.bulletsRepo.RemoveBullet(index)
	return nil
}
