package input_adapters

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// AiInputAdapter адаптер для работы с AI через Lua скрипты
type AiInputAdapter struct {
	tankActions    interfaces.ITankActionsUseCases
	tank           *types.TankEntity
	aiContext      *types.GameAiContext
	updateInterval int
	tickCounter    int
	L              *lua.LState
}

// NewAiInputAdapter создает новый AI адаптер
func NewAiInputAdapter(
	tankActions interfaces.ITankActionsUseCases,
	tank *types.TankEntity,
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
		tankActions:    tankActions,
		tank:           tank,
		aiContext:      aiContext,
		updateInterval: updateInterval,
		tickCounter:    0,
		L:              L,
	}, nil
}

// Update обновляет AI логику для танка
func (a *AiInputAdapter) Update(dt float64) {
	// Пропускаем неактивных врагов
	if a.tank == nil || !a.tank.IsActive() {
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
}

// CallEnemyAI вызывает Lua функцию для AI врага
func (a *AiInputAdapter) CallEnemyAI(
	context *types.GameAiContext,
) (bool, int) {
	if a.L == nil || a.tank == nil {
		return false, 0
	}

	// Конвертируем танк в Lua таблицу
	tankTable := ConvertTankToLua(a.L, a.tank)

	// Конвертируем контекст в Lua таблицу
	contextTable := ConvertContextToLua(a.L, context)

	// Вызываем Lua функцию
	err := a.L.CallByParam(lua.P{
		Fn:      a.L.GetGlobal("updateEnemyAI"),
		NRet:    2,
		Protect: true,
	}, tankTable, contextTable)
	if err != nil {
		return false, 0
	}

	// Получаем результаты
	shouldMove := a.L.ToBool(1)
	directionInt := int(a.L.ToNumber(2))

	a.L.Pop(2)

	return shouldMove, directionInt
}

// applyDecision применяет решение AI к танку через TankActionsUseCases
func (a *AiInputAdapter) applyDecision(
	decision types.EnemyAIDecision,
) {
	if a.tank == nil {
		return
	}
	a.tankActions.ApplyDecision(a.tank, decision)
}
