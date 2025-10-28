package adapters

import (
	"log"

	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// AIController контроллер для управления танком через AI
type AIController struct {
	*TankControllerBase
	aiUseCases  *use_cases.AIUseCases
	enemy       *types.TankEntity
	tickCounter int // Счетчик тиков для управления частотой обновления AI
}

// NewAIController создает новый AI контроллер
func NewAIController(
	tankUseCases use_cases.ITankUseCasesRef,
	bulletUseCases use_cases.IBulletUseCases,
	aiUseCases *use_cases.AIUseCases,
	enemy *types.TankEntity,
) *AIController {
	return &AIController{
		TankControllerBase: NewTankControllerBase(tankUseCases, bulletUseCases),
		aiUseCases:         aiUseCases,
		enemy:              enemy,
		tickCounter:        0,
	}
}

// Update обрабатывает AI логику для танка
func (ai *AIController) Update() {
	if ai.enemy == nil {
		return
	}

	// Пропускаем взрывающихся или не заспавненных врагов
	if ai.enemy.State == types.TankStateExploding || ai.enemy.State == types.TankStateSpawning {
		return
	}

	// Увеличиваем счетчик тиков
	ai.tickCounter++

	// Проверяем, нужно ли обновлять AI
	if ai.tickCounter >= ai.aiUseCases.GetUpdateInterval() {
		// Получаем решение от AI
		decision := ai.aiUseCases.UpdateAI(ai.enemy)

		// Применяем решение
		ai.aiUseCases.ApplyDecision(ai.enemy, decision)

		// Сбрасываем счетчик
		ai.tickCounter = 0
	}

	// Двигаем танк
	ai.moveTank()
}

// moveTank двигает танк в его текущем направлении
func (ai *AIController) moveTank() {
	if ai.enemy == nil {
		return
	}

	// Пропускаем взрывающихся или не заспавненных врагов
	if ai.enemy.State == types.TankStateExploding || ai.enemy.State == types.TankStateSpawning {
		return
	}

	// Двигаем танк через TankUseCases
	if err := ai.TankUseCases.MoveTank(ai.enemy.Direction, use_cases.DT); err != nil {
		log.Printf("ERROR: Failed to move AI tank: %v", err)
	}
}

// RotateTank поворачивает танк в указанном направлении
func (ai *AIController) RotateTank(direction types.Direction) {
	if ai.enemy == nil {
		return
	}
	if err := ai.TankUseCases.RotateTank(direction); err != nil {
		log.Printf("ERROR: Failed to rotate AI tank: %v", err)
	}
}

// StopTank останавливает танк
func (ai *AIController) StopTank() {
	if ai.enemy == nil {
		return
	}
	if err := ai.TankUseCases.StopTank(false); err != nil {
		log.Printf("ERROR: Failed to stop AI tank: %v", err)
	}
}
