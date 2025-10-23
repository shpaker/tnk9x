package use_cases

import (
	"github.com/shpaker/gonflict/internal/repositories"
	"github.com/shpaker/gonflict/internal/types"
)

// MapUseCases реализация интерфейса MapUseCases
type MapUseCases struct {
	blocksRepo repositories.IBlocksRepository
}

// NewMapUseCases создает новый экземпляр MapUseCases
func NewMapUseCases(blocksRepo repositories.IBlocksRepository) *MapUseCases {
	return &MapUseCases{
		blocksRepo: blocksRepo,
	}
}

// GetBlocks возвращает все блоки карты
func (uc *MapUseCases) GetBlocks() []types.Block {
	blocks := uc.blocksRepo.GetAllBlocks()
	if blocks == nil {
		return []types.Block{}
	}
	return *blocks
}

// RemoveBlock удаляет блок по указателю
func (uc *MapUseCases) RemoveBlock(block *types.Block) error {
	if block == nil {
		return nil
	}
	uc.blocksRepo.RemoveBlockByPointer(block)
	return nil
}
