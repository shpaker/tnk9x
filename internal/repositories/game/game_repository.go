package game

import (
	"github.com/shpaker/gonflict/internal/types"
)

// GameRepositoriesRegistry содержит все игровые репозитории
type GameRepositoriesRegistry struct {
	blocks     IBlocksRepository
	bullets    IBulletsRepository
	animations IAnimationsRepository
	tanks      ITanksRepository
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
func (gr *GameRepositoriesRegistry) BlocksRepository() IBlocksRepository {
	return gr.blocks
}

// BulletsRepository возвращает репозиторий пуль
func (gr *GameRepositoriesRegistry) BulletsRepository() IBulletsRepository {
	return gr.bullets
}

// AnimationsRepository возвращает репозиторий анимаций
func (gr *GameRepositoriesRegistry) AnimationsRepository() IAnimationsRepository {
	return gr.animations
}

// TanksRepository возвращает репозиторий танков
func (gr *GameRepositoriesRegistry) TanksRepository() ITanksRepository {
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
