package types

// type StageSessionEntity struct {
// 	enemiesCount     uint
// 	enemiesRemaining uint
// }

// GameSessionEntity представляет данные игровой сессии
// Данные передаются между уровнями и сохраняются на протяжении всей игры
type GameSessionEntity struct {
	PlayerLives int // Количество жизней игрока
	Score       int // Общий счёт за всю сессию
	Level       int // Текущий уровень
	// stageSession *StageSessionEntity
	targetState *StateType // Целевое состояние для перехода (nil если переход не требуется) - приватное поле
}

// NewGameSessionEntity создает новую сессию с начальными значениями
func NewGameSessionEntity() *GameSessionEntity {
	initialState := StateTypeStageSelect // По умолчанию начинаем с выбора уровня
	return &GameSessionEntity{
		PlayerLives: 3, // По умолчанию 3 жизни
		Score:       0,
		Level:       1,
		targetState: &initialState,
	}
}

// GetTargetState возвращает TargetState и обнуляет его после чтения
func (s *GameSessionEntity) GetTargetState() *StateType {
	if s == nil {
		return nil
	}
	result := s.targetState
	s.targetState = nil // Обнуляем после чтения
	return result
}

// SetTargetState устанавливает TargetState
func (s *GameSessionEntity) SetTargetState(state *StateType) {
	if s != nil {
		s.targetState = state
	}
}
