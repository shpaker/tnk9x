package game

import "github.com/shpaker/gonflict/internal/types"

// IBulletsRepository определяет интерфейс для работы с пулями
type IBulletsRepository interface {
	// AddBullet добавляет пулю в репозиторий
	AddBullet(bullet types.BulletEntity)

	// GetAllBullets возвращает все пули
	GetAllBullets() []types.BulletEntity

	// RemoveBullet удаляет пулю по индексу
	RemoveBullet(index int) error
}

// IBlocksRepository определяет интерфейс для работы с блоками
type IBlocksRepository interface {
	// AddBlock добавляет блок в репозиторий
	AddBlock(block types.BlockEntity)

	// GetAllBlocks возвращает все блоки
	GetAllBlocks() *[]types.BlockEntity

	// RemoveBlockByPointer удаляет блок по указателю
	RemoveBlockByPointer(block *types.BlockEntity) error
}
