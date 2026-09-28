package scripting

import (
	"errors"

	lua "github.com/yuin/gopher-lua"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IAIScriptEngine = (*luaEngine)(nil)

// luaEngine — единственная точка интеграции с gopher-lua: наружу
// отдаются только доменные типы. Вся логика поведения — в скрипте;
// движок передаёт снимок мира, исполняет запросы таблицы ai
// и разбирает решение.
type luaEngine struct {
	L *lua.LState

	// Services
	navigationService interfaces.INavigationService

	// Снимок мира текущего вызова UpdateEnemyAI; nil вне вызова
	context *types.EnemyAIContext
}

func NewLuaEngine(
	navigationService interfaces.INavigationService,
) interfaces.IAIScriptEngine {
	L := lua.NewState()

	_ = L.DoString("math.randomseed(os.time())")

	engine := &luaEngine{
		L:                 L,
		navigationService: navigationService,
	}
	engine.registerAPI()

	return engine
}

func (e *luaEngine) LoadScript(source string) error {
	return e.L.DoString(source)
}

func (e *luaEngine) SetGlobalNumber(name string, value float64) {
	e.L.SetGlobal(name, lua.LNumber(value))
}

// UpdateEnemyAI вызывает Lua-функцию updateEnemyAI(ctx), возвращающую
// таблицу {direction=, move=, shoot=}. Отсутствующие поля — безопасные
// значения: текущее направление, стоять, не стрелять.
// Реализует interfaces.IAIScriptEngine.
func (e *luaEngine) UpdateEnemyAI(
	context types.EnemyAIContext,
) (types.EnemyAIDecision, error) {
	fn := e.L.GetGlobal("updateEnemyAI")
	if fn == lua.LNil {
		return types.EnemyAIDecision{}, errors.New(
			"function not found: updateEnemyAI",
		)
	}

	e.context = &context
	defer func() { e.context = nil }()

	err := e.L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	}, e.contextTable(context))
	if err != nil {
		return types.EnemyAIDecision{}, err
	}

	result := e.L.Get(-1)
	e.L.Pop(1)

	decision := types.EnemyAIDecision{Direction: context.Self.Direction}
	table, ok := result.(*lua.LTable)
	if !ok {
		return decision, nil
	}

	if n, ok := table.RawGetString("direction").(lua.LNumber); ok {
		decision.Direction = types.Direction(int(n))
	}
	decision.Move = lua.LVAsBool(table.RawGetString("move"))
	decision.Shoot = lua.LVAsBool(table.RawGetString("shoot"))

	return decision, nil
}

func (e *luaEngine) Close() {
	if e.L != nil {
		e.L.Close()
	}
}

// Снимок мира

func (e *luaEngine) contextTable(context types.EnemyAIContext) *lua.LTable {
	table := e.L.NewTable()
	table.RawSetString("tick", lua.LNumber(context.Tick))
	table.RawSetString("stage", lua.LNumber(context.StageNumber))
	table.RawSetString("self", e.tankTable(context.Self))

	players := e.L.NewTable()
	for _, player := range context.Players {
		players.Append(e.tankTable(player))
	}
	table.RawSetString("players", players)

	enemies := e.L.NewTable()
	for _, enemy := range context.Enemies {
		enemies.Append(e.tankTable(enemy))
	}
	table.RawSetString("enemies", enemies)

	bullets := e.L.NewTable()
	for _, bullet := range context.Bullets {
		item := e.L.NewTable()
		item.RawSetString("x", lua.LNumber(bullet.Position.X))
		item.RawSetString("y", lua.LNumber(bullet.Position.Y))
		item.RawSetString("w", lua.LNumber(bullet.Size.Width))
		item.RawSetString("h", lua.LNumber(bullet.Size.Height))
		item.RawSetString("dir", lua.LNumber(bullet.Direction))
		item.RawSetString("enemy", lua.LBool(bullet.FromEnemy))
		bullets.Append(item)
	}
	table.RawSetString("bullets", bullets)

	hq := e.L.NewTable()
	hq.RawSetString("x", lua.LNumber(context.HQ.Position.X))
	hq.RawSetString("y", lua.LNumber(context.HQ.Position.Y))
	hq.RawSetString("size", lua.LNumber(context.HQ.Size.Width))
	hq.RawSetString("intact", lua.LBool(context.HQ.Intact))
	table.RawSetString("hq", hq)

	return table
}

