package services

import (
	"errors"

	lua "github.com/yuin/gopher-lua"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

type aiTypeConverter struct {
	luaEngine interfaces.ILuaEngine
}

func NewAITypeConverter(
	luaEngine interfaces.ILuaEngine,
) interfaces.IAITypeConverter {
	return &aiTypeConverter{
		luaEngine: luaEngine,
	}
}

func (c *aiTypeConverter) LuaToDecision(
	results []lua.LValue,
) (types.EnemyAIDecision, error) {
	if len(results) < 2 {
		return types.EnemyAIDecision{}, errors.New(
			"insufficient results from Lua",
		)
	}

	shouldMove := c.luaEngine.ToBool(results[0])
	if !shouldMove {
		return types.EnemyAIDecision{}, nil
	}

	directionInt := int(c.luaEngine.ToNumber(results[1]))

	return types.EnemyAIDecision{
		Direction: types.Direction(directionInt),
	}, nil
}
