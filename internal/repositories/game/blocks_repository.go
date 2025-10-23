package game

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/models"
)

type BlocksRepository struct {
	blocks []models.Block
}

func NewBlocksRepository() *BlocksRepository {
	return &BlocksRepository{
		blocks: make([]models.Block, 0),
	}
}

// AddBlock добавляет блок в репозиторий
func (br *BlocksRepository) AddBlock(block models.Block) {
	br.blocks = append(br.blocks, block)
}

// GetAllBlocks возвращает все блоки
func (br *BlocksRepository) GetAllBlocks() *[]models.Block {
	return &br.blocks
}

// RemoveBlockByPointer удаляет блок по указателю
func (br *BlocksRepository) RemoveBlockByPointer(block *models.Block) error {
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
