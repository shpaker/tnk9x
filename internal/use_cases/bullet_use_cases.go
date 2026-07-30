package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

var _ interfaces.IBulletUseCases = (*BulletUseCases)(nil)

type BulletUseCases struct {
	bulletsRepository interfaces.IBulletsRepository
	effectsRepository interfaces.IEffectsRepository
	tilesUseCases     interfaces.ITilesUseCases
	tankSpriteSize    uint
}

func NewBulletUseCases(
	bulletsRepository interfaces.IBulletsRepository,
	effectsRepository interfaces.IEffectsRepository,
	tilesUseCases interfaces.ITilesUseCases,
	tankSpriteSize uint,
) *BulletUseCases {
	return &BulletUseCases{
		bulletsRepository: bulletsRepository,
		effectsRepository: effectsRepository,
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
		bulletY = tank.Position.Y + float64(uc.tankSpriteSize)
	case types.DirectionLeft:
		bulletX = tank.Position.X - 4
	case types.DirectionRight:
		bulletX = tank.Position.X + float64(uc.tankSpriteSize)
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

// SpawnImpact создаёт короткую анимацию взрыва в точке пули;
// вызывается при попадании в стену, штаб, танк или границу поля
func (uc *BulletUseCases) SpawnImpact(bullet *types.BulletEntity) {
	if uc.effectsRepository == nil || bullet == nil {
		return
	}

	animation, err := uc.tilesUseCases.CreateBulletExplosionAnimation()
	if err != nil {
		return
	}
	uc.tilesUseCases.StartAnimation(animation)

	uc.effectsRepository.AddEffect(&types.EffectEntity{
		Position: bullet.Position,
		Size:     bullet.Size,
		Image:    animation,
	})
}

func (uc *BulletUseCases) GetImpacts() []*types.EffectEntity {
	if uc.effectsRepository == nil {
		return nil
	}
	return uc.effectsRepository.GetAllEffects()
}

// pruneFinishedImpacts убирает доигравшие анимации взрывов
func (uc *BulletUseCases) pruneFinishedImpacts() {
	if uc.effectsRepository == nil {
		return
	}
	effects := uc.effectsRepository.GetAllEffects()
	for i := len(effects) - 1; i >= 0; i-- {
		effect := effects[i]
		if effect == nil {
			continue
		}
		animation, ok := effect.Image.(*image_providers.AnimationProvider)
		if ok && animation.IsFinished() {
			// Провайдер тоже удаляется, иначе репозиторий анимаций
			// растёт с каждым выстрелом
			uc.tilesUseCases.RemoveAnimation(animation)
			uc.effectsRepository.RemoveEffect(effect)
		}
	}
}

func (uc *BulletUseCases) UpdateBullets(dt float64) error {
	uc.pruneFinishedImpacts()

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
