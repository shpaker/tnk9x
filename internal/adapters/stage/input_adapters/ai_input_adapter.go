package input_adapters

import (
	"math"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// Проверяем, что AiInputAdapter реализует интерфейс IAiInputAdapter
var _ interfaces.IAiInputAdapter = (*AiInputAdapter)(nil)

// AiInputAdapter адаптер для работы с AI через Lua скрипты
// Теперь использует AIUseCases вместо прямой работы с Lua
type AiInputAdapter struct {
	tankActions    interfaces.ITankActionsUseCases
	tanks          []*types.TankEntity
	updateInterval int
	tickCounter    int
	aiUseCases     *use_cases.AIUseCases
	lastShotTick   map[*types.TankEntity]int // Тик последней стрельбы по каждому танку
	shootCooldown  int                       // Количество тиков между выстрелами (20)
}

// NewAiInputAdapter создает новый AI адаптер
func NewAiInputAdapter(
	tankActions interfaces.ITankActionsUseCases,
	tank *types.TankEntity,
	updateInterval int,
	aiUseCases *use_cases.AIUseCases,
) (*AiInputAdapter, error) {
	adapter := &AiInputAdapter{
		tankActions:    tankActions,
		tanks:          make([]*types.TankEntity, 0, 1),
		updateInterval: updateInterval,
		tickCounter:    0,
		aiUseCases:     aiUseCases,
		lastShotTick:   make(map[*types.TankEntity]int),
		shootCooldown:  20, // 20 тиков между выстрелами
	}

	if tank != nil {
		adapter.AddTank(tank)
	}

	return adapter, nil
}

// Update обновляет AI логику для танка
func (a *AiInputAdapter) Update(dt float64) {
	// Увеличиваем счетчик тиков
	a.tickCounter++

	if len(a.tanks) == 0 {
		return
	}

	for _, tank := range a.tanks {
		if tank == nil || !tank.IsActive() {
			continue
		}

		// Проверяем координаты кратные 8 (+/- 2) для установки состояния braking
		if tank.State == types.TankStateMoving {
			a.checkAndSetBraking(tank)
		}

		// Запускаем AI когда танк остановлен
		if tank.IsStopped() {
			a.updateAI(tank)
		}
	}
}

// updateAI выполняет AI логику и применяет решение к танку
func (a *AiInputAdapter) updateAI(tank *types.TankEntity) {
	// Используем AIUseCases для выполнения AI логики
	if a.aiUseCases != nil {
		decision, err := a.aiUseCases.ExecuteAI(tank)
		if err == nil && tank != nil {
			// Применяем решение (даже если Direction == 0, это валидное направление UP)
			a.tankActions.ApplyDecision(tank, decision)

			// Проверяем, можно ли стрелять (прошло достаточно тиков с последнего выстрела)
			if a.canShoot(tank) {
				_ = a.tankActions.Shoot(tank)
				a.lastShotTick[tank] = a.tickCounter
			}
		}
	}
}

// canShoot проверяет, можно ли стрелять (прошло достаточно тиков с последнего выстрела)
func (a *AiInputAdapter) canShoot(tank *types.TankEntity) bool {
	lastShotTick, ok := a.lastShotTick[tank]
	if !ok {
		lastShotTick = -a.shootCooldown
	}

	ticksSinceLastShot := a.tickCounter - lastShotTick
	return ticksSinceLastShot >= a.shootCooldown
}

// checkAndSetBraking проверяет координаты кратные 8 (+/- 2) и устанавливает состояние braking
func (a *AiInputAdapter) checkAndSetBraking(tank *types.TankEntity) {
	if tank == nil {
		return
	}

	// Получаем координату по направлению движения
	var coord float64
	switch tank.Direction {
	case types.DirectionUp, types.DirectionDown:
		coord = tank.Position.Y
	case types.DirectionLeft, types.DirectionRight:
		coord = tank.Position.X
	default:
		return
	}

	// Проверяем, близка ли координата к кратному 8 (+/- 2)
	// Вычисляем остаток от деления на 8
	remainder := math.Mod(coord, 8)
	if remainder < 0 {
		remainder += 8
	}

	// Если координата близка к кратному 8 (+/- 2)
	// remainder <= 2 означает близость к 0, 8, 16, ...
	// remainder >= 6 означает близость к 8, 16, 24, ... (через отрицательный остаток)
	if remainder <= 2 || remainder >= 6 {
		// Устанавливаем состояние braking через Rotate с тем же направлением
		// Rotate проверяет, что танк в состоянии Moving, поэтому не будет
		// переключать в Braking если уже в Braking
		tank.State = types.TankStateBraking
	}
}

// AddTank добавляет танк под контроль AI адаптера
func (a *AiInputAdapter) AddTank(tank *types.TankEntity) {
	if tank == nil {
		return
	}

	for _, current := range a.tanks {
		if current == tank {
			return
		}
	}

	a.tanks = append(a.tanks, tank)
	if a.lastShotTick == nil {
		a.lastShotTick = make(map[*types.TankEntity]int)
	}
	a.lastShotTick[tank] = a.tickCounter - a.shootCooldown
}

// RemoveTank удаляет танк из управления AI адаптера
func (a *AiInputAdapter) RemoveTank(tank *types.TankEntity) {
	if tank == nil {
		return
	}

	for i, current := range a.tanks {
		if current == tank {
			a.tanks = append(a.tanks[:i], a.tanks[i+1:]...)
			break
		}
	}

	if a.lastShotTick != nil {
		delete(a.lastShotTick, tank)
	}
}
