package services

import (
	"errors"

	lua "github.com/yuin/gopher-lua"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
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

func (c *aiTypeConverter) TankToLua(
	tank *types.TankEntity,
) (*lua.LTable, error) {
	if tank == nil {
		return nil, errors.New("tank is nil")
	}

	t := c.luaEngine.NewTable()
	t.RawSetString("x", lua.LNumber(tank.Position.X))
	t.RawSetString("y", lua.LNumber(tank.Position.Y))
	t.RawSetString("direction", lua.LNumber(int(tank.Direction)))
	t.RawSetString("speed", lua.LNumber(tank.Speed))

	return t, nil
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
