package game

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

var _ interfaces.IAnimationsRepository = (*AnimationsRepository)(nil)

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

func (ar *AnimationsRepository) RemoveAnimation(
	animation *image_providers.AnimationProvider,
) {
	for i, current := range ar.animations {
		if current == animation {
			ar.animations = append(
				ar.animations[:i],
				ar.animations[i+1:]...,
			)
			return
		}
	}
}
