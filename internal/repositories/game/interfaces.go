package game

import "github.com/shpaker/gonflict/internal/models"

// IBulletsRepository определяет интерфейс для работы с пулями
type IBulletsRepository interface {
	// AddBullet добавляет пулю в репозиторий
	AddBullet(bullet models.Bullet)

	// GetAllBullets возвращает все пули
	GetAllBullets() []models.Bullet

	// RemoveBullet удаляет пулю по индексу
	RemoveBullet(index int) error

	// ClearAllBullets очищает все пули
	ClearAllBullets()

	// GetBulletsCount возвращает количество пуль
	GetBulletsCount() int
}

// IBlocksRepository определяет интерфейс для работы с блоками
type IBlocksRepository interface {
	// AddBlock добавляет блок в репозиторий
	AddBlock(block models.Block)

	// GetAllBlocks возвращает все блоки
	GetAllBlocks() *[]models.Block

	// RemoveBlockByPointer удаляет блок по указателю
	RemoveBlockByPointer(block *models.Block) error
}
