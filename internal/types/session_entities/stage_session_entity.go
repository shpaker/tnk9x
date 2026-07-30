package session_entities

import (
	"github.com/shpaker/tnk9x/internal/types"
)

const (
	defaultStageRespawnDelay = 3 * 60

	// defaultMaxActiveEnemies — канонический лимит NES: не более 4
	// врагов на экране одновременно
	defaultMaxActiveEnemies = 4
)

// StageSessionEntity — состояние одного этапа: счётчики спавна и
// уничтожения врагов, пауза. Данные забега (жизни, очки, звёзды,
// номер этапа) делегируются RunSessionEntity и переживают этап.
type StageSessionEntity struct {
	run *RunSessionEntity

	totalEnemies     uint
	spawnedEnemies   uint
	destroyedEnemies uint

	// enemyQueue — уровни врагов волны этапа в порядке спавна
	enemyQueue []uint

	// Танки, уже учтённые счётчиком destroyedEnemies
	countedDestroyedEnemies map[*types.TankEntity]struct{}

	enemyRespawnDelay uint
	enemySpawnTicks   uint

	maxActiveEnemies uint

	// enemyFreezeTicks — оставшееся время заморозки врагов (бонус «таймер»)
	enemyFreezeTicks uint

	// shovelTicks — оставшееся время стального кольца вокруг штаба
	shovelTicks uint

	// stageTicks — время с начала этапа; определяет фазу AI врагов
	stageTicks uint

	isPaused bool
}

func NewStageSessionEntity(run *RunSessionEntity) *StageSessionEntity {
	if run == nil {
		run = NewRunSessionEntity()
	}

	return &StageSessionEntity{
		run:                     run,
		totalEnemies:            20,
		destroyedEnemies:        0,
		countedDestroyedEnemies: make(map[*types.TankEntity]struct{}),
		enemyRespawnDelay:       defaultStageRespawnDelay,
		enemySpawnTicks:         defaultStageRespawnDelay,
		maxActiveEnemies:        defaultMaxActiveEnemies,
	}
}

func (s *StageSessionEntity) RunSession() *RunSessionEntity {
	return s.run
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

// SetEnemyQueue задаёт состав волны этапа;
// размер волны определяет totalEnemies
func (s *StageSessionEntity) SetEnemyQueue(tiers []uint) {
	s.enemyQueue = append([]uint(nil), tiers...)
	if len(s.enemyQueue) > 0 {
		s.totalEnemies = uint(len(s.enemyQueue))
	}
}

// NextEnemyTier — уровень следующего врага по очереди волны
func (s *StageSessionEntity) NextEnemyTier() uint {
	if int(s.spawnedEnemies) < len(s.enemyQueue) {
		return s.enemyQueue[s.spawnedEnemies]
	}
	return 0
}

// NextEnemySpawnIndex — порядковый номер следующего спавна (0-based);
// точки спавна перебираются по нему циклически
func (s *StageSessionEntity) NextEnemySpawnIndex() uint {
	return s.spawnedEnemies
}

// ResetStage готовит сессию к новому этапу: счётчики спавна, пауза и
// пер-этапные итоги очищаются; жизни и очки забега не трогаются
func (s *StageSessionEntity) ResetStage() {
	s.spawnedEnemies = 0
	s.destroyedEnemies = 0
	s.isPaused = false
	s.enemyFreezeTicks = 0
	s.shovelTicks = 0
	s.stageTicks = 0
	s.ClearDestroyedEnemiesTracking()
	s.run.ResetStageTallies()
	s.ResetEnemySpawnCountdown()
}

func (s *StageSessionEntity) SetEnemyFreezeTicks(ticks uint) {
	s.enemyFreezeTicks = ticks
}

func (s *StageSessionEntity) AreEnemiesFrozen() bool {
	return s.enemyFreezeTicks > 0
}

func (s *StageSessionEntity) TickEnemyFreeze() {
	if s.enemyFreezeTicks > 0 {
		s.enemyFreezeTicks--
	}
}

func (s *StageSessionEntity) IncrementStageTicks() {
	s.stageTicks++
}

func (s *StageSessionEntity) GetStageTicks() uint {
	return s.stageTicks
}

func (s *StageSessionEntity) SetShovelTicks(ticks uint) {
	s.shovelTicks = ticks
}

func (s *StageSessionEntity) GetShovelTicks() uint {
	return s.shovelTicks
}

func (s *StageSessionEntity) GetPlayerLives(num types.PlayerTankNum) uint {
	return s.run.GetPlayerLives(num)
}

func (s *StageSessionEntity) GetPlayerInitialLives(
	num types.PlayerTankNum,
) uint {
	return s.run.GetPlayerInitialLives(num)
}

func (s *StageSessionEntity) IsPlayerDefeated(num types.PlayerTankNum) bool {
	return s.run.IsPlayerDefeated(num)
}

func (s *StageSessionEntity) SetPlayerLives(
	num types.PlayerTankNum,
	lives uint,
) {
	s.run.SetPlayerLives(num, lives)
}

func (s *StageSessionEntity) DecrementPlayerLives(num types.PlayerTankNum) {
	s.run.DecrementPlayerLives(num)
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

func (s *StageSessionEntity) GetMaxActiveEnemies() uint {
	return s.maxActiveEnemies
}

func (s *StageSessionEntity) SetMaxActiveEnemies(value uint) {
	if value < 1 {
		value = defaultMaxActiveEnemies
	}
	if value > 10 {
		value = 10
	}
	s.maxActiveEnemies = value
}

func (s *StageSessionEntity) GetPlayerCount() uint {
	return s.run.GetPlayerCount()
}

func (s *StageSessionEntity) GetStageNumber() uint {
	return s.run.GetStage()
}

func (s *StageSessionEntity) SetStageNumber(number uint) {
	s.run.SetStage(number)
}

func (s *StageSessionEntity) SetPlayerCount(count uint) {
	s.run.SetPlayerCount(count)
}
