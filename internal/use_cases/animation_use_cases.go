package use_cases

import (
	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/types"
)

// AnimationUseCases управляет анимацией объектов
type AnimationUseCases struct {
	animationsRepo game.IAnimationsRepository
}

// NewAnimationUseCases создает новый экземпляр AnimationUseCases
func NewAnimationUseCases(animationsRepo game.IAnimationsRepository) *AnimationUseCases {
	return &AnimationUseCases{
		animationsRepo: animationsRepo,
	}
}

// AddAnimation добавляет анимацию в репозиторий
func (uc *AnimationUseCases) AddAnimation(animation *types.TileAnimationEntity) {
	uc.animationsRepo.AddAnimation(animation)
}

// UpdateAnimations обновляет все анимации в репозитории
func (uc *AnimationUseCases) UpdateAnimations() {
	animations := uc.animationsRepo.GetAllAnimations()
	for _, animation := range animations {
		if animation != nil {
			animation.UpdateAnimation()
		}
	}
}

// StartAnimation запускает анимацию объекта
func (uc *AnimationUseCases) StartAnimation(animation *types.TileAnimationEntity) {
	if animation == nil {
		return
	}
	animation.StartAnimation()
}

// StopAnimation останавливает анимацию объекта
func (uc *AnimationUseCases) StopAnimation(animation *types.TileAnimationEntity) {
	if animation == nil {
		return
	}
	animation.StopAnimation()
}
