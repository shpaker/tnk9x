package use_cases

import (
	"github.com/shpaker/gonflict/internal/types"
)

// TankRenderUseCases отвечает за графику и рендеринг танка
type TankRenderUseCases struct {
	AnimationGetter types.IImageIDGetter
}

// NewTankRenderUseCases создает новый экземпляр TankRenderUseCases
func NewTankRenderUseCases() *TankRenderUseCases {
	return &TankRenderUseCases{}
}

// IsSpawnAnimationFinished проверяет, завершена ли анимация спавна
func (uc *TankRenderUseCases) IsSpawnAnimationFinished() bool {
	if anim, ok := uc.AnimationGetter.(*types.TileAnimationEntity); ok {
		return anim.IsFinished()
	}
	return false
}

// IsExplosionAnimationFinished проверяет, завершена ли анимация взрыва
func (uc *TankRenderUseCases) IsExplosionAnimationFinished() bool {
	if anim, ok := uc.AnimationGetter.(*types.TileAnimationEntity); ok {
		return anim.IsFinished()
	}
	return false
}
