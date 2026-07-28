package use_cases

import (
	"github.com/shpaker/tnk9x/internal/types"
)

type MapUseCases struct {
	mapEntity *types.MapEntity
}

func NewMapUseCases(
	mapEntity *types.MapEntity,
) *MapUseCases {
	return &MapUseCases{
		mapEntity: mapEntity,
	}
}

func (uc *MapUseCases) GetBlocks() types.MapBlocks {
	if uc.mapEntity == nil {
		return types.MapBlocks{}
	}
	return uc.mapEntity.GetBlocks()
}

func (uc *MapUseCases) RemoveBlock(block *types.BlockEntity) error {
	if uc.mapEntity == nil {
		return nil
	}
	return uc.mapEntity.RemoveBlock(block)
}

func (uc *MapUseCases) GetSizePx() types.Size {
	if uc.mapEntity == nil {
		return types.Size{}
	}
	return uc.mapEntity.GetSizePx()
}

func (uc *MapUseCases) GetRandomBonusSpawnPosition() types.Position {
	if uc.mapEntity == nil {
		return types.Position{X: 0, Y: 0}
	}
	return uc.mapEntity.GetRandomBonusSpawnPosition()
}
