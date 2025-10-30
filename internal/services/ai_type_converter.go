package services

import (
	"errors"

	lua "github.com/yuin/gopher-lua"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// IAITypeConverter интерфейс для конвертации типов между Go и Lua (Application Layer)
type IAITypeConverter interface {
	// TankToLua конвертирует TankEntity в Lua таблицу
	TankToLua(tank *types.TankEntity) (*lua.LTable, error)

	// ContextToLua конвертирует GameAiContext в Lua таблицу
	ContextToLua(context *types.GameAiContext) (*lua.LTable, error)

	// LuaToDecision конвертирует результаты Lua функции в EnemyAIDecision
	LuaToDecision(results []lua.LValue) (types.EnemyAIDecision, error)
}

// aiTypeConverterImpl реализация IAITypeConverter
type aiTypeConverterImpl struct {
	luaEngine interfaces.ILuaEngine
}

// NewAITypeConverter создает новый конвертер типов AI
func NewAITypeConverter(luaEngine interfaces.ILuaEngine) IAITypeConverter {
	return &aiTypeConverterImpl{
		luaEngine: luaEngine,
	}
}

// TankToLua конвертирует TankEntity в Lua таблицу
func (c *aiTypeConverterImpl) TankToLua(
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

// ContextToLua конвертирует GameAiContext в Lua таблицу
func (c *aiTypeConverterImpl) ContextToLua(
	context *types.GameAiContext,
) (*lua.LTable, error) {
	if context == nil {
		return nil, errors.New("context is nil")
	}

	ctx := c.luaEngine.NewTable()

	// Добавляем игрока если есть
	if context.Player != nil {
		playerTable, err := c.TankToLua(context.Player)
		if err != nil {
			return nil, err
		}
		ctx.RawSetString("player", playerTable)
	}

	// Можно добавить другие поля контекста
	// Например, enemies, bullets, blocks и т.д.

	return ctx, nil
}

// LuaToDecision конвертирует результаты Lua функции в EnemyAIDecision
func (c *aiTypeConverterImpl) LuaToDecision(
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
