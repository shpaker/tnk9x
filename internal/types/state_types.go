package types

// TransitionTarget — целевое состояние приложения
type TransitionTarget int

const (
	TransitionNone TransitionTarget = iota
	TransitionToTitle
	TransitionToCurtain
	TransitionToStage
	TransitionToScore
	TransitionToGameOver
)

// StateTransition — запрос смены состояния, возвращаемый из Update стейта;
// нулевое значение означает «остаться в текущем состоянии»
type StateTransition struct {
	Target      TransitionTarget
	Level       uint
	PlayerCount uint

	// StageWon — исход этапа для экрана итогов
	StageWon bool

	// NewRun — переход с титула: забег сбрасывается,
	// на шторке доступен выбор этапа
	NewRun bool
}
