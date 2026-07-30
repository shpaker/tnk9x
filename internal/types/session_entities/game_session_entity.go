package session_entities

// GameSessionEntity — корневая сессия приложения: держит данные
// забега и сессию этапа; используется только в internal/app
type GameSessionEntity struct {
	runSession   *RunSessionEntity
	stageSession *StageSessionEntity
}

func NewGameSessionEntity() *GameSessionEntity {
	runSession := NewRunSessionEntity()
	return &GameSessionEntity{
		runSession:   runSession,
		stageSession: NewStageSessionEntity(runSession),
	}
}

func (s *GameSessionEntity) RunSession() *RunSessionEntity {
	if s == nil {
		return nil
	}
	return s.runSession
}

func (s *GameSessionEntity) StageSession() *StageSessionEntity {
	if s == nil {
		return nil
	}
	return s.stageSession
}
