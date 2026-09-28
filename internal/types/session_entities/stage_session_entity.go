package session_entities

import (
	"github.com/shpaker/tnk9x/internal/types"
)

const (
	defaultStagePlayer1Lives = 3
	defaultStagePlayer2Lives = 3
)

// noSpawner — пустой слот истории спаунеров
const noSpawner = -1

// recentSpawnersCount — глубина истории использованных спаунеров
const recentSpawnersCount = 2

type StageSessionEntity struct {
	// Сценарий уровня: волны врагов и лимит активных
	level *types.LevelEntity

	totalEnemies     uint
	spawnedEnemies   uint
	destroyedEnemies uint
	// Танки, уже учтённые счётчиком destroyedEnemies
	countedDestroyedEnemies map[*types.TankEntity]struct{}

	// Курсор волн: следующая волна и танк в ней
	waveIndex     int
	waveTankIndex int
	// История спаунеров: последний использованный — первым
	recentSpawners [recentSpawnersCount]int

	playerLives        []uint
	playerInitialLives []uint
	// Потерянные жизни за уровень — для подсчёта звёзд
	playerDeaths uint

	enemySpawnTicks uint

	enemyFreezeTicks uint // Оставшиеся тики заморозки врагов бонусом-таймером

	// Длительность уровня в тиках без пауз
	stageTicks uint

	playerCount uint

	stageNumber uint

	isPaused bool
}

func NewStageSessionEntity() *StageSessionEntity {
	playerLives := make([]uint, 2)
	playerInitialLives := make([]uint, 2)
	playerLives[types.PlayerTankNumPlayer1] = defaultStagePlayer1Lives
	playerInitialLives[types.PlayerTankNumPlayer1] = defaultStagePlayer1Lives
	playerLives[types.PlayerTankNumPlayer2] = defaultStagePlayer2Lives
	playerInitialLives[types.PlayerTankNumPlayer2] = defaultStagePlayer2Lives

	return &StageSessionEntity{
		countedDestroyedEnemies: make(map[*types.TankEntity]struct{}),
		recentSpawners:          [recentSpawnersCount]int{noSpawner, noSpawner},
		playerLives:             playerLives,
		playerInitialLives:      playerInitialLives,
		playerCount:             1,
	}
}

// SetUpLevel задаёт сценарий уровня: число врагов — сумма волн
func (s *StageSessionEntity) SetUpLevel(level *types.LevelEntity) {
	s.level = level
	s.totalEnemies = 0
	if level != nil {
		s.totalEnemies = level.GetTotalEnemies()
	}
}

func (s *StageSessionEntity) GetLevel() *types.LevelEntity {
	return s.level
}

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

// TrackDestroyedEnemy учитывает уничтоженного врага ровно один раз:
// повторный вызов для того же танка возвращает false и счётчик не меняет
func (s *StageSessionEntity) TrackDestroyedEnemy(
	tank *types.TankEntity,
) bool {
	if tank == nil {
		return false
	}
	if s.countedDestroyedEnemies == nil {
		s.countedDestroyedEnemies = make(map[*types.TankEntity]struct{})
	}
	if _, exists := s.countedDestroyedEnemies[tank]; exists {
		return false
	}
	s.countedDestroyedEnemies[tank] = struct{}{}
	s.destroyedEnemies++
	return true
}

// ClearDestroyedEnemiesTracking сбрасывает учёт танков,
// не затрагивая счётчик уничтоженных
func (s *StageSessionEntity) ClearDestroyedEnemiesTracking() {
	s.countedDestroyedEnemies = make(map[*types.TankEntity]struct{})
}

func (s *StageSessionEntity) IsPaused() bool {
	return s.isPaused
}

func (s *StageSessionEntity) SetPaused(paused bool) {
	s.isPaused = paused
}

func (s *StageSessionEntity) EnemiesForSpawnCount() uint {
	return s.totalEnemies - s.spawnedEnemies
}

func (s *StageSessionEntity) Reset() {
	s.spawnedEnemies = 0
	s.destroyedEnemies = 0
	s.waveIndex = 0
	s.waveTankIndex = 0
	s.recentSpawners = [recentSpawnersCount]int{noSpawner, noSpawner}
	s.playerDeaths = 0
	s.stageTicks = 0
	s.enemySpawnTicks = 0
	s.isPaused = false
	s.enemyFreezeTicks = 0
	s.ClearDestroyedEnemiesTracking()

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
}

func (s *StageSessionEntity) GetPlayerLives(num types.PlayerTankNum) uint {
	if int(num) >= 0 && int(num) < len(s.playerLives) {
		return s.playerLives[num]
	}
	return 0
}

