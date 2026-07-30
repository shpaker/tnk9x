package game

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IEffectsRepository = (*EffectsRepository)(nil)

// EffectsRepository хранит короткоживущие визуальные эффекты уровня
type EffectsRepository struct {
	effects []*types.EffectEntity
}

func NewEffectsRepository() *EffectsRepository {
	return &EffectsRepository{
		effects: make([]*types.EffectEntity, 0),
	}
}

func (er *EffectsRepository) AddEffect(effect *types.EffectEntity) {
	if effect == nil {
		return
	}
	er.effects = append(er.effects, effect)
}

func (er *EffectsRepository) GetAllEffects() []*types.EffectEntity {
	return er.effects
}

func (er *EffectsRepository) RemoveEffect(effect *types.EffectEntity) {
	for i, current := range er.effects {
		if current == effect {
			er.effects = append(er.effects[:i], er.effects[i+1:]...)
			return
		}
	}
}
