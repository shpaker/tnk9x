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

// RestoreBlock возвращает блок в исходное целое состояние: отколотый
// кирпич — в полный размер, снесённый — обратно на карту. Блок не
// восстанавливается, пока его клетку занимает кто-то из obstacles
// (танк или пуля): возвращает false
func (uc *MapUseCases) RestoreBlock(
	block *types.BlockEntity,
	obstacles []types.IEntityCollider,
) bool {
	origin, size := block.Data.Position, block.Data.Size
	for _, obstacle := range obstacles {
		if types.RectsOverlap(
			origin,
			size,
			obstacle.GetPosition(),
			obstacle.GetSize(),
		) {
			return false
		}
	}

	block.Position = origin
	block.Size = size
	for _, existing := range uc.mapEntity.GetBlocks() {
		if existing == block {
			return true
		}
	}
	uc.mapEntity.AddBlock(block)
	return true
}

// IsBlockIntact — блок на карте в исходном размере
func (uc *MapUseCases) IsBlockIntact(block *types.BlockEntity) bool {
	if block.Position != block.Data.Position || block.Size != block.Data.Size {
		return false
	}
	for _, existing := range uc.mapEntity.GetBlocks() {
		if existing == block {
			return true
		}
	}
	return false
}

func (uc *MapUseCases) GetSizePx() types.Size {
	if uc.mapEntity == nil {
		return types.Size{}
	}
	return uc.mapEntity.GetSizePx()
}

// GetBorders реализует IMapUseCases
func (uc *MapUseCases) GetBorders() []*types.BorderEntity {
	if uc.mapEntity == nil {
		return nil
	}
	return uc.mapEntity.GetBorders()
}

func (uc *MapUseCases) GetRandomBonusSpawnPosition() types.Position {
	if uc.mapEntity == nil {
		return types.Position{X: 0, Y: 0}
	}
	return uc.mapEntity.GetRandomBonusSpawnPosition()
}

// IsIceAt проверяет, лежит ли точка на блоке льда;
// CheckColliders не подходит — лёд на GROUND, танки на SURFACE
// IsWaterUnder реализует IMapUseCases
func (uc *MapUseCases) IsWaterUnder(
	position types.Position,
	size types.Size,
) bool {
	for _, block := range uc.GetBlocks() {
		if block == nil || block.Data == nil || block.Data.Name != types.Water {
			continue
		}
		if types.RectsOverlap(
			position,
			size,
			block.GetPosition(),
			block.GetSize(),
		) {
			return true
		}
	}
	return false
}

func (uc *MapUseCases) IsIceAt(position types.Position) bool {
	for _, block := range uc.GetBlocks() {
		if block == nil || block.Data == nil || block.Data.Name != types.Ice {
			continue
		}

		if types.RectContains(block.GetPosition(), block.GetSize(), position) {
			return true
		}
	}
	return false
}
