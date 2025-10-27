package adapters

import "github.com/shpaker/gonflict/internal/use_cases"

// ITankController общий интерфейс для контроллеров танков
// Реализуется KeyboardInputAdapter (для игрока) и AIController (для врагов)
type ITankController interface {
	// Update обновляет состояние контроллера (вызывается каждый кадр)
	Update()
}

// TankControllerBase базовая структура для контроллеров
type TankControllerBase struct {
	TankUseCases   use_cases.ITankUseCasesRef // Используем ITankUseCasesRef для вызова методов управления
	BulletUseCases use_cases.IBulletUseCases
}

// NewTankControllerBase создает базовый контроллер
func NewTankControllerBase(
	tankUseCases use_cases.ITankUseCasesRef,
	bulletUseCases use_cases.IBulletUseCases,
) *TankControllerBase {
	return &TankControllerBase{
		TankUseCases:   tankUseCases,
		BulletUseCases: bulletUseCases,
	}
}
