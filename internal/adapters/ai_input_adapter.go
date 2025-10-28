package adapters

import (
	"log"

	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
	lua "github.com/yuin/gopher-lua"
)

// AiInputAdapter адаптер для работы с AI через Lua скрипты
type AiInputAdapter struct {
	tankUseCases use_cases.ITankUseCasesRef
	aiUseCases   *use_cases.AIUseCases
	tickCounter  int
	L            *lua.LState
}

// NewAiInputAdapter создает новый AI адаптер
func NewAiInputAdapter(
	tankUseCases use_cases.ITankUseCasesRef,
	aiUseCases *use_cases.AIUseCases,
	scriptPath string,
) (*AiInputAdapter, error) {
	L := lua.NewState()

	// Инициализируем генератор случайных чисел
	L.DoString("math.randomseed(os.time())")

	// Загружаем скрипт
	if err := L.DoFile(scriptPath); err != nil {
		L.Close()
		return nil, err
	}

	return &AiInputAdapter{
		tankUseCases: tankUseCases,
		aiUseCases:   aiUseCases,
		tickCounter:  0,
		L:            L,
	}, nil
}

// Update обновляет AI логику для танка
func (a *AiInputAdapter) Update() {
	// Получаем танк через use cases
	tank := a.tankUseCases.GetTank()
	if tank == nil {
		return
	}

	// Пропускаем взрывающихся или не заспавненных врагов
	if tank.State == types.TankStateExploding || tank.State == types.TankStateSpawning {
		return
	}

	// Увеличиваем счетчик тиков
	a.tickCounter++

	// Проверяем, нужно ли обновлять AI
	if a.tickCounter >= a.aiUseCases.GetUpdateInterval() {
		// Если есть Lua, используем его напрямую
		if a.L != nil && a.aiUseCases.GetAIContext() != nil {
			shouldMove, directionInt := a.CallEnemyAI(a.aiUseCases.GetAIContext())
			decision := types.EnemyAIDecision{
				ShouldMove:   shouldMove,
				NewDirection: IntToDirection(directionInt),
			}
			// Применяем решение
			a.aiUseCases.ApplyDecision(tank, decision)
		}

		// Сбрасываем счетчик
		a.tickCounter = 0
	}

	// Двигаем танк
	a.moveTank()
}

// moveTank двигает танк в его текущем направлении
func (a *AiInputAdapter) moveTank() {
	// Получаем танк через use cases
	tank := a.tankUseCases.GetTank()
	if tank == nil {
		return
	}

	// Пропускаем взрывающихся или не заспавненных врагов
	if tank.State == types.TankStateExploding || tank.State == types.TankStateSpawning {
		return
	}

	// Двигаем танк через TankUseCases
	if err := a.tankUseCases.MoveTank(tank.Direction, use_cases.DT); err != nil {
		log.Printf("ERROR: Failed to move AI tank: %v", err)
	}
}

// CallEnemyAI вызывает Lua функцию для AI врага
func (a *AiInputAdapter) CallEnemyAI(context *types.GameAiContext) (bool, int) {
	if a.L == nil {
		return false, 0
	}

	// Получаем танк через use cases
	tank := a.tankUseCases.GetTank()
	if tank == nil {
		return false, 0
	}

	// Конвертируем танк в Lua таблицу
	tankTable := a.convertTankToLua(tank)

	// Конвертируем контекст в Lua таблицу
	contextTable := a.convertContextToLua(context)

	// Вызываем Lua функцию
	err := a.L.CallByParam(lua.P{
		Fn:      a.L.GetGlobal("updateEnemyAI"),
		NRet:    2,
		Protect: true,
	}, tankTable, contextTable)

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
func (a *AiInputAdapter) convertTankToLua(tank *types.TankEntity) *lua.LTable {
	t := a.L.NewTable()
	t.RawSetString("x", lua.LNumber(tank.Position.X))
	t.RawSetString("y", lua.LNumber(tank.Position.Y))
	t.RawSetString("direction", lua.LNumber(DirectionToInt(tank.Direction)))
	t.RawSetString("speed", lua.LNumber(tank.Speed))
	return t
}

// convertContextToLua конвертирует контекст в Lua таблицу
func (a *AiInputAdapter) convertContextToLua(context *types.GameAiContext) *lua.LTable {
	ctx := a.L.NewTable()

	// Добавляем игрока если есть
	if context.Player != nil {
		ctx.RawSetString("player", a.convertTankToLua(context.Player))
	}

	return ctx
}

// DirectionToInt конвертирует направление в число
func DirectionToInt(d types.Direction) int {
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

// IntToDirection конвертирует число в направление
func IntToDirection(i int) types.Direction {
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
