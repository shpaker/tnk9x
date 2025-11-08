package types

// StateType тип состояния игры
type StateType int

const (
	StateTypeStage StateType = iota
	StateTypeStageSelect
)
