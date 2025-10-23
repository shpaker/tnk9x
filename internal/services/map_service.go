package services

import (
	"github.com/shpaker/gonflict/internal/models"
	"github.com/shpaker/gonflict/internal/repositories/game"
)

type MapService struct {
	blocksRepo game.IBlocksRepository
}

func NewMapService(
	blocksRepo game.IBlocksRepository,
) MapService {
	return MapService{
		blocksRepo: blocksRepo,
	}
}

// GetBlocks возвращает список блоков-стен для коллизий
func (s *MapService) GetBlocks() []models.Block {
	return *s.blocksRepo.GetAllBlocks()
}

// RemoveBlock удаляет блок из репозитория
func (s *MapService) RemoveBlock(block *models.Block) {
	if block == nil {
		return
	}
	s.blocksRepo.RemoveBlockByPointer(block)
}
