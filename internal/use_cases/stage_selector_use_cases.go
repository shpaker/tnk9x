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
