package adapters

import (
	"github.com/shpaker/gonflict/internal/types"
)

// EnemyAILua Lua реализация AI
type EnemyAILua struct {
	luaAdapter *LuaAdapter
}

// NewEnemyAILua создает новый AI на основе Lua скрипта
func NewEnemyAILua(scriptPath string) (*EnemyAILua, error) {
	luaAdapter, err := NewLuaAdapter(scriptPath)
	if err != nil {
		return nil, err
	}

	return &EnemyAILua{luaAdapter: luaAdapter}, nil
}

// Update обновляет AI для врага
func (ai *EnemyAILua) Update(enemy *types.TankEntity, context *types.GameAIContext) types.EnemyAIDecision {
	// Вызываем Lua функцию через адаптер
	shouldMove, directionInt := ai.luaAdapter.CallEnemyAI(enemy, context)

	return types.EnemyAIDecision{
		ShouldMove:   shouldMove,
		NewDirection: intToDirection(directionInt),
	}
}

// Close закрывает Lua состояние
func (ai *EnemyAILua) Close() {
	if ai.luaAdapter != nil {
		ai.luaAdapter.Close()
	}
}