func (e *luaEngine) tankTable(tank types.AITankInfo) *lua.LTable {
	table := e.L.NewTable()
	table.RawSetString("id", lua.LNumber(tank.ID))
	table.RawSetString("x", lua.LNumber(tank.Position.X))
	table.RawSetString("y", lua.LNumber(tank.Position.Y))
	table.RawSetString("size", lua.LNumber(tank.Size.Width))
	table.RawSetString("dir", lua.LNumber(tank.Direction))
	table.RawSetString("level", lua.LNumber(tank.Level))
	table.RawSetString("hp", lua.LNumber(tank.HitPoints))
	table.RawSetString("bonus", lua.LBool(tank.WithBonus))
	table.RawSetString("reinforced", lua.LBool(tank.Reinforced))
	return table
}

// API скрипта: таблица ai

func (e *luaEngine) registerAPI() {
	api := e.L.NewTable()
	e.L.SetFuncs(api, map[string]lua.LGFunction{
		"findPath": e.findPath,
		"castRay":  e.castRay,
		"tileAt":   e.tileAt,
	})
	e.L.SetGlobal("ai", api)
}

// findPath(fromX, fromY, toX, toY [, {brickCost=, steelPassable=}])
// возвращает направление первого шага и длину пути либо nil
func (e *luaEngine) findPath(L *lua.LState) int {
	if e.context == nil {
		L.Push(lua.LNil)
		return 1
	}

	from := types.Position{X: float64(L.CheckNumber(1)), Y: float64(L.CheckNumber(2))}
	to := types.Position{X: float64(L.CheckNumber(3)), Y: float64(L.CheckNumber(4))}
	options := types.NavOptions{TankSize: e.context.Self.Size.Width}
	if table, ok := L.Get(5).(*lua.LTable); ok {
		if n, ok := table.RawGetString("brickCost").(lua.LNumber); ok {
			options.BrickCost = int(n)
		}
		options.SteelPassable = lua.LVAsBool(table.RawGetString("steelPassable"))
	}

	step, found := e.navigationService.FindPath(e.context.Grid, from, to, options)
	if !found {
		L.Push(lua.LNil)
		return 1
	}

	L.Push(lua.LNumber(step.Direction))
	L.Push(lua.LNumber(step.Length))
	return 2
}

// castRay(x, y, direction) возвращает тип первого препятствия на линии
// огня ("edge", "brick", "steel", "player", "enemy", "hq") и расстояние
func (e *luaEngine) castRay(L *lua.LState) int {
	if e.context == nil {
		L.Push(lua.LNil)
		return 1
	}

	origin := types.Position{X: float64(L.CheckNumber(1)), Y: float64(L.CheckNumber(2))}
	direction := types.Direction(L.CheckInt(3))

	hit := e.navigationService.CastRay(
		e.context.Grid,
		origin,
		direction,
		e.rayTargets(),
	)

	L.Push(lua.LString(hit.Kind))
	L.Push(lua.LNumber(hit.Distance))
	return 2
}

func (e *luaEngine) rayTargets() []types.RayTarget {
	context := e.context
	targets := make([]types.RayTarget, 0, len(context.Players)+len(context.Enemies)+1)
	for _, player := range context.Players {
		targets = append(targets, types.RayTarget{
			Kind:     types.RayHitPlayer,
			Position: player.Position,
			Size:     player.Size,
		})
	}
	for _, enemy := range context.Enemies {
		targets = append(targets, types.RayTarget{
			Kind:     types.RayHitEnemy,
			Position: enemy.Position,
			Size:     enemy.Size,
		})
	}
	if context.HQ.Intact {
		targets = append(targets, types.RayTarget{
			Kind:     types.RayHitHQ,
			Position: context.HQ.Position,
			Size:     context.HQ.Size,
		})
	}
	return targets
}

// tileAt(x, y) возвращает тип блока в точке карты либо nil
func (e *luaEngine) tileAt(L *lua.LState) int {
	if e.context == nil || e.context.Grid == nil {
		L.Push(lua.LNil)
		return 1
	}

	block := e.context.Grid.GetBlockAt(
		float64(L.CheckNumber(1)),
		float64(L.CheckNumber(2)),
	)
	if block == "" {
		L.Push(lua.LNil)
		return 1
	}

	L.Push(lua.LString(block))
	return 1
}
