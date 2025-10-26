package game

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/types"
)

type BlocksRepository struct {
	blocks []types.BlockEntity
}

func NewBlocksRepository() *BlocksRepository {
	return &BlocksRepository{
		blocks: make([]types.BlockEntity, 0),
	}
}

// AddBlock добавляет блок в репозиторий
func (br *BlocksRepository) AddBlock(block types.BlockEntity) {
	br.blocks = append(br.blocks, block)
}

// GetAllBlocks возвращает все блоки
func (br *BlocksRepository) GetAllBlocks() *[]types.BlockEntity {
	return &br.blocks
}

// RemoveBlockByPointer удаляет блок по указателю
func (br *BlocksRepository) RemoveBlockByPointer(block *types.BlockEntity) error {
	if block == nil {
		return fmt.Errorf("block pointer cannot be nil")
	}

	for i := range br.blocks {
		if &br.blocks[i] == block {
			br.blocks = append(br.blocks[:i], br.blocks[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("block not found in repository")
}
