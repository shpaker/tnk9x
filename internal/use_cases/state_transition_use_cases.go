package use_cases

import (
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

type StateTransitionUseCases struct{}

func NewStateTransitionUseCases() *StateTransitionUseCases {
	return &StateTransitionUseCases{}
}

func (uc *StateTransitionUseCases) ToGame(
	session *session_entities.GameSessionEntity,
	levelNumber uint,
) *session_entities.GameSessionEntity {
	if session != nil {
		session.Level = int(levelNumber)
		targetState := types.StateTypeStage
		session.SetTargetState(&targetState)
	}
	return session
}
