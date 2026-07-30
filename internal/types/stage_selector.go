package types

type StageSelectorEntity struct {
	CurrentStage uint
	min          uint
	max          uint
}

func NewStageSelector(maxStages uint) *StageSelectorEntity {
	return &StageSelectorEntity{
		CurrentStage: 1,
		min:          1,
		max:          maxStages,
	}
}

func (s *StageSelectorEntity) GetMinStages() uint {
	return s.min
}

func (s *StageSelectorEntity) GetMaxStages() uint {
	return s.max
}

// TitleViewData — состояние титульного экрана для отрисовки:
// плоский DTO, рендер показывает значения как есть
type TitleViewData struct {
	TwoPlayersSelected bool
	HiScore            uint
}
