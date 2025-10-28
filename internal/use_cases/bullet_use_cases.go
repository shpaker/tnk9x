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
	// Проверяем, активен ли танк
	if !tank.IsActive() {
		log.Printf("DEBUG: Cannot shoot - tank is not active")
		return nil // Просто игнорируем выстрел
	}

	log.Printf("DEBUG: ShootBullet called for tank at position (%.2f, %.2f) direction %d",
		tank.Position.X, tank.Position.Y, tank.Direction)

	// Создаем тайл для пули с ID "bullet"
	bulletImageGetter, err := uc.tilesUseCases.CreateStaticTile("bullet")
	if err != nil {
		log.Printf("ERROR: Failed to create bullet tile: %v", err)
		return err
	}
	log.Printf("DEBUG: Created bullet tile successfully")

	// Вычисляем позицию пули в зависимости от направления танка
	bulletX := tank.Position.X + TankSpriteSize/2 - 2
	bulletY := tank.Position.Y + TankSpriteSize/2 - 2

	// Корректируем позицию в зависимости от направления
	switch tank.Direction {
	case types.DirectionUp:
		bulletY = tank.Position.Y - 4
	case types.DirectionDown:
		bulletY = tank.Position.Y + TankSpriteSize/2
	case types.DirectionLeft:
		bulletX = tank.Position.X - 4
	case types.DirectionRight:
		bulletX = tank.Position.X + TankSpriteSize/2
	}

	bullet := types.BulletEntity{
		ImageGetter: bulletImageGetter,
		Position: types.Position{
			X: bulletX,
			Y: bulletY,
		},
		Speed:     120.0, // Скорость пули
		Direction: tank.Direction,
		Owner:     tank,
		Altitude:  types.SURFACE, // Пули на уровне поверхности
	}

	log.Printf("DEBUG: Created bullet at position (%.2f, %.2f) direction %d",
		bullet.Position.X, bullet.Position.Y, bullet.Direction)

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
			bullet.Position.Y -= delta
		case types.DirectionDown:
			bullet.Position.Y += delta
		case types.DirectionLeft:
			bullet.Position.X -= delta
		case types.DirectionRight:
			bullet.Position.X += delta
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
