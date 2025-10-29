package use_cases

import (
	"errors"

	"github.com/shpaker/gonflict/internal/types"
)

// TankRenderUseCases отвечает за графику и рендеринг танка
type TankRenderUseCases struct {
	animationGetter types.IImageIDGetter
}

// NewTankRenderUseCases создает новый экземпляр TankRenderUseCases
func NewTankRenderUseCases() *TankRenderUseCases {
	return &TankRenderUseCases{}
}

// GetImageID возвращает ID изображения танка
func (uc *TankRenderUseCases) GetImageID() (string, error) {
	if uc.animationGetter == nil {
		return "", errors.New("AnimationGetter is nil")
	}
	return uc.animationGetter.GetImageID()
}

// GetAnimationGetter возвращает AnimationGetter для доступа к offset
func (uc *TankRenderUseCases) GetAnimationGetter() types.IImageIDGetter {
	return uc.animationGetter
}

// SetAnimationGetter устанавливает AnimationGetter
func (uc *TankRenderUseCases) SetAnimationGetter(
	animationGetter types.IImageIDGetter,
) {
	uc.animationGetter = animationGetter
}

// IsSpawnAnimationFinished проверяет, завершена ли анимация спавна
func (uc *TankRenderUseCases) IsSpawnAnimationFinished() bool {
	if anim, ok := uc.animationGetter.(*types.TileAnimationEntity); ok {
		return anim.IsFinished()
	}
	return false
}

// IsExplosionAnimationFinished проверяет, завершена ли анимация взрыва
func (uc *TankRenderUseCases) IsExplosionAnimationFinished() bool {
	if anim, ok := uc.animationGetter.(*types.TileAnimationEntity); ok {
		return anim.IsFinished()
	}
	return false
}
