package game

import (
	"github.com/shpaker/gonflict/internal/types"
)

type AnimationsRepository struct {
	animations []*types.TileAnimationEntity
}

func NewAnimationsRepository() *AnimationsRepository {
	return &AnimationsRepository{
		animations: make([]*types.TileAnimationEntity, 0),
	}
}

// AddAnimation добавляет анимацию в репозиторий
func (ar *AnimationsRepository) AddAnimation(
	animation *types.TileAnimationEntity,
) {
	ar.animations = append(ar.animations, animation)
}

// GetAllAnimations возвращает все анимации
func (ar *AnimationsRepository) GetAllAnimations() []*types.TileAnimationEntity {
	return ar.animations
}
