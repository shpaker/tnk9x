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

// IAnimationsRepository определяет интерфейс для работы с анимациями
type IAnimationsRepository interface {
	// AddAnimation добавляет анимацию в репозиторий
	AddAnimation(animation *types.TileAnimationEntity)

	// GetAllAnimations возвращает все анимации
	GetAllAnimations() []*types.TileAnimationEntity
}

// ITanksRepository определяет интерфейс для работы с танками
type ITanksRepository interface {
	// AddTank добавляет танк в репозиторий
	AddTank(tank *types.TankEntity)

	// GetAllTanks возвращает все танки
	GetAllTanks() []*types.TankEntity

	// RemoveTank удаляет танк по индексу
	RemoveTank(index int) error

	// GetTank возвращает танк по индексу
	GetTank(index int) (*types.TankEntity, error)
}
