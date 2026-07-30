package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IMapUseCases = (*MapUseCases)(nil)

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

// IsIceAt проверяет, лежит ли точка на блоке льда;
// CheckColliders не подходит — лёд на GROUND, танки на SURFACE
func (uc *MapUseCases) IsIceAt(position types.Position) bool {
	for _, block := range uc.GetBlocks() {
		if block == nil || block.Data == nil || block.Data.Name != types.Ice {
			continue
		}

		blockPos := block.GetPosition()
		blockSize := block.GetSize()

		if position.X >= blockPos.X &&
			position.X < blockPos.X+float64(blockSize.Width) &&
			position.Y >= blockPos.Y &&
			position.Y < blockPos.Y+float64(blockSize.Height) {
			return true
		}
	}
	return false
}
