package input_adapters

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/shpaker/gonflict/internal/types"
)

// ConvertTankToLua конвертирует танк в Lua таблицу
func ConvertTankToLua(L *lua.LState, tank *types.TankEntity) *lua.LTable {
	t := L.NewTable()
	t.RawSetString("x", lua.LNumber(tank.Position.X))
	t.RawSetString("y", lua.LNumber(tank.Position.Y))
	t.RawSetString("direction", lua.LNumber(int(tank.Direction)))
	t.RawSetString("speed", lua.LNumber(tank.Speed))
	return t
}

// ConvertContextToLua конвертирует контекст в Lua таблицу
func ConvertContextToLua(
	L *lua.LState,
	context *types.GameAiContext,
) *lua.LTable {
	ctx := L.NewTable()

	// Добавляем игрока если есть
	if context.Player != nil {
		ctx.RawSetString("player", ConvertTankToLua(L, context.Player))
	}

	return ctx
}
