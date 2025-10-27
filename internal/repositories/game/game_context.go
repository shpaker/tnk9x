package game

import "github.com/shpaker/gonflict/internal/types"

// GameContext представляет контекст игры для AI
type GameContext struct {
	Player  *types.TankEntity   // Игрок
	Enemies []*types.TankEntity // Враги
	Bullets []types.BulletEntity // Пули
	Blocks  []types.BlockEntity  // Блоки/стены
}

