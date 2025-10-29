package use_cases

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// BulletUseCases реализация интерфейса BulletUseCases
type BulletUseCases struct {
	bulletsRepo   interfaces.IBulletsRepository
	tilesUseCases interfaces.ITilesUseCases
}

// NewBulletUseCases создает новый экземпляр BulletUseCases
func NewBulletUseCases(
	bulletsRepo interfaces.IBulletsRepository,
	tilesUseCases interfaces.ITilesUseCases,
) *BulletUseCases {
	return &BulletUseCases{
		bulletsRepo:   bulletsRepo,
		tilesUseCases: tilesUseCases,
	}
}

// ShootBullet создает новую пулю от указанного танка
func (uc *BulletUseCases) ShootBullet(tank *types.TankEntity) error {
	// Проверяем, активен ли танк
	if !tank.IsActive() {
		return nil // Просто игнорируем выстрел
	}

	// Создаем тайл для пули с ID "bullet"
	bulletImageGetter, err := uc.tilesUseCases.CreateStaticTile("bullet")
	if err != nil {
		return err
	}

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

	uc.bulletsRepo.AddBullet(bullet)
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
