package use_cases

import (
	"github.com/shpaker/gonflict/internal/types"
)

// MapUseCases реализация интерфейса MapUseCases
type MapUseCases struct {
	mapEntity *types.MapEntity
}

// NewMapUseCases создает новый экземпляр MapUseCases
func NewMapUseCases(
	mapEntity *types.MapEntity,
) *MapUseCases {
	return &MapUseCases{
		mapEntity: mapEntity,
	}
}

// GetBlocks возвращает все блоки карты
func (uc *MapUseCases) GetBlocks() types.MapBlocks {
	if uc.mapEntity == nil {
		return types.MapBlocks{}
	}
	return uc.mapEntity.GetBlocks()
}

// RemoveBlock удаляет блок по указателю
func (uc *MapUseCases) RemoveBlock(block *types.BlockEntity) error {
	if uc.mapEntity == nil {
		return nil
	}
	return uc.mapEntity.RemoveBlock(block)
}

func (uc *MapUseCases) RemoveBlocks(blocks []*types.BlockEntity) error {
	if uc.mapEntity == nil {
		return nil
	}
	for _, block := range blocks {
		_ = uc.mapEntity.RemoveBlock(block)
	}
	return nil
}

// GetSizePx возвращает размер карты в пикселях
func (uc *MapUseCases) GetSizePx() types.Size {
	if uc.mapEntity == nil {
		return types.Size{}
	}
	return uc.mapEntity.GetSizePx()
}
