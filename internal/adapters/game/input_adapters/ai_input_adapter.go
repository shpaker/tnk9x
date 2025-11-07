package input_adapters

import (
	"math"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// AiInputAdapter адаптер для работы с AI через Lua скрипты
// Теперь использует AIUseCases вместо прямой работы с Lua
type AiInputAdapter struct {
	tankActions    interfaces.ITankActionsUseCases
	tank           *types.TankEntity
	updateInterval int
	tickCounter    int
	aiUseCases     *use_cases.AIUseCases
	lastShotTick   int // Тик последней стрельбы
	shootCooldown  int // Количество тиков между выстрелами (20)
}

// NewAiInputAdapter создает новый AI адаптер
func NewAiInputAdapter(
	tankActions interfaces.ITankActionsUseCases,
	tank *types.TankEntity,
	updateInterval int,
	aiUseCases *use_cases.AIUseCases,
) (*AiInputAdapter, error) {
	return &AiInputAdapter{
		tankActions:    tankActions,
		tank:           tank,
		updateInterval: updateInterval,
		tickCounter:    0,
		aiUseCases:     aiUseCases,
		lastShotTick:   -20, // Инициализируем так, чтобы можно было стрелять сразу
		shootCooldown:  20,  // 20 тиков между выстрелами
	}, nil
}

// Update обновляет AI логику для танка
func (a *AiInputAdapter) Update(dt float64) {
	// Пропускаем неактивных врагов
	if a.tank == nil || !a.tank.IsActive() {
		return
	}

	// Увеличиваем счетчик тиков
	a.tickCounter++

	// Проверяем координаты кратные 8 (+/- 2) для установки состояния braking
	// Только если танк движется и не находится уже в состоянии Braking
	if a.tank.State == types.TankStateMoving {
		a.checkAndSetBraking()
	}

	// Запускаем AI когда танк остановлен
	if a.tank.IsStopped() {
		a.updateAI()
	}
}

// updateAI выполняет AI логику и применяет решение к танку
func (a *AiInputAdapter) updateAI() {
	// Используем AIUseCases для выполнения AI логики
	if a.aiUseCases != nil {
		decision, err := a.aiUseCases.ExecuteAI(a.tank)
		if err == nil && a.tank != nil {
			// Применяем решение (даже если Direction == 0, это валидное направление UP)
			a.tankActions.ApplyDecision(a.tank, decision)

			// Проверяем, можно ли стрелять (прошло достаточно тиков с последнего выстрела)
			if a.canShoot() {
				_ = a.tankActions.Shoot(a.tank)
				a.lastShotTick = a.tickCounter
			}
		}
	}
}

// canShoot проверяет, можно ли стрелять (прошло достаточно тиков с последнего выстрела)
func (a *AiInputAdapter) canShoot() bool {
	ticksSinceLastShot := a.tickCounter - a.lastShotTick
	return ticksSinceLastShot >= a.shootCooldown
}

// checkAndSetBraking проверяет координаты кратные 8 (+/- 2) и устанавливает состояние braking
func (a *AiInputAdapter) checkAndSetBraking() {
	if a.tank == nil {
		return
	}

	// Получаем координату по направлению движения
	var coord float64
	switch a.tank.Direction {
	case types.DirectionUp, types.DirectionDown:
		coord = a.tank.Position.Y
	case types.DirectionLeft, types.DirectionRight:
		coord = a.tank.Position.X
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
		a.tank.State = types.TankStateBraking
	}
}
