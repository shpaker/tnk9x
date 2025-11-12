package tank_use_cases

import (
	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// TankRenderUseCases отвечает за графику и рендеринг танка
type TankRenderUseCases struct {
	tilesUseCases *use_cases.TilesUseCases
}

// NewTankRenderUseCases создает новый экземпляр TankRenderUseCases
func NewTankRenderUseCases(
	tilesUseCases *use_cases.TilesUseCases,
) *TankRenderUseCases {
	return &TankRenderUseCases{
		tilesUseCases: tilesUseCases,
	}
}

// IsSpawnAnimationFinished проверяет, завершена ли анимация спавна
func (uc *TankRenderUseCases) IsSpawnAnimationFinished(
	tank *types.TankEntity,
) bool {
	if tank.Image == nil {
		return false
	}
	if anim, ok := tank.Image.(*image_providers.AnimationProvider); ok {
		return anim.IsFinished()
	}
	return false
}

// IsExplosionAnimationFinished проверяет, завершена ли анимация взрыва
func (uc *TankRenderUseCases) IsExplosionAnimationFinished(
	tank *types.TankEntity,
) bool {
	if tank.Image == nil {
		return false
	}
	if anim, ok := tank.Image.(*image_providers.AnimationProvider); ok {
		return anim.IsFinished()
	}
	return false
}

// UpdateTankAnimation пересоздает анимацию танка в соответствии с текущим направлением.
func (uc *TankRenderUseCases) UpdateTankAnimation(
	tank *types.TankEntity,
) {
	if uc.tilesUseCases == nil || tank == nil {
		return
	}

	animationName := tank.GetTankAnimationName()
	tankAnimation, err := uc.tilesUseCases.CreateTankAnimationTile(
		animationName,
		tank.IsEnemy(),
	)
	if err != nil {
		return
	}

	if tank.Image != nil {
		if anim, ok := tank.Image.(*image_providers.AnimationProvider); ok {
			uc.tilesUseCases.StopAnimation(anim)
		}
	}

	tank.Image = tankAnimation
	uc.tilesUseCases.AddAnimation(tankAnimation)

	uc.SyncAnimationWithState(tank)
}

// SyncAnimationWithState синхронизирует анимацию гусениц с состоянием танка
func (uc *TankRenderUseCases) SyncAnimationWithState(
	tank *types.TankEntity,
) {
	if uc.tilesUseCases == nil {
		return
	}

	if tank.Image == nil {
		return
	}
	anim, ok := tank.Image.(*image_providers.AnimationProvider)
	if !ok {
		return
	}

	// Если танк стоит - анимация должна быть остановлена
	if tank.State == types.TankStateStopped {
		if anim.IsAnimating {
			uc.tilesUseCases.StopAnimation(anim)
		}
		return
	}

	// Определяем, должна ли анимация быть запущена
	shouldAnimate := tank.State == types.TankStateMoving ||
		tank.State == types.TankStateBraking

	// Синхронизируем: если состояние требует анимации, но она остановлена - запускаем
	// Если состояние не требует анимации, но она запущена - останавливаем
	if shouldAnimate && !anim.IsAnimating {
		uc.tilesUseCases.StartAnimation(anim)
	} else if !shouldAnimate && anim.IsAnimating {
		uc.tilesUseCases.StopAnimation(anim)
	}
}
