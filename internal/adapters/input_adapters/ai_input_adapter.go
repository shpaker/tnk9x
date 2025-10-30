package input_adapters

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// AiInputAdapter адаптер для работы с AI через Lua скрипты
// Теперь использует AIUseCases вместо прямой работы с Lua
type AiInputAdapter struct {
	tankActions    interfaces.ITankActionsUseCases
	tank           *types.TankEntity
	aiContext      *types.GameAiContext
	updateInterval int
	tickCounter    int
	aiUseCases     *use_cases.AIUseCases
}

// NewAiInputAdapter создает новый AI адаптер
func NewAiInputAdapter(
	tankActions interfaces.ITankActionsUseCases,
	tank *types.TankEntity,
	aiContext *types.GameAiContext,
	updateInterval int,
	aiUseCases *use_cases.AIUseCases,
) (*AiInputAdapter, error) {
	return &AiInputAdapter{
		tankActions:    tankActions,
		tank:           tank,
		aiContext:      aiContext,
		updateInterval: updateInterval,
		tickCounter:    0,
		aiUseCases:     aiUseCases,
	}, nil
}

// Update обновляет AI логику для танка
func (a *AiInputAdapter) Update(dt float64) {
	// Пропускаем неактивных врагов
	if a.tank == nil || !a.tank.IsActive() {
		return
	}

	// Проверяем, нужно ли обновлять AI
	if a.tickCounter == 0 {
		// Используем AIUseCases для выполнения AI логики
		if a.aiUseCases != nil && a.aiContext != nil {
			decision, err := a.aiUseCases.ExecuteAI(a.tank, a.aiContext)
			if err == nil && decision.Direction != 0 {
				// Применяем решение
				a.applyDecision(decision)
			}
		}
	}
	// Увеличиваем счетчик тиков
	a.tickCounter++

	if a.tickCounter >= a.updateInterval {
		a.tickCounter = 0
	}
}

// applyDecision применяет решение AI к танку через TankActionsUseCases
func (a *AiInputAdapter) applyDecision(
	decision types.EnemyAIDecision,
) {
	if a.tank == nil {
		return
	}
	a.tankActions.ApplyDecision(a.tank, decision)
}

// Close освобождает ресурсы
func (a *AiInputAdapter) Close() {
	if a.aiUseCases != nil {
		a.aiUseCases.Close()
	}
}
