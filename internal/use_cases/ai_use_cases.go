package use_cases

import "github.com/shpaker/gonflict/internal/types"

// AIUseCases управляет AI врагов
type AIUseCases struct {
	ai             IEnemyAI
	aiContext      *types.GameAiContext
	updateInterval int
}

// NewAIUseCases создает новый AIUseCases
func NewAIUseCases(ai IEnemyAI, aiContext *types.GameAiContext, updateInterval int) *AIUseCases {
	return &AIUseCases{
		ai:             ai,
		aiContext:      aiContext,
		updateInterval: updateInterval,
	}
}

// GetUpdateInterval возвращает интервал обновления AI
func (uc *AIUseCases) GetUpdateInterval() int {
	return uc.updateInterval
}

// UpdateAI обновляет AI для врага
func (uc *AIUseCases) UpdateAI(enemy *types.TankEntity) types.EnemyAIDecision {
	if enemy == nil || uc.ai == nil {
		return types.EnemyAIDecision{ShouldMove: false}
	}

	// Пропускаем взрывающихся или не заспавненных врагов
	if enemy.State == types.TankStateExploding || enemy.State == types.TankStateSpawning || enemy.State == types.TankStateExploded {
		return types.EnemyAIDecision{ShouldMove: false}
	}

	// Получаем решение от AI
	return uc.ai.Update(enemy, uc.aiContext)
}

// ApplyDecision применяет решение AI к врагу
func (uc *AIUseCases) ApplyDecision(enemy *types.TankEntity, decision types.EnemyAIDecision) {
	if decision.ShouldMove && enemy.Speed == 0 {
		enemy.Direction = decision.NewDirection
		enemy.Speed = 32.0
	}
}
