package ai

import (
	"errors"

	lua "github.com/yuin/gopher-lua"

	"github.com/shpaker/gonflict/internal/interfaces"
)

type luaEngine struct {
	L *lua.LState
}

func NewLuaEngine() interfaces.ILuaEngine {
	L := lua.NewState()

	_ = L.DoString("math.randomseed(os.time())")

	return &luaEngine{L: L}
}

func (e *luaEngine) Execute(script string) error {
	return e.L.DoString(script)
}

func (e *luaEngine) CallFunction(
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

	results := make([]lua.LValue, 2)
	results[0] = e.L.Get(-2)
	results[1] = e.L.Get(-1)
	e.L.Pop(2)

	return results, nil
}

func (e *luaEngine) NewTable() *lua.LTable {
	return e.L.NewTable()
}

func (e *luaEngine) SetGlobal(name string, value lua.LValue) {
	e.L.SetGlobal(name, value)
}

func (e *luaEngine) ToBool(value lua.LValue) bool {
	if value == nil || value == lua.LNil {
		return false
	}
	if b, ok := value.(lua.LBool); ok {
		return bool(b)
	}
	return false
}

func (e *luaEngine) ToNumber(value lua.LValue) lua.LNumber {
	if value == nil || value == lua.LNil {
		return 0
	}
	if n, ok := value.(lua.LNumber); ok {
		return n
	}
	return 0
}

func (e *luaEngine) Close() {
	if e.L != nil {
		e.L.Close()
	}
}
