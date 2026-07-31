package types

// TransitionTarget — целевое состояние приложения
type TransitionTarget int

const (
	TransitionNone TransitionTarget = iota
	TransitionToStage
	TransitionToStageSelect
	TransitionToQuit
)

// StateTransition — запрос смены состояния, возвращаемый из Update стейта;
// нулевое значение означает «остаться в текущем состоянии»
type StateTransition struct {
	Target           TransitionTarget
	Level            uint
	PlayerCount      uint
	MaxActiveEnemies uint
}
