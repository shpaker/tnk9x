package use_cases

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// MapUseCases реализация интерфейса MapUseCases
type MapUseCases struct {
	blocksRepository interfaces.IBlocksRepository
}

// NewMapUseCases создает новый экземпляр MapUseCases
func NewMapUseCases(
	blocksRepository interfaces.IBlocksRepository,
) *MapUseCases {
	return &MapUseCases{
		blocksRepository: blocksRepository,
	}
}

// GetBlocks возвращает все блоки карты
func (uc *MapUseCases) GetBlocks() []types.BlockEntity {
	blocks := uc.blocksRepository.GetAllBlocks()
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
	return uc.blocksRepository.RemoveBlockByPointer(block)
}
