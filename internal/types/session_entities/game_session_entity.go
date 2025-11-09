package session_entities

import "github.com/shpaker/gonflict/internal/types"

// GameSessionEntity представляет данные игровой сессии
// Данные передаются между уровнями и сохраняются на протяжении всей игры
type GameSessionEntity struct {
	Score        int              // Общий счёт за всю сессию
	Level        int              // Текущий уровень
	targetState  *types.StateType // Целевое состояние для перехода (nil если переход не требуется) - приватное поле
	stageSession *StageSessionEntity
}

// NewGameSessionEntity создает новую сессию с начальными значениями
func NewGameSessionEntity() *GameSessionEntity {
	initialState := types.StateTypeStageSelect // По умолчанию начинаем с выбора уровня
	return &GameSessionEntity{
		Score:        0,
		Level:        1,
		targetState:  &initialState,
		stageSession: NewStageSessionEntity(),
	}
}

// GetTargetState возвращает TargetState и обнуляет его после чтения
func (s *GameSessionEntity) GetTargetState() *types.StateType {
	if s == nil {
		return nil
	}
	result := s.targetState
	s.targetState = nil // Обнуляем после чтения
	return result
}

// SetTargetState устанавливает TargetState
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
