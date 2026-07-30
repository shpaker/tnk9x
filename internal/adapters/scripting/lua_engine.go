package scripting

import (
	"errors"

	lua "github.com/yuin/gopher-lua"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IAIScriptEngine = (*luaEngine)(nil)

// luaEngine — единственная точка интеграции с gopher-lua: наружу
// отдаются только доменные типы.
type luaEngine struct {
	L *lua.LState
}

func NewLuaEngine() interfaces.IAIScriptEngine {
	L := lua.NewState()

	_ = L.DoString("math.randomseed(os.time())")

	return &luaEngine{L: L}
}

func (e *luaEngine) LoadScript(source string) error {
	return e.L.DoString(source)
}

func (e *luaEngine) SetGlobalNumber(name string, value float64) {
	e.L.SetGlobal(name, lua.LNumber(value))
}

// UpdateEnemyAI вызывает Lua-функцию updateEnemyAI(x, y, direction,
// state, phase, targetX, targetY, hasTarget), возвращающую пару
// (shouldMove, direction). При shouldMove == false возвращается
// нулевое решение без ошибки.
func (e *luaEngine) UpdateEnemyAI(
	x, y float64,
	direction, state int,
	context types.EnemyAIContext,
) (types.EnemyAIDecision, error) {
	fn := e.L.GetGlobal("updateEnemyAI")
	if fn == lua.LNil {
		return types.EnemyAIDecision{}, errors.New(
			"function not found: updateEnemyAI",
		)
	}

	err := e.L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    2,
		Protect: true,
	},
		lua.LNumber(x),
		lua.LNumber(y),
		lua.LNumber(direction),
		lua.LNumber(state),
		lua.LNumber(context.Phase),
		lua.LNumber(context.TargetX),
		lua.LNumber(context.TargetY),
		lua.LBool(context.HasTarget),
	)
	if err != nil {
		return types.EnemyAIDecision{}, err
	}

	shouldMoveValue := e.L.Get(-2)
	directionValue := e.L.Get(-1)
	e.L.Pop(2)

	shouldMove := false
	if b, ok := shouldMoveValue.(lua.LBool); ok {
		shouldMove = bool(b)
	}
	if !shouldMove {
		return types.EnemyAIDecision{}, nil
	}

	directionNumber := lua.LNumber(0)
	if n, ok := directionValue.(lua.LNumber); ok {
		directionNumber = n
	}

	return types.EnemyAIDecision{
		Direction: types.Direction(int(directionNumber)),
	}, nil
}

func (e *luaEngine) Close() {
	if e.L != nil {
		e.L.Close()
	}
}
