package use_cases

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// AIUseCases представляет Use Case для работы с AI логикой (Application Layer)
type AIUseCases struct {
	luaEngine     interfaces.ILuaEngine
	typeConverter interfaces.IAITypeConverter
}

// NewAIUseCases создает новый AI Use Case
func NewAIUseCases(
	luaEngine interfaces.ILuaEngine,
	typeConverter interfaces.IAITypeConverter,
) *AIUseCases {
	return &AIUseCases{
		luaEngine:     luaEngine,
		typeConverter: typeConverter,
	}
}

// ExecuteAI выполняет AI логику для танка и возвращает решение
// Это бизнес-операция, которая скрывает детали работы с Lua
func (uc *AIUseCases) ExecuteAI(
	tank *types.TankEntity,
	context *types.GameAiContext,
) (types.EnemyAIDecision, error) {
	// 1. Конвертируем входные данные в Lua типы
	tankTable, err := uc.typeConverter.TankToLua(tank)
	if err != nil {
		return types.EnemyAIDecision{}, err
	}

	contextTable, err := uc.typeConverter.ContextToLua(context)
	if err != nil {
		return types.EnemyAIDecision{}, err
	}

	// 2. Вызываем Lua функцию через engine
	results, err := uc.luaEngine.CallFunction(
		"updateEnemyAI",
		tankTable,
		contextTable,
	)
	if err != nil {
		return types.EnemyAIDecision{}, err
	}

	// 3. Конвертируем результат обратно в доменный тип
	return uc.typeConverter.LuaToDecision(results)
}

// Close освобождает ресурсы Lua VM
func (uc *AIUseCases) Close() {
	if uc.luaEngine != nil {
		uc.luaEngine.Close()
	}
}
