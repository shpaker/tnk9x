package adapters

import (
	"log"

	"github.com/shpaker/gonflict/internal/types"
	lua "github.com/yuin/gopher-lua"
)

// LuaAdapter адаптер для работы с Lua скриптами
type LuaAdapter struct {
	L *lua.LState
}

// NewLuaAdapter создает новый Lua адаптер
func NewLuaAdapter(scriptPath string) (*LuaAdapter, error) {
	L := lua.NewState()

	// Инициализируем генератор случайных чисел
	L.DoString("math.randomseed(os.time())")

	// Загружаем скрипт
	if err := L.DoFile(scriptPath); err != nil {
		L.Close()
		return nil, err
	}

	return &LuaAdapter{L: L}, nil
}

// Close закрывает Lua состояние
func (a *LuaAdapter) Close() {
	if a.L != nil {
		a.L.Close()
	}
}

// CallEnemyAI вызывает Lua функцию для AI врага
func (a *LuaAdapter) CallEnemyAI(enemy *types.TankEntity, context *types.GameAIContext) (bool, int) {
	// Конвертируем врага в Lua таблицу
	enemyTable := a.convertTankToLua(enemy)

	// Конвертируем контекст в Lua таблицу
	contextTable := a.convertContextToLua(context)

	// Вызываем Lua функцию
	err := a.L.CallByParam(lua.P{
		Fn:      a.L.GetGlobal("updateEnemyAI"),
		NRet:    2,
		Protect: true,
	}, enemyTable, contextTable)

	if err != nil {
		log.Printf("Error calling Lua AI: %v", err)
		return false, 0
	}

	// Получаем результаты
	shouldMove := a.L.ToBool(1)
	directionInt := int(a.L.ToNumber(2))

	a.L.Pop(2)

	return shouldMove, directionInt
}

// convertTankToLua конвертирует танк в Lua таблицу
func (a *LuaAdapter) convertTankToLua(tank *types.TankEntity) *lua.LTable {
	t := a.L.NewTable()
	t.RawSetString("x", lua.LNumber(tank.WorldPosition.X))
	t.RawSetString("y", lua.LNumber(tank.WorldPosition.Y))
	t.RawSetString("direction", lua.LNumber(directionToInt(tank.Direction)))
	t.RawSetString("speed", lua.LNumber(tank.Speed))
	return t
}

// convertContextToLua конвертирует контекст в Lua таблицу
func (a *LuaAdapter) convertContextToLua(context *types.GameAIContext) *lua.LTable {
	ctx := a.L.NewTable()

	// Добавляем игрока если есть
	if context.Player != nil {
		ctx.RawSetString("player", a.convertTankToLua(context.Player))
	}

	return ctx
}

// directionToInt конвертирует направление в число
func directionToInt(d types.Direction) int {
	switch d {
	case types.DirectionUp:
		return 0
	case types.DirectionDown:
		return 1
	case types.DirectionLeft:
		return 2
	case types.DirectionRight:
		return 3
	default:
		return 0
	}
}

// intToDirection конвертирует число в направление
func intToDirection(i int) types.Direction {
	switch i {
	case 0:
		return types.DirectionUp
	case 1:
		return types.DirectionDown
	case 2:
		return types.DirectionLeft
	case 3:
		return types.DirectionRight
	default:
		return types.DirectionUp
	}
}
