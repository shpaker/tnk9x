package game

import (
	"github.com/shpaker/gonflict/internal/types"
)

// GameRepository - фасад, который содержит все игровые репозитории
type GameRepository struct {
	blocks     *BlocksRepository
	bullets    *BulletsRepository
	animations *AnimationsRepository
	tanks      *TanksRepository
}

// NewGameRepository создает новый GameRepository со всеми репозиториями
func NewGameRepository() *GameRepository {
	return &GameRepository{
		blocks:     NewBlocksRepository(),
		bullets:    NewBulletsRepository(),
		animations: NewAnimationsRepository(),
		tanks:      NewTanksRepository(),
	}
}

// === Методы для доступа к репозиториям ===

// BlocksRepository возвращает репозиторий блоков
func (gr *GameRepository) BlocksRepository() IBlocksRepository {
	return gr.blocks
}

// BulletsRepository возвращает репозиторий пуль
func (gr *GameRepository) BulletsRepository() IBulletsRepository {
	return gr.bullets
}

// AnimationsRepository возвращает репозиторий анимаций
func (gr *GameRepository) AnimationsRepository() IAnimationsRepository {
	return gr.animations
}

// TanksRepository возвращает репозиторий танков
func (gr *GameRepository) TanksRepository() ITanksRepository {
	return gr.tanks
}

// === Вспомогательные методы для работы с танками ===

// AddPlayerTank добавляет танк игрока
func (gr *GameRepository) AddPlayerTank(tank *types.TankEntity) {
	gr.tanks.SetPlayer(tank)
}

// GetPlayerTank возвращает танк игрока
func (gr *GameRepository) GetPlayerTank() *types.TankEntity {
	return gr.tanks.GetPlayer()
}

// GetAllEnemies возвращает всех врагов
func (gr *GameRepository) GetAllEnemies() []*types.TankEntity {
	return gr.tanks.GetAllEnemies()
}

// AddEnemy добавляет танк врага
func (gr *GameRepository) AddEnemy(enemy *types.TankEntity) {
	gr.tanks.AddEnemy(enemy)
}

// GetGameContext возвращает контекст игры для AI
func (gr *GameRepository) GetGameContext() *GameContext {
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
