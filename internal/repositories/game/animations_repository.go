package game

import (
	image_providers "github.com/shpaker/tnk25/internal/types/image_providers"
)

type AnimationsRepository struct {
	animations []*image_providers.AnimationProvider
}

func NewAnimationsRepository() *AnimationsRepository {
	return &AnimationsRepository{
		animations: make([]*image_providers.AnimationProvider, 0),
	}
}

func (ar *AnimationsRepository) AddAnimation(
	animation *image_providers.AnimationProvider,
) {
	ar.animations = append(ar.animations, animation)
}

func (ar *AnimationsRepository) GetAllAnimations() []*image_providers.AnimationProvider {
	return ar.animations
}
