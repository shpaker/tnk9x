package game

import (
	"github.com/shpaker/gonflict/internal/types"
)

// GameRepositoryFacade - фасад, который содержит все игровые репозитории
type GameRepositoryFacade struct {
	blocks     *BlocksRepository
	bullets    *BulletsRepository
	animations *AnimationsRepository
	tanks      *TanksRepository
}

// NewGameRepositoryFacade создает новый GameRepositoryFacade со всеми репозиториями
func NewGameRepositoryFacade() *GameRepositoryFacade {
	return &GameRepositoryFacade{
		blocks:     NewBlocksRepository(),
		bullets:    NewBulletsRepository(),
		animations: NewAnimationsRepository(),
		tanks:      NewTanksRepository(),
	}
}

// === Методы для доступа к репозиториям ===

// BlocksRepository возвращает репозиторий блоков
func (gr *GameRepositoryFacade) BlocksRepository() IBlocksRepository {
	return gr.blocks
}

// BulletsRepository возвращает репозиторий пуль
func (gr *GameRepositoryFacade) BulletsRepository() IBulletsRepository {
	return gr.bullets
}

// AnimationsRepository возвращает репозиторий анимаций
func (gr *GameRepositoryFacade) AnimationsRepository() IAnimationsRepository {
	return gr.animations
}

// TanksRepository возвращает репозиторий танков
func (gr *GameRepositoryFacade) TanksRepository() ITanksRepository {
	return gr.tanks
}

// === Вспомогательные методы для работы с танками ===

// AddPlayerTank добавляет танк игрока
func (gr *GameRepositoryFacade) AddPlayerTank(tank *types.TankEntity) {
	gr.tanks.SetPlayer(tank)
}

// GetPlayerTank возвращает танк игрока
func (gr *GameRepositoryFacade) GetPlayerTank() *types.TankEntity {
	return gr.tanks.GetPlayer()
}

// GetAllEnemies возвращает всех врагов
func (gr *GameRepositoryFacade) GetAllEnemies() []*types.TankEntity {
	return gr.tanks.GetAllEnemies()
}

// AddEnemy добавляет танк врага
func (gr *GameRepositoryFacade) AddEnemy(enemy *types.TankEntity) {
	gr.tanks.AddEnemy(enemy)
}

// GetGameContext возвращает контекст игры для AI
func (gr *GameRepositoryFacade) GetGameContext() *GameContext {
	// Получаем игрока
	var player *types.TankEntity
	if gr.tanks.HasPlayer() {
		player = gr.tanks.GetPlayer()
	}

	// Получаем всех врагов
	enemies := gr.tanks.GetAllEnemies()

	// Получаем все пули
	bullets := gr.bullets.GetAllBullets()

	// Получаем все блоки
	blocksPtr := gr.blocks.GetAllBlocks()
	var blocks []types.BlockEntity
	if blocksPtr != nil {
		blocks = *blocksPtr
	}

	return &GameContext{
		Player:  player,
		Enemies: enemies,
		Bullets: bullets,
		Blocks:  blocks,
	}
}
