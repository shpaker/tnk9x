package types

// SessionEntity представляет данные игровой сессии
// Данные передаются между уровнями и сохраняются на протяжении всей игры
type SessionEntity struct {
	PlayerLives int // Количество жизней игрока
	Score       int // Общий счёт за всю сессию
	Level       int // Текущий уровень
}

// NewSessionEntity создает новую сессию с начальными значениями
func NewSessionEntity() *SessionEntity {
	return &SessionEntity{
		PlayerLives: 3, // По умолчанию 3 жизни
		Score:       0,
		Level:       1,
	}
}

// LoseLife уменьшает количество жизней игрока на 1
func (s *SessionEntity) LoseLife() {
	if s.PlayerLives > 0 {
		s.PlayerLives--
	}
}

// AddScore добавляет очки к общему счёту
func (s *SessionEntity) AddScore(points int) {
	s.Score += points
}

// IsGameOver возвращает true если у игрока не осталось жизней
func (s *SessionEntity) IsGameOver() bool {
	return s.PlayerLives <= 0
}

// NextLevel переходит на следующий уровень
func (s *SessionEntity) NextLevel() {
	s.Level++
}
