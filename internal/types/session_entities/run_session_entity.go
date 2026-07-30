package session_entities

import (
	"github.com/shpaker/tnk9x/internal/types"
)

const (
	defaultRunPlayerLives = 3

	// runPlayersCount — максимум игроков в забеге
	runPlayersCount = 2

	// enemyTiersCount — число типов вражеских танков (уровни 0-3)
	enemyTiersCount = 4

	// enemyKillBaseScore — очки за врага уровня 0; каждый следующий
	// уровень добавляет столько же (100/200/300/400 как в NES)
	enemyKillBaseScore = 100

	// bonusPickupScore — очки за подбор любого бонуса
	bonusPickupScore = 500

	// extraLifeScoreThreshold — порог очков для разовой доп. жизни
	extraLifeScoreThreshold = 20000
)

// RunSessionEntity — данные одного прохождения: живут между этапами
// и сбрасываются только при старте новой игры с титульного экрана.
// HI-SCORE живёт всё время работы приложения.
type RunSessionEntity struct {
	playerCount uint
	stage       uint

	playerLives    [runPlayersCount]uint
	starLevels     [runPlayersCount]uint
	score          [runPlayersCount]uint
	stageKills     [runPlayersCount][enemyTiersCount]uint
	extraLifeGiven [runPlayersCount]bool

	hiScore uint
}

func NewRunSessionEntity() *RunSessionEntity {
	run := &RunSessionEntity{}
	run.ResetRun(1, 1)
	return run
}

// ResetRun начинает новый забег: жизни, очки, звёзды и учёт этапа
// сбрасываются; HI-SCORE сохраняется
func (r *RunSessionEntity) ResetRun(playerCount, stage uint) {
	r.SetPlayerCount(playerCount)
	if stage == 0 {
		stage = 1
	}
	r.stage = stage

	r.playerLives[types.PlayerTankNumPlayer1] = defaultRunPlayerLives
	if r.playerCount > 1 {
		r.playerLives[types.PlayerTankNumPlayer2] = defaultRunPlayerLives
	} else {
		r.playerLives[types.PlayerTankNumPlayer2] = 0
	}

	for i := 0; i < runPlayersCount; i++ {
		r.starLevels[i] = 0
		r.score[i] = 0
		r.extraLifeGiven[i] = false
	}
	r.ResetStageTallies()
}

// ResetStageTallies обнуляет пер-этапный подсчёт убийств
// (для экрана итогов); очки и жизни не трогает
func (r *RunSessionEntity) ResetStageTallies() {
	for i := 0; i < runPlayersCount; i++ {
		for tier := 0; tier < enemyTiersCount; tier++ {
			r.stageKills[i][tier] = 0
		}
	}
}

func (r *RunSessionEntity) GetStage() uint {
	return r.stage
}

func (r *RunSessionEntity) SetStage(stage uint) {
	if stage == 0 {
		stage = 1
	}
	r.stage = stage
}

func (r *RunSessionEntity) GetPlayerCount() uint {
	return r.playerCount
}

func (r *RunSessionEntity) SetPlayerCount(count uint) {
	if count < 1 {
		count = 1
	}
	if count > runPlayersCount {
		count = runPlayersCount
	}
	r.playerCount = count

	if count == 1 {
		r.playerLives[types.PlayerTankNumPlayer2] = 0
	} else if r.playerLives[types.PlayerTankNumPlayer2] == 0 {
		r.playerLives[types.PlayerTankNumPlayer2] = defaultRunPlayerLives
	}
}

func (r *RunSessionEntity) isValidPlayer(num types.PlayerTankNum) bool {
	return int(num) >= 0 && int(num) < runPlayersCount
}

func (r *RunSessionEntity) GetPlayerLives(num types.PlayerTankNum) uint {
	if !r.isValidPlayer(num) {
		return 0
	}
	return r.playerLives[num]
}

func (r *RunSessionEntity) SetPlayerLives(
	num types.PlayerTankNum,
	lives uint,
) {
	if r.isValidPlayer(num) {
		r.playerLives[num] = lives
	}
}

func (r *RunSessionEntity) DecrementPlayerLives(num types.PlayerTankNum) {
	if r.isValidPlayer(num) && r.playerLives[num] > 0 {
		r.playerLives[num]--
	}
}

func (r *RunSessionEntity) IsPlayerDefeated(num types.PlayerTankNum) bool {
	return r.GetPlayerLives(num) == 0
}

func (r *RunSessionEntity) GetPlayerInitialLives(
	num types.PlayerTankNum,
) uint {
	if !r.isValidPlayer(num) {
		return 0
	}
	if num == types.PlayerTankNumPlayer2 && r.playerCount < 2 {
		return 0
	}
	return defaultRunPlayerLives
}

func (r *RunSessionEntity) GetStarLevel(num types.PlayerTankNum) uint {
	if !r.isValidPlayer(num) {
		return 0
	}
	return r.starLevels[num]
}

func (r *RunSessionEntity) SetStarLevel(
	num types.PlayerTankNum,
	level uint,
) {
	if r.isValidPlayer(num) {
		r.starLevels[num] = level
	}
}

// AddEnemyKill начисляет очки за уничтоженного врага указанного уровня
// и учитывает его в пер-этапном подсчёте; возвращает true, если
// начисление дало дополнительную жизнь
func (r *RunSessionEntity) AddEnemyKill(
	num types.PlayerTankNum,
	tier uint,
) bool {
	if !r.isValidPlayer(num) {
		return false
	}
	if tier >= enemyTiersCount {
		tier = enemyTiersCount - 1
	}
	r.stageKills[num][tier]++
	return r.addScore(num, uint(enemyKillBaseScore)*(tier+1))
}

// AddBonusPoints начисляет очки за подбор бонуса; возвращает true,
// если начисление дало дополнительную жизнь
func (r *RunSessionEntity) AddBonusPoints(num types.PlayerTankNum) bool {
	if !r.isValidPlayer(num) {
		return false
	}
	return r.addScore(num, bonusPickupScore)
}

func (r *RunSessionEntity) addScore(
	num types.PlayerTankNum,
	points uint,
) bool {
	r.score[num] += points
	if r.score[num] > r.hiScore {
		r.hiScore = r.score[num]
	}
	if !r.extraLifeGiven[num] && r.score[num] >= extraLifeScoreThreshold {
		r.extraLifeGiven[num] = true
		r.playerLives[num]++
		return true
	}
	return false
}

func (r *RunSessionEntity) GetScore(num types.PlayerTankNum) uint {
	if !r.isValidPlayer(num) {
		return 0
	}
	return r.score[num]
}

// GetStageKills возвращает число убитых на текущем этапе врагов
// по уровням 0-3
func (r *RunSessionEntity) GetStageKills(
	num types.PlayerTankNum,
) [enemyTiersCount]uint {
	if !r.isValidPlayer(num) {
		return [enemyTiersCount]uint{}
	}
	return r.stageKills[num]
}

func (r *RunSessionEntity) GetHiScore() uint {
	return r.hiScore
}
