package game

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/types"
)

type BlocksRepository struct {
	blocks []types.Block
}

func NewBlocksRepository() *BlocksRepository {
	return &BlocksRepository{
		blocks: make([]types.Block, 0),
	}
}

// AddBlock добавляет блок в репозиторий
func (br *BlocksRepository) AddBlock(block types.Block) {
	br.blocks = append(br.blocks, block)
}

// GetAllBlocks возвращает все блоки
func (br *BlocksRepository) GetAllBlocks() *[]types.Block {
	return &br.blocks
}

// RemoveBlockByPointer удаляет блок по указателю
func (br *BlocksRepository) RemoveBlockByPointer(block *types.Block) error {
	if block == nil {
		return fmt.Errorf("указатель на блок не может быть nil")
	}

	for i := range br.blocks {
		if &br.blocks[i] == block {
			br.blocks = append(br.blocks[:i], br.blocks[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("блок не найден в репозитории")
}
