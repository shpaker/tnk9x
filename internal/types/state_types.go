package types

// TransitionTarget — целевое состояние приложения
type TransitionTarget int

const (
	TransitionNone TransitionTarget = iota
	TransitionToStage
	TransitionToLevelSelect
	TransitionToMainMenu
	TransitionToQuit
)

// StateTransition — запрос смены состояния, возвращаемый из Update стейта;
// нулевое значение означает «остаться в текущем состоянии»
type StateTransition struct {
	Target TransitionTarget
	// Level — уровень для запуска (TransitionToStage) или уровень,
	// на который встаёт курсор экрана выбора (TransitionToLevelSelect)
	Level uint
	// CarryOver — перенести жизни и прокачку танков с прошлого уровня
	CarryOver bool
}
