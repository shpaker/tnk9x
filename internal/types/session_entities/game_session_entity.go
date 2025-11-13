package session_entities

import "github.com/shpaker/tnk25/internal/types"

type GameSessionEntity struct {
	Score        int
	Level        int
	targetState  *types.StateType
	stageSession *StageSessionEntity
}

func NewGameSessionEntity() *GameSessionEntity {
	initialState := types.StateTypeStageSelect
	return &GameSessionEntity{
		Score:        0,
		Level:        1,
		targetState:  &initialState,
		stageSession: NewStageSessionEntity(),
	}
}

func (s *GameSessionEntity) GetTargetState() *types.StateType {
	if s == nil {
		return nil
	}
	result := s.targetState
	s.targetState = nil
	return result
}

func (s *GameSessionEntity) SetTargetState(state *types.StateType) {
	if s != nil {
		s.targetState = state
	}
}

func (s *GameSessionEntity) StageSession() *StageSessionEntity {
	if s == nil {
		return nil
	}
	return s.stageSession
}
