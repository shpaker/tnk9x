package session_entities

// stageSessionEntity хранит состояние текущего уровня
// TODO: добавить методы доступа по мере внедрения механик волн/врагов

type stageSessionEntity struct {
	totalEnemies     uint
	spawnedEnemies   uint
	destroyedEnemies uint
}

func NewStageSessionEntity() stageSessionEntity {
	return stageSessionEntity{
		totalEnemies:     20,
		destroyedEnemies: 0,
	}
}

func (s *stageSessionEntity) IsCompleted() bool {
	return s.destroyedEnemies >= s.totalEnemies
}

func (s *stageSessionEntity) GetNextEnemyNumber() uint {
	return s.spawnedEnemies + 1
}

func (s *stageSessionEntity) IncrementSpawnedEnemies() {
	s.spawnedEnemies++
}

func (s *stageSessionEntity) IncrementDestroyedEnemies() {
	s.destroyedEnemies++
}

func (s *stageSessionEntity) Reset() {
	s.spawnedEnemies = 0
	s.destroyedEnemies = 0
}
