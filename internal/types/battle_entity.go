package types

type BattleEntity struct {
	EnemiesRemaining int
	EnemiesTotal     int
	EnemiesKilled    int
	IsVictory        bool
	IsDefeat         bool
}

func NewBattleEntity(enemiesTotal int) *BattleEntity {
	return &BattleEntity{
		EnemiesRemaining: enemiesTotal,
		EnemiesTotal:     enemiesTotal,
		EnemiesKilled:    0,
		IsVictory:        false,
		IsDefeat:         false,
	}
}

func (b *BattleEntity) EnemyKilled() {
	if b.EnemiesRemaining > 0 {
		b.EnemiesRemaining--
		b.EnemiesKilled++
	}

	if b.EnemiesRemaining == 0 {
		b.IsVictory = true
	}
}

func (b *BattleEntity) SetDefeat() {
	b.IsDefeat = true
}

func (b *BattleEntity) IsFinished() bool {
	return b.IsVictory || b.IsDefeat
}
