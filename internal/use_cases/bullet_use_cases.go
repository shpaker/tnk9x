package use_cases

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/repositories"
	"github.com/shpaker/gonflict/internal/types"
)

// BulletUseCases реализация интерфейса BulletUseCases
type BulletUseCases struct {
	bulletsRepo repositories.IBulletsRepository
}

// NewBulletUseCases создает новый экземпляр BulletUseCases
func NewBulletUseCases(bulletsRepo repositories.IBulletsRepository) *BulletUseCases {
	return &BulletUseCases{
		bulletsRepo: bulletsRepo,
	}
}

// ShootBullet создает новую пулю от указанного танка
func (uc *BulletUseCases) ShootBullet(tank *types.TankEntity) error {
	// Создаем простую пулю (квадрат 4x4 пикселя)
	bulletImage := ebiten.NewImage(4, 4)
	bulletImage.Fill(color.RGBA{255, 255, 0, 255}) // Желтый цвет

	// Вычисляем позицию пули в зависимости от направления танка
	bulletX := tank.WorldPosition.X + TankSpriteSize/2 - 2
	bulletY := tank.WorldPosition.Y + TankSpriteSize/2 - 2

	// Корректируем позицию в зависимости от направления
	switch tank.Direction {
	case types.DirectionUp:
		bulletY = tank.WorldPosition.Y - 4
	case types.DirectionDown:
		bulletY = tank.WorldPosition.Y + TankSpriteSize
	case types.DirectionLeft:
		bulletX = tank.WorldPosition.X - 4
	case types.DirectionRight:
		bulletX = tank.WorldPosition.X + TankSpriteSize
	}

	bullet := types.BulletEntity{
		Image: bulletImage,
		WorldPosition: types.Position{
			X: bulletX,
			Y: bulletY,
		},
		Speed:     200.0, // Скорость пули
		Direction: tank.Direction,
		Owner:     tank,
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
