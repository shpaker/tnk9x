package use_cases

import (
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/types/session_entities"
)

// StateTransitionUseCases реализация для управления переходами между состояниями
type StateTransitionUseCases struct {
	// Stateless - не хранит состояние конкретных сущностей
}

// NewStateTransitionUseCases создает новый экземпляр StateTransitionUseCases
func NewStateTransitionUseCases() *StateTransitionUseCases {
	return &StateTransitionUseCases{}
}

// ToStageSelect создает переход к состоянию выбора уровня
func (uc *StateTransitionUseCases) ToStageSelect(
	session *session_entities.GameSessionEntity,
) *session_entities.GameSessionEntity {
	if session != nil {
		targetState := types.StateTypeStageSelect
		session.SetTargetState(&targetState)
	}
	return session
}

// ToGame создает переход к игровому состоянию на указанный уровень
func (uc *StateTransitionUseCases) ToGame(
	session *session_entities.GameSessionEntity,
	levelNumber uint,
) *session_entities.GameSessionEntity {
	// Обновляем уровень и целевое состояние в сессии
	if session != nil {
		session.Level = int(levelNumber)
		targetState := types.StateTypeStage
		session.SetTargetState(&targetState)
	}
	return session
}