func (s *StageSessionEntity) GetPlayerInitialLives(
	num types.PlayerTankNum,
) uint {
	if int(num) >= 0 && int(num) < len(s.playerInitialLives) {
		if s.playerInitialLives[num] == 0 {
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

func (s *StageSessionEntity) IsPlayerDefeated(num types.PlayerTankNum) bool {
	return s.GetPlayerLives(num) == 0
}

func (s *StageSessionEntity) SetPlayerLives(
	num types.PlayerTankNum,
	lives uint,
) {
	if int(num) >= 0 && int(num) < len(s.playerLives) {
		s.playerLives[num] = lives
	}
}

// DecrementPlayerLives списывает жизнь за гибель танка игрока
func (s *StageSessionEntity) DecrementPlayerLives(num types.PlayerTankNum) {
	if int(num) >= 0 && int(num) < len(s.playerLives) {
		if s.playerLives[num] == 0 {
			return
		}
		s.playerLives[num]--
		s.playerDeaths++
	}
}

// RestorePlayerLife возвращает жизнь, списанную при неудачном
// респауне: повторная попытка спишет её снова
func (s *StageSessionEntity) RestorePlayerLife(num types.PlayerTankNum) {
	if int(num) >= 0 && int(num) < len(s.playerLives) {
		s.playerLives[num]++
		if s.playerDeaths > 0 {
			s.playerDeaths--
		}
	}
}

func (s *StageSessionEntity) GetPlayerDeaths() uint {
	return s.playerDeaths
}

// RegisterEnemySpawned учитывает спаун врага из текущей позиции
// курсора волн и запускает паузу до следующего
func (s *StageSessionEntity) RegisterEnemySpawned(
	spawnerIndex int,
	delayTicks uint,
) {
	s.IncrementSpawnedEnemies()
	s.enemySpawnTicks = delayTicks
	for i := recentSpawnersCount - 1; i > 0; i-- {
		s.recentSpawners[i] = s.recentSpawners[i-1]
	}
	s.recentSpawners[0] = spawnerIndex
}

// GetRecentSpawners — недавние спаунеры, последний — первым;
// пустые слоты равны -1
func (s *StageSessionEntity) GetRecentSpawners() [recentSpawnersCount]int {
	return s.recentSpawners
}

// GetWaveCursor — индекс волны и танка в ней для следующего спауна
func (s *StageSessionEntity) GetWaveCursor() (int, int) {
	return s.waveIndex, s.waveTankIndex
}

// SetWaveCursor переводит курсор волн на следующий танк
func (s *StageSessionEntity) SetWaveCursor(waveIndex, tankIndex int) {
	s.waveIndex = waveIndex
	s.waveTankIndex = tankIndex
}

func (s *StageSessionEntity) GetSpawnedEnemies() uint {
	return s.spawnedEnemies
}

func (s *StageSessionEntity) GetDestroyedEnemies() uint {
	return s.destroyedEnemies
}

// UpdateStageTicks продвигает таймер уровня на тик
func (s *StageSessionEntity) UpdateStageTicks() {
	s.stageTicks++
}

func (s *StageSessionEntity) GetStageTicks() uint {
	return s.stageTicks
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

// FreezeEnemies запускает заморозку врагов на заданное число тиков
func (s *StageSessionEntity) FreezeEnemies(ticks uint) {
	s.enemyFreezeTicks = ticks
}

func (s *StageSessionEntity) AreEnemiesFrozen() bool {
	return s.enemyFreezeTicks > 0
}

func (s *StageSessionEntity) UpdateEnemyFreezeCountdown() {
	if s.enemyFreezeTicks > 0 {
		s.enemyFreezeTicks--
	}
}

// GetMaxActiveEnemies — лимит одновременно активных врагов уровня
func (s *StageSessionEntity) GetMaxActiveEnemies() uint {
	if s.level == nil {
		return 0
	}
	return s.level.GetMaxActive()
}

func (s *StageSessionEntity) GetPlayerCount() uint {
	return s.playerCount
}

func (s *StageSessionEntity) GetStageNumber() uint {
	return s.stageNumber
}

func (s *StageSessionEntity) SetStageNumber(number uint) {
	s.stageNumber = number
}

func (s *StageSessionEntity) SetPlayerCount(count uint) {
	if count < 1 {
		count = 1
	}
	if count > 2 {
		count = 2
	}
	s.playerCount = count

	if count == 1 {
		s.playerLives[types.PlayerTankNumPlayer2] = 0
		s.playerInitialLives[types.PlayerTankNumPlayer2] = 0
	} else {
		if s.playerInitialLives[types.PlayerTankNumPlayer2] == 0 {
			s.playerLives[types.PlayerTankNumPlayer2] = defaultStagePlayer2Lives
			s.playerInitialLives[types.PlayerTankNumPlayer2] = defaultStagePlayer2Lives
		}
	}
}
