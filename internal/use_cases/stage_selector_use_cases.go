package use_cases

import (
	"fmt"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IStageSelectorUseCases = (*StageSelectorUseCases)(nil)

type StageSelectorUseCases struct{}

func NewStageSelectorUseCases() *StageSelectorUseCases {
	return &StageSelectorUseCases{}
}

func (uc *StageSelectorUseCases) Next(
	selector *types.StageSelectorEntity,
) uint {
	maxStage := selector.GetMaxStages()
	minStage := selector.GetMinStages()

	selector.CurrentStage++

	if selector.CurrentStage > maxStage {
		selector.CurrentStage = minStage
	}

	return selector.CurrentStage
}

func (uc *StageSelectorUseCases) Previous(
	selector *types.StageSelectorEntity,
) uint {
	maxStage := selector.GetMaxStages()
	minStage := selector.GetMinStages()

	selector.CurrentStage--

	if selector.CurrentStage < minStage {
		selector.CurrentStage = maxStage
	}

	return selector.CurrentStage
}

func (uc *StageSelectorUseCases) String(
	selector *types.StageSelectorEntity,
) string {
	return fmt.Sprintf("STAGE %02d", selector.CurrentStage)
}

func (uc *StageSelectorUseCases) Select(
	selector *types.StageSelectorEntity,
) uint {
	return selector.CurrentStage
}

// Диапазон лимита одновременно активных врагов
const (
	minMaxActiveEnemies uint = 3
	maxMaxActiveEnemies uint = 10
)

// NextMaxActiveEnemies увеличивает лимит врагов с переходом 10 -> 3
func (uc *StageSelectorUseCases) NextMaxActiveEnemies(current uint) uint {
	if current < minMaxActiveEnemies || current >= maxMaxActiveEnemies {
		return minMaxActiveEnemies
	}
	return current + 1
}

// PreviousMaxActiveEnemies уменьшает лимит врагов с переходом 3 -> 10
func (uc *StageSelectorUseCases) PreviousMaxActiveEnemies(
	current uint,
) uint {
	if current <= minMaxActiveEnemies || current > maxMaxActiveEnemies {
		return maxMaxActiveEnemies
	}
	return current - 1
}
