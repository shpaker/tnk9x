package use_cases

import (
	"errors"

	lua "github.com/yuin/gopher-lua"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

type AIUseCases struct {
	luaEngine     interfaces.ILuaEngine
	typeConverter interfaces.IAITypeConverter
}

func NewAIUseCases(
	luaEngine interfaces.ILuaEngine,
	typeConverter interfaces.IAITypeConverter,
) *AIUseCases {
	return &AIUseCases{
		luaEngine:     luaEngine,
		typeConverter: typeConverter,
	}
}

func (uc *AIUseCases) ExecuteAI(
	tank *types.TankEntity,
) (types.EnemyAIDecision, error) {
	if tank == nil {
		return types.EnemyAIDecision{}, errors.New("tank is nil")
	}

	x := lua.LNumber(tank.Position.X)
	y := lua.LNumber(tank.Position.Y)
	direction := lua.LNumber(int(tank.Direction))
	state := lua.LNumber(int(tank.State))

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

	return uc.typeConverter.LuaToDecision(results)
}

func (uc *AIUseCases) Close() {
	if uc.luaEngine != nil {
		uc.luaEngine.Close()
	}
}
