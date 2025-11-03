package use_cases

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/types"
)

// StageSelectorUseCases реализация для операций с выбором уровня
type StageSelectorUseCases struct {
	// Stateless - не хранит состояние конкретных сущностей
}

// NewStageSelectorUseCases создает новый экземпляр StageSelectorUseCases
func NewStageSelectorUseCases() *StageSelectorUseCases {
	return &StageSelectorUseCases{}
}

// Next переходит на следующий уровень (циклически) и возвращает новый текущий уровень
func (uc *StageSelectorUseCases) Next(
	selector *types.StageSelectorEntity,
) uint {
	maxStage := selector.GetMaxStages()
	minStage := selector.GetMinStages()

	// Увеличиваем текущий уровень
	selector.CurrentStage++

	// Если превысили максимум, возвращаемся к минимуму
	if selector.CurrentStage > maxStage {
		selector.CurrentStage = minStage
	}

	return selector.CurrentStage
}

// Previous переходит на предыдущий уровень (циклически) и возвращает новый текущий уровень
func (uc *StageSelectorUseCases) Previous(
	selector *types.StageSelectorEntity,
) uint {
	maxStage := selector.GetMaxStages()
	minStage := selector.GetMinStages()

	// Уменьшаем текущий уровень
	selector.CurrentStage--

	// Если меньше минимума, переходим к максимуму
	if selector.CurrentStage < minStage {
		selector.CurrentStage = maxStage
	}

	return selector.CurrentStage
}

// String возвращает строковое представление выбранного уровня в формате "STAGE 01", "STAGE 19" и т.д.
// Если номер уровня меньше двух символов, добавляется ведущий ноль
func (uc *StageSelectorUseCases) String(
	selector *types.StageSelectorEntity,
) string {
	return fmt.Sprintf("STAGE %02d", selector.CurrentStage)
}

// Select подтверждает выбор уровня и возвращает номер выбранного уровня
func (uc *StageSelectorUseCases) Select(
	selector *types.StageSelectorEntity,
) uint {
	return selector.CurrentStage
}
