package types

// BattleEntity представляет данные конкретного боя (уровня)
// Данные сбрасываются при переходе на следующий уровень
type BattleEntity struct {
	EnemiesRemaining int  // Количество оставшихся танков врагов
	EnemiesTotal     int  // Общее количество врагов на уровне
	EnemiesKilled    int  // Количество уничтоженных врагов
	IsVictory        bool // Флаг победы в битве
	IsDefeat         bool // Флаг поражения в битве
}

// NewBattleEntity создает новую битву с указанным количеством врагов
func NewBattleEntity(enemiesTotal int) *BattleEntity {
	return &BattleEntity{
		EnemiesRemaining: enemiesTotal,
		EnemiesTotal:     enemiesTotal,
		EnemiesKilled:    0,
		IsVictory:        false,
		IsDefeat:         false,
	}
}

// EnemyKilled вызывается при уничтожении врага
func (b *BattleEntity) EnemyKilled() {
	if b.EnemiesRemaining > 0 {
		b.EnemiesRemaining--
		b.EnemiesKilled++
	}

	// Проверяем победу
	if b.EnemiesRemaining == 0 {
		b.IsVictory = true
	}
}

// SetDefeat устанавливает флаг поражения
func (b *BattleEntity) SetDefeat() {
	b.IsDefeat = true
}

// IsFinished возвращает true если битва завершена (победа или поражение)
func (b *BattleEntity) IsFinished() bool {
	return b.IsVictory || b.IsDefeat
}
