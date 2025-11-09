package session_entities

const (
	defaultStageRespawnDelay = 3 * 60
	defaultStagePlayer1Lives = 3
)

// StageSessionEntity хранит состояние текущего уровня
// TODO: добавить методы доступа по мере внедрения механик волн/врагов

type StageSessionEntity struct {
	// Прогресс уровня
	totalEnemies     uint
	spawnedEnemies   uint
	destroyedEnemies uint

	// Состояние первого игрока
	player1Lives        uint
	player1InitialLives uint

	// Параметры респавна
	enemyRespawnDelay uint
	enemySpawnTicks   uint

	// Ограничения
	maxActiveEnemies uint
}

func NewStageSessionEntity() *StageSessionEntity {
	return &StageSessionEntity{
		totalEnemies:        20,
		destroyedEnemies:    0,
		player1Lives:        defaultStagePlayer1Lives,
		player1InitialLives: defaultStagePlayer1Lives,
		enemyRespawnDelay:   defaultStageRespawnDelay,
		enemySpawnTicks:     defaultStageRespawnDelay,
		maxActiveEnemies:    5,
	}
}

// --- Проверки прогресса уровня ---

// AreAllEnemiesDefeated возвращает true, если все враги уровня уничтожены
func (s *StageSessionEntity) AreAllEnemiesDefeated() bool {
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
	s.player1Lives = s.GetPlayer1InitialLives()
	s.ResetEnemySpawnCountdown()
}

// --- Управление жизнями первого игрока ---
func (s *StageSessionEntity) GetPlayer1Lives() uint {
	return s.player1Lives
}

func (s *StageSessionEntity) GetPlayer1InitialLives() uint {
	if s.player1InitialLives == 0 {
		return defaultStagePlayer1Lives
	}
	return s.player1InitialLives
}

// IsPlayer1Defeated возвращает true, если у первого игрока не осталось жизней
func (s *StageSessionEntity) IsPlayer1Defeated() bool {
	return s.player1Lives == 0
}

func (s *StageSessionEntity) SetPlayer1Lives(lives uint) {
	s.player1Lives = lives
}

func (s *StageSessionEntity) DecrementPlayer1Lives() {
	if s.player1Lives == 0 {
		return
	}
	s.player1Lives--
}

func (s *StageSessionEntity) RegisterEnemySpawned() {
	s.IncrementSpawnedEnemies()
	s.ResetEnemySpawnCountdown()
}

func (s *StageSessionEntity) GetTotalEnemies() uint {
	return s.totalEnemies
}

func (s *StageSessionEntity) GetRemainingEnemies() uint {
	if s.destroyedEnemies >= s.totalEnemies {
		return 0
	}
	return s.totalEnemies - s.destroyedEnemies
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
		delay = defaultStageRespawnDelay
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

func (s *StageSessionEntity) GetMaxActiveEnemies() uint {
	return s.maxActiveEnemies
}
