package types

// SessionEntity представляет данные игровой сессии
// Данные передаются между уровнями и сохраняются на протяжении всей игры
type SessionEntity struct {
	PlayerLives int        // Количество жизней игрока
	Score       int        // Общий счёт за всю сессию
	Level       int        // Текущий уровень
	targetState *StateType // Целевое состояние для перехода (nil если переход не требуется) - приватное поле
}

// NewSessionEntity создает новую сессию с начальными значениями
func NewSessionEntity() *SessionEntity {
	initialState := StateTypeStageSelect // По умолчанию начинаем с выбора уровня
	return &SessionEntity{
		PlayerLives: 3, // По умолчанию 3 жизни
		Score:       0,
		Level:       1,
		targetState: &initialState,
	}
}

// GetTargetState возвращает TargetState и обнуляет его после чтения
func (s *SessionEntity) GetTargetState() *StateType {
	if s == nil {
		return nil
	}
	result := s.targetState
	s.targetState = nil // Обнуляем после чтения
	return result
}

// SetTargetState устанавливает TargetState
func (s *SessionEntity) SetTargetState(state *StateType) {
	if s != nil {
		s.targetState = state
	}
}
