package use_cases

import (
	"github.com/shpaker/gonflict/internal/types"
)

// HQRenderUseCases отвечает за графику и рендеринг базы
type HQRenderUseCases struct {
	AnimationGetter types.IImageIDGetter
}

// NewHQRenderUseCases создает новый экземпляр HQRenderUseCases
func NewHQRenderUseCases() *HQRenderUseCases {
	return &HQRenderUseCases{
		AnimationGetter: nil,
	}
}

// IsExplosionAnimationFinished возвращает true если анимация взрыва завершена
func (r *HQRenderUseCases) IsExplosionAnimationFinished() bool {
	if r.AnimationGetter == nil {
		return true
	}

	// Проверяем, что анимация завершена
	if tileAnim, ok := r.AnimationGetter.(*types.TileAnimationEntity); ok {
		return tileAnim.IsFinished()
	}

	return true
}
