package input_adapters

import (
	"log"

	lua "github.com/yuin/gopher-lua"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// AiInputAdapter адаптер для работы с AI через Lua скрипты
type AiInputAdapter struct {
	tankUseCases   interfaces.ITankUseCasesRef
	aiContext      *types.GameAiContext
	updateInterval int
	tickCounter    int
	L              *lua.LState
}

// NewAiInputAdapter создает новый AI адаптер
func NewAiInputAdapter(
	tankUseCases interfaces.ITankUseCasesRef,
	aiContext *types.GameAiContext,
	updateInterval int,
	script string,
) (*AiInputAdapter, error) {
	L := lua.NewState()

	// Инициализируем генератор случайных чисел
	L.DoString("math.randomseed(os.time())")

	// Загружаем скрипт из строки
	if err := L.DoString(script); err != nil {
		L.Close()
		return nil, err
	}

	return &AiInputAdapter{
		tankUseCases:   tankUseCases,
		aiContext:      aiContext,
		updateInterval: updateInterval,
		tickCounter:    0,
		L:              L,
	}, nil
}

// Update обновляет AI логику для танка
func (a *AiInputAdapter) Update() {
	// Пропускаем неактивных врагов
	if !a.tankUseCases.IsActive() {
		return
	}

	// Проверяем, нужно ли обновлять AI
	if a.tickCounter == 0 {
		// Если есть Lua, используем его напрямую
		if a.L != nil && a.aiContext != nil {
			shouldMove, directionInt := a.CallEnemyAI(a.aiContext)
			if shouldMove {
				decision := types.EnemyAIDecision{
					Direction: types.Direction(directionInt),
				}
				// Применяем решение
				a.applyDecision(decision)
			}
		}
	}
	// Увеличиваем счетчик тиков
	a.tickCounter++

	if a.tickCounter >= a.updateInterval {
		a.tickCounter = 0
	}

	// Двигаем танк
	if err := a.tankUseCases.Update(use_cases.DT); err != nil {
		log.Printf("ERROR: Failed to update AI tank: %v", err)
	}
}

// CallEnemyAI вызывает Lua функцию для AI врага
func (a *AiInputAdapter) CallEnemyAI(
	context *types.GameAiContext,
) (bool, int) {
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

// ApplyDecision применяет решение AI к танку
func (a *AiInputAdapter) applyDecision(
	decision types.EnemyAIDecision,
) {
	if a.tankUseCases.IsStopped() {
		a.tankUseCases.Rotate(decision.Direction)
		a.tankUseCases.Move()
	}
}

// convertTankToLua конвертирует танк в Lua таблицу
func (a *AiInputAdapter) convertTankToLua(
	tank *types.TankEntity,
) *lua.LTable {
	t := a.L.NewTable()
	t.RawSetString("x", lua.LNumber(tank.Position.X))
	t.RawSetString("y", lua.LNumber(tank.Position.Y))
	t.RawSetString("direction", lua.LNumber(int(tank.Direction)))
	t.RawSetString("speed", lua.LNumber(tank.Speed))
	return t
}

// convertContextToLua конвертирует контекст в Lua таблицу
func (a *AiInputAdapter) convertContextToLua(
	context *types.GameAiContext,
) *lua.LTable {
	ctx := a.L.NewTable()

	// Добавляем игрока если есть
	if context.Player != nil {
		ctx.RawSetString("player", a.convertTankToLua(context.Player))
	}

	return ctx
}
