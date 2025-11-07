package use_cases

import (
	"errors"

	lua "github.com/yuin/gopher-lua"

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
) (types.EnemyAIDecision, error) {
	if tank == nil {
		return types.EnemyAIDecision{}, errors.New("tank is nil")
	}

	// Подготавливаем параметры танка как отдельные значения
	x := lua.LNumber(tank.Position.X)
	y := lua.LNumber(tank.Position.Y)
	direction := lua.LNumber(int(tank.Direction))
	state := lua.LNumber(int(tank.State))

	// Вызываем Lua функцию через engine с явными параметрами
	results, err := uc.luaEngine.CallFunction(
		"updateEnemyAI",
		x,
		y,
		direction,
		state,
	)
	if err != nil {
		return types.EnemyAIDecision{}, err
	}

	// Конвертируем результат обратно в доменный тип
	return uc.typeConverter.LuaToDecision(results)
}

// Close освобождает ресурсы Lua VM
func (uc *AIUseCases) Close() {
	if uc.luaEngine != nil {
		uc.luaEngine.Close()
	}
}
