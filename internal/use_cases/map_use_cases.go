package use_cases

import (
	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/types"
)

// MapUseCases реализация интерфейса MapUseCases
type MapUseCases struct {
	blocksRepo game.IBlocksRepository
}

// NewMapUseCases создает новый экземпляр MapUseCases
func NewMapUseCases(blocksRepo game.IBlocksRepository) *MapUseCases {
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
	uc.blocksRepo.RemoveBlockByPointer(block)
	return nil
}
