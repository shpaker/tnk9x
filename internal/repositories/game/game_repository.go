package game

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// GameRepositoriesRegistry содержит все игровые репозитории
type GameRepositoriesRegistry struct {
	blocks     interfaces.IBlocksRepository
	bullets    interfaces.IBulletsRepository
	animations interfaces.IAnimationsRepository
	tanks      interfaces.ITanksRepository
}

// NewGameRepositoriesRegistry создает новый GameRepositoriesRegistry со всеми репозиториями
func NewGameRepositoriesRegistry() *GameRepositoriesRegistry {
	return &GameRepositoriesRegistry{
		blocks:     NewBlocksRepository(),
		bullets:    NewBulletsRepository(),
		animations: NewAnimationsRepository(),
		tanks:      NewTanksRepository(),
	}
}

// === Методы для доступа к репозиториям ===

// BlocksRepository возвращает репозиторий блоков
func (gr *GameRepositoriesRegistry) BlocksRepository() interfaces.IBlocksRepository {
	return gr.blocks
}

// BulletsRepository возвращает репозиторий пуль
func (gr *GameRepositoriesRegistry) BulletsRepository() interfaces.IBulletsRepository {
	return gr.bullets
}

// AnimationsRepository возвращает репозиторий анимаций
func (gr *GameRepositoriesRegistry) AnimationsRepository() interfaces.IAnimationsRepository {
	return gr.animations
}

// TanksRepository возвращает репозиторий танков
func (gr *GameRepositoriesRegistry) TanksRepository() interfaces.ITanksRepository {
	return gr.tanks
}

// GetGameContext возвращает контекст игры для AI
func (gr *GameRepositoriesRegistry) GetGameContext() *types.GameAiContext {
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

	return &types.GameAiContext{
		Player:  player,
		Enemies: enemies,
		Bullets: bullets,
		Blocks:  blocks,
	}
}
