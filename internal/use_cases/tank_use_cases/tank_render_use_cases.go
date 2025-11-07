package tank_use_cases

import (
	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

// TankRenderUseCases отвечает за графику и рендеринг танка
type TankRenderUseCases struct{}

// NewTankRenderUseCases создает новый экземпляр TankRenderUseCases
func NewTankRenderUseCases() *TankRenderUseCases {
	return &TankRenderUseCases{}
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
