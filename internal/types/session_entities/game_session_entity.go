package session_entities

type GameSessionEntity struct {
	Score        int
	Level        int
	stageSession *StageSessionEntity
}

func NewGameSessionEntity() *GameSessionEntity {
	return &GameSessionEntity{
		Score:        0,
		Level:        1,
		stageSession: NewStageSessionEntity(),
	}
}

func (s *GameSessionEntity) StageSession() *StageSessionEntity {
	if s == nil {
		return nil
	}
	return s.stageSession
}
