package session_entities

type GameSessionEntity struct {
	Score        int
	Level        int
	stageSession *StageSessionEntity
	// victories — выигранных уровней за сессию
	victories uint
}

func NewGameSessionEntity() *GameSessionEntity {
	return &GameSessionEntity{
		Score:        0,
		Level:        1,
		stageSession: NewStageSessionEntity(),
	}
}

// AddVictory засчитывает выигранный уровень
func (s *GameSessionEntity) AddVictory() { s.victories++ }

// GetVictories — выигранных уровней за сессию
func (s *GameSessionEntity) GetVictories() uint { return s.victories }

func (s *GameSessionEntity) StageSession() *StageSessionEntity {
	if s == nil {
		return nil
	}
	return s.stageSession
}
