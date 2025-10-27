package game

import "github.com/shpaker/gonflict/internal/types"

// IGameRepositoryFacade определяет интерфейс для главного репозитория игры
type IGameRepositoryFacade interface {
	BlocksRepository() IBlocksRepository
	BulletsRepository() IBulletsRepository
	AnimationsRepository() IAnimationsRepository
	TanksRepository() ITanksRepository

	// Вспомогательные методы для работы с танками
	AddPlayerTank(tank *types.TankEntity)
	GetPlayerTank() *types.TankEntity
	GetAllEnemies() []*types.TankEntity
	AddEnemy(enemy *types.TankEntity)

	// Метод для получения контекста игры для AI
	GetGameContext() *types.GameAiContext
}

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
	// === Методы для работы с игроком ===
	SetPlayer(player *types.TankEntity)
	GetPlayer() *types.TankEntity
	HasPlayer() bool
	ClearPlayer()

	// === Методы для работы с врагами ===
	AddEnemy(enemy *types.TankEntity)
	GetAllEnemies() []*types.TankEntity
	RemoveEnemy(index int) error

	// === Методы для обратной совместимости ===
	AddTank(tank *types.TankEntity)
	GetAllTanks() []*types.TankEntity
	RemoveTank(tank *types.TankEntity) error
}
