package types

// StateType тип состояния игры
type StateType int

const (
	StateTypeGame StateType = iota
	StateTypeStageSelect
)
