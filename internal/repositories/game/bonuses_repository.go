package game

import (
	"fmt"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IBonusesRepository = (*BonusesRepository)(nil)

type BonusesRepository struct {
	bonuses []*types.BonusEntity
}

func NewBonusesRepository() *BonusesRepository {
	return &BonusesRepository{
		bonuses: make([]*types.BonusEntity, 0),
	}
}

func (br *BonusesRepository) AddBonus(bonus *types.BonusEntity) {
	if bonus == nil {
		return
	}
	br.bonuses = append(br.bonuses, bonus)
}

func (br *BonusesRepository) GetAllBonuses() []*types.BonusEntity {
	return br.bonuses
}

func (br *BonusesRepository) RemoveBonus(bonus *types.BonusEntity) error {
	if bonus == nil {
		return nil
	}

	for i, b := range br.bonuses {
		if b == bonus {
			br.bonuses = append(br.bonuses[:i], br.bonuses[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("bonus not found")
}

func (br *BonusesRepository) RemoveBonusesWithoutOwner() {
	filtered := make([]*types.BonusEntity, 0, len(br.bonuses))
	for _, bonus := range br.bonuses {
		if bonus != nil && bonus.GetOwner() == nil {
			continue
		}
		filtered = append(filtered, bonus)
	}
	br.bonuses = filtered
}
