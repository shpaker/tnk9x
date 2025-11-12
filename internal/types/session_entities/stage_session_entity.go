package session_entities

import (
	"github.com/shpaker/gonflict/internal/types"
)

const (
	defaultStageRespawnDelay = 3 * 60
	defaultStagePlayer1Lives = 3
	defaultStagePlayer2Lives = 3
)

// StageSessionEntity хранит состояние текущего уровня
// TODO: добавить методы доступа по мере внедрения механик волн/врагов

type StageSessionEntity struct {
	// Прогресс уровня
	totalEnemies     uint
	spawnedEnemies   uint
	destroyedEnemies uint

	// Состояние игроков (массив жизней)
	playerLives        []uint
	playerInitialLives []uint

	// Параметры респавна
	enemyRespawnDelay uint
	enemySpawnTicks   uint

	// Ограничения
	maxActiveEnemies uint

	// Количество игроков
	playerCount uint
}

func NewStageSessionEntity() *StageSessionEntity {
	playerLives := make([]uint, 2)
	playerInitialLives := make([]uint, 2)

	// Инициализируем жизни для каждого игрока
	playerLives[types.PlayerTankNumPlayer1] = defaultStagePlayer1Lives
	playerInitialLives[types.PlayerTankNumPlayer1] = defaultStagePlayer1Lives
	playerLives[types.PlayerTankNumPlayer2] = defaultStagePlayer2Lives
	playerInitialLives[types.PlayerTankNumPlayer2] = defaultStagePlayer2Lives

	return &StageSessionEntity{
		totalEnemies:       20,
		destroyedEnemies:   0,
		playerLives:        playerLives,
		playerInitialLives: playerInitialLives,
		enemyRespawnDelay:  defaultStageRespawnDelay,
		enemySpawnTicks:    defaultStageRespawnDelay,
		maxActiveEnemies:   5,
		playerCount:        1,
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
	// Сбрасываем жизни игроков в зависимости от выбранного количества
	playerCount := int(s.GetPlayerCount())
	if playerCount < 1 {
		playerCount = 1
	}
	if playerCount > 2 {
		playerCount = 2
	}
	for i := 0; i < playerCount; i++ {
		s.playerLives[i] = s.GetPlayerInitialLives(types.PlayerTankNum(i))
	}
	s.ResetEnemySpawnCountdown()
}

// --- Управление жизнями игроков по номеру ---

// GetPlayerLives возвращает количество жизней игрока по номеру
func (s *StageSessionEntity) GetPlayerLives(num types.PlayerTankNum) uint {
	if int(num) >= 0 && int(num) < len(s.playerLives) {
		return s.playerLives[num]
	}
	return 0
}

// GetPlayerInitialLives возвращает начальное количество жизней игрока по номеру
func (s *StageSessionEntity) GetPlayerInitialLives(
	num types.PlayerTankNum,
) uint {
	if int(num) >= 0 && int(num) < len(s.playerInitialLives) {
		if s.playerInitialLives[num] == 0 {
			// Возвращаем значение по умолчанию в зависимости от номера
			if num == types.PlayerTankNumPlayer1 {
				return defaultStagePlayer1Lives
			} else if num == types.PlayerTankNumPlayer2 {
				return defaultStagePlayer2Lives
			}
		}
		return s.playerInitialLives[num]
	}
	return 0
}

// IsPlayerDefeated возвращает true, если у игрока не осталось жизней
func (s *StageSessionEntity) IsPlayerDefeated(num types.PlayerTankNum) bool {
	return s.GetPlayerLives(num) == 0
}

// SetPlayerLives устанавливает количество жизней игрока по номеру
func (s *StageSessionEntity) SetPlayerLives(
	num types.PlayerTankNum,
	lives uint,
) {
	if int(num) >= 0 && int(num) < len(s.playerLives) {
		s.playerLives[num] = lives
	}
}

// DecrementPlayerLives уменьшает количество жизней игрока по номеру
func (s *StageSessionEntity) DecrementPlayerLives(num types.PlayerTankNum) {
	if int(num) >= 0 && int(num) < len(s.playerLives) {
		if s.playerLives[num] == 0 {
			return
		}
		s.playerLives[num]--
	}
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

// SetMaxActiveEnemies устанавливает максимальное количество активных врагов
func (s *StageSessionEntity) SetMaxActiveEnemies(value uint) {
	// Ограничиваем значение диапазоном от 3 до 10
	if value < 3 {
		value = 3
	}
	if value > 10 {
		value = 10
	}
	s.maxActiveEnemies = value
}

// GetPlayerCount возвращает количество игроков
func (s *StageSessionEntity) GetPlayerCount() uint {
	return s.playerCount
}

// SetPlayerCount устанавливает количество игроков и обнуляет жизни второго игрока, если выбран один игрок
func (s *StageSessionEntity) SetPlayerCount(count uint) {
	// Ограничиваем значение диапазоном от 1 до 2
	if count < 1 {
		count = 1
	}
	if count > 2 {
		count = 2
	}
	s.playerCount = count

	// Если выбран один игрок, обнуляем жизни второго игрока
	if count == 1 {
		s.playerLives[types.PlayerTankNumPlayer2] = 0
		s.playerInitialLives[types.PlayerTankNumPlayer2] = 0
	} else {
		// Если выбран второй игрок, восстанавливаем жизни по умолчанию
		if s.playerInitialLives[types.PlayerTankNumPlayer2] == 0 {
			s.playerLives[types.PlayerTankNumPlayer2] = defaultStagePlayer2Lives
			s.playerInitialLives[types.PlayerTankNumPlayer2] = defaultStagePlayer2Lives
		}
	}
}
