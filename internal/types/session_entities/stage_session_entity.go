package session_entities

// StageSessionEntity хранит состояние текущего уровня
// TODO: добавить методы доступа по мере внедрения механик волн/врагов

type StageSessionEntity struct {
	// Прогресс уровня
	totalEnemies     uint
	spawnedEnemies   uint
	destroyedEnemies uint

	// Параметры респавна
	enemyRespawnDelay uint
	enemySpawnTicks   uint

	// Ограничения
	maxActiveEnemies uint
}

func NewStageSessionEntity() *StageSessionEntity {
	const defaultRespawnDelay = 3 * 60
	return &StageSessionEntity{
		totalEnemies:      20,
		destroyedEnemies:  0,
		enemyRespawnDelay: defaultRespawnDelay,
		enemySpawnTicks:   defaultRespawnDelay,
		maxActiveEnemies:  5,
	}
}

// --- Проверки прогресса уровня ---

func (s *StageSessionEntity) IsCompleted() bool {
	return s.destroyedEnemies >= s.totalEnemies
}

func (s *StageSessionEntity) GetNextEnemyNumber() uint {
	return s.spawnedEnemies + 1
}

func (s *StageSessionEntity) IncrementSpawnedEnemies() {
	s.spawnedEnemies++
}

func (s *StageSessionEntity) IncrementDestroyedEnemies() {
	s.destroyedEnemies++
}

func (s *StageSessionEntity) EnemiesForSpawnCount() uint {
	return s.totalEnemies - s.spawnedEnemies
}

// --- Управление состоянием ---

func (s *StageSessionEntity) Reset() {
	s.spawnedEnemies = 0
	s.destroyedEnemies = 0
	s.ResetEnemySpawnCountdown()
}

func (s *StageSessionEntity) RegisterEnemySpawned() {
	s.IncrementSpawnedEnemies()
	s.ResetEnemySpawnCountdown()
}

func (s *StageSessionEntity) CanSpawnNextEnemy() bool {
	if s.enemySpawnTicks > 0 {
		return false
	}
	return s.EnemiesForSpawnCount() > 0
}

func (s *StageSessionEntity) UpdateEnemySpawnCountdown() {
	if s.enemySpawnTicks > 0 {
		s.enemySpawnTicks--
	}
}

func (s *StageSessionEntity) SetEnemyRespawnDelay(delay uint) {
	if delay == 0 {
		delay = 3 * 60
	}
	s.enemyRespawnDelay = delay
	if s.enemySpawnTicks == 0 || s.enemySpawnTicks > delay {
		s.enemySpawnTicks = delay
	}
}

func (s *StageSessionEntity) ResetEnemySpawnCountdown() {
	s.enemySpawnTicks = s.enemyRespawnDelay
}

// --- Геттеры ограничений ---

func (s *StageSessionEntity) MaxActiveEnemies() uint {
	return s.maxActiveEnemies
}
