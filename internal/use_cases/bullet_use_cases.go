package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IBulletUseCases = (*BulletUseCases)(nil)

type BulletUseCases struct {
	bulletsRepository interfaces.IBulletsRepository
	tilesUseCases     interfaces.ITilesUseCases
	tankSpriteSize    uint
}

func NewBulletUseCases(
	bulletsRepository interfaces.IBulletsRepository,
	tilesUseCases interfaces.ITilesUseCases,
	tankSpriteSize uint,
) *BulletUseCases {
	return &BulletUseCases{
		bulletsRepository: bulletsRepository,
		tilesUseCases:     tilesUseCases,
		tankSpriteSize:    tankSpriteSize,
	}
}

func (uc *BulletUseCases) ShootBullet(tank *types.TankEntity) error {
	if !tank.IsActive() {
		return nil
	}

	bulletImageGetter, err := uc.tilesUseCases.CreateStaticTile("bullet")
	if err != nil {
		return err
	}

	bulletX := tank.Position.X + float64(uc.tankSpriteSize)/2 - 2
	bulletY := tank.Position.Y + float64(uc.tankSpriteSize)/2 - 2

	switch tank.Direction {
	case types.DirectionUp:
		bulletY = tank.Position.Y - 4
	case types.DirectionDown:
		bulletY = tank.Position.Y + float64(uc.tankSpriteSize)/2
	case types.DirectionLeft:
		bulletX = tank.Position.X - 4
	case types.DirectionRight:
		bulletX = tank.Position.X + float64(uc.tankSpriteSize)/2
	}

	// Получаем спецификации танка для пули
	specs := tank.GetSpecs()

	bullet := types.NewBulletEntity(
		types.Position{
			X: bulletX,
			Y: bulletY,
		},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		bulletImageGetter,
		tank.Direction,
		specs,
		tank,
	)

	_ = uc.bulletsRepository.AddBullet(bullet)
	return nil
}

func (uc *BulletUseCases) UpdateBullets(dt float64) error {
	bullets := uc.bulletsRepository.GetAllBullets()
	for i := len(bullets) - 1; i >= 0; i-- {
		bullet := bullets[i]
		if bullet == nil {
			continue
		}

		delta := bullet.GetSpeed() * dt
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

func (uc *BulletUseCases) GetBullets() []*types.BulletEntity {
	return uc.bulletsRepository.GetAllBullets()
}

func (uc *BulletUseCases) RemoveBullet(bullet *types.BulletEntity) error {
	return uc.bulletsRepository.RemoveBullet(bullet)
}
