package ai

import (
	"errors"

	lua "github.com/yuin/gopher-lua"

	"github.com/shpaker/gonflict/internal/interfaces"
)

// luaEngineImpl реализация interfaces.ILuaEngine
type luaEngineImpl struct {
	L *lua.LState
}

// NewLuaEngine создает новый Lua engine
func NewLuaEngine() interfaces.ILuaEngine {
	L := lua.NewState()
	// Инициализируем генератор случайных чисел
	L.DoString("math.randomseed(os.time())")

	return &luaEngineImpl{L: L}
}

// Execute выполняет Lua скрипт из строки
func (e *luaEngineImpl) Execute(script string) error {
	return e.L.DoString(script)
}

// CallFunction вызывает Lua функцию с параметрами и возвращает результаты
func (e *luaEngineImpl) CallFunction(
	functionName string,
	args ...lua.LValue,
) ([]lua.LValue, error) {
	fn := e.L.GetGlobal(functionName)
	if fn == lua.LNil {
		return nil, errors.New("function not found: " + functionName)
	}

	err := e.L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    2,
		Protect: true,
	}, args...)
	if err != nil {
		return nil, err
	}

	// Собираем результаты
	results := make([]lua.LValue, 2)
	results[0] = e.L.Get(-2)
	results[1] = e.L.Get(-1)
	e.L.Pop(2)

	return results, nil
}

// NewTable создает новую Lua таблицу
func (e *luaEngineImpl) NewTable() *lua.LTable {
	return e.L.NewTable()
}

// ToBool конвертирует Lua значение в bool
func (e *luaEngineImpl) ToBool(value lua.LValue) bool {
	if value == nil || value == lua.LNil {
		return false
	}
	if b, ok := value.(lua.LBool); ok {
		return bool(b)
	}
	return false
}

// ToNumber конвертирует Lua значение в число
func (e *luaEngineImpl) ToNumber(value lua.LValue) lua.LNumber {
	if value == nil || value == lua.LNil {
		return 0
	}
	if n, ok := value.(lua.LNumber); ok {
		return n
	}
	return 0
}

// Close освобождает ресурсы Lua VM
func (e *luaEngineImpl) Close() {
	if e.L != nil {
		e.L.Close()
	}
}
