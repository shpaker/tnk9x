package game

import (
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

type AnimationsRepository struct {
	animations []*image_providers.AnimationProvider
}

func NewAnimationsRepository() *AnimationsRepository {
	return &AnimationsRepository{
		animations: make([]*image_providers.AnimationProvider, 0),
	}
}

// AddAnimation добавляет анимацию в репозиторий
func (ar *AnimationsRepository) AddAnimation(
	animation *image_providers.AnimationProvider,
) {
	ar.animations = append(ar.animations, animation)
}

// GetAllAnimations возвращает все анимации
func (ar *AnimationsRepository) GetAllAnimations() []*image_providers.AnimationProvider {
	return ar.animations
}
