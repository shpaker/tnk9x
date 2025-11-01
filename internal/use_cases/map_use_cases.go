package use_cases

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// MapUseCases реализация интерфейса MapUseCases
type MapUseCases struct {
	blocksRepo interfaces.IBlocksRepository
}

// NewMapUseCases создает новый экземпляр MapUseCases
func NewMapUseCases(blocksRepo interfaces.IBlocksRepository) *MapUseCases {
	return &MapUseCases{
		blocksRepo: blocksRepo,
	}
}

// GetBlocks возвращает все блоки карты
func (uc *MapUseCases) GetBlocks() []types.BlockEntity {
	blocks := uc.blocksRepo.GetAllBlocks()
	if blocks == nil {
		return []types.BlockEntity{}
	}
	return *blocks
}

// RemoveBlock удаляет блок по указателю
func (uc *MapUseCases) RemoveBlock(block *types.BlockEntity) error {
	if block == nil {
		return nil
	}
	return uc.blocksRepo.RemoveBlockByPointer(block)
}
