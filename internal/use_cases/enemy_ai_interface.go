package use_cases

import "github.com/shpaker/gonflict/internal/types"

// IEnemyAI интерфейс для AI вражеских танков
type IEnemyAI interface {
	Update(enemy *types.TankEntity, context *types.GameAiContext) types.EnemyAIDecision
}
