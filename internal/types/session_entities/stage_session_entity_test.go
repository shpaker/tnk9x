package session_entities

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

func TestStageSessionEntity_AreAllEnemiesDefeated(t *testing.T) {
	session := &StageSessionEntity{
		totalEnemies:     2,
		spawnedEnemies:   2,
		destroyedEnemies: 2,
	}

	if !session.AreAllEnemiesDefeated() {
		t.Fatalf(
			"expected stage to be completed when destroyedEnemies >= totalEnemies",
		)
	}
}

func TestStageSessionEntity_GetNextEnemyNumber(t *testing.T) {
	session := &StageSessionEntity{
		totalEnemies:   5,
		spawnedEnemies: 2,
	}

	next := session.GetNextEnemyNumber()

	if next != 3 {
		t.Fatalf("expected next enemy number to be 3, got %d", next)
	}
}

func TestStageSessionEntity_IncrementSpawnedEnemies(t *testing.T) {
	session := &StageSessionEntity{
		totalEnemies:   10,
		spawnedEnemies: 3,
	}

	session.IncrementSpawnedEnemies()

	if session.spawnedEnemies != 4 {
		t.Fatalf(
			"expected spawnedEnemies to increment to 4, got %d",
			session.spawnedEnemies,
		)
	}
}

func TestStageSessionEntity_IncrementDestroyedEnemies(t *testing.T) {
	session := &StageSessionEntity{
		totalEnemies:     10,
		destroyedEnemies: 6,
	}

	session.IncrementDestroyedEnemies()

	if session.destroyedEnemies != 7 {
		t.Fatalf(
			"expected destroyedEnemies to increment to 7, got %d",
			session.destroyedEnemies,
		)
	}
}

func TestStageSessionEntity_Player1Lives_Default(t *testing.T) {
	session := NewStageSessionEntity()

	if session.GetPlayerLives(
		types.PlayerTankNumPlayer1,
	) != defaultStagePlayer1Lives {
		t.Fatalf(
			"expected default player lives to be %d, got %d",
			defaultStagePlayer1Lives,
			session.GetPlayerLives(types.PlayerTankNumPlayer1),
		)
	}

	if session.IsPlayerDefeated(types.PlayerTankNumPlayer1) {
		t.Fatalf("expected player to have lives remaining by default")
	}

	if session.GetPlayerInitialLives(
		types.PlayerTankNumPlayer1,
	) != defaultStagePlayer1Lives {
		t.Fatalf(
			"expected initial player lives to be %d, got %d",
			defaultStagePlayer1Lives,
			session.GetPlayerInitialLives(types.PlayerTankNumPlayer1),
		)
	}
}

func TestStageSessionEntity_Player1Lives_Decrement(t *testing.T) {
	session := NewStageSessionEntity()
	session.SetPlayerLives(types.PlayerTankNumPlayer1, 1)
	session.DecrementPlayerLives(types.PlayerTankNumPlayer1)
	session.DecrementPlayerLives(types.PlayerTankNumPlayer1)

	if session.GetPlayerLives(types.PlayerTankNumPlayer1) != 0 {
		t.Fatalf(
			"expected player lives to stop at 0, got %d",
			session.GetPlayerLives(types.PlayerTankNumPlayer1),
		)
	}

	if !session.IsPlayerDefeated(types.PlayerTankNumPlayer1) {
		t.Fatalf("expected player to be defeated with zero lives")
	}
}

func TestStageSessionEntity_RemainingEnemies(t *testing.T) {
	session := NewStageSessionEntity()
	session.totalEnemies = 2

	if session.GetRemainingEnemies() != 2 {
		t.Fatalf(
			"expected remaining enemies to be 2, got %d",
			session.GetRemainingEnemies(),
		)
	}

	session.IncrementDestroyedEnemies()
	if session.GetRemainingEnemies() != 1 {
		t.Fatalf(
			"expected remaining enemies to be 1, got %d",
			session.GetRemainingEnemies(),
		)
	}

	session.IncrementDestroyedEnemies()
	if session.GetRemainingEnemies() != 0 {
		t.Fatalf(
			"expected remaining enemies to be 0, got %d",
			session.GetRemainingEnemies(),
		)
	}
}

// Пауза хранится в сессии; Reset снимает её
func TestStageSessionEntity_Pause(t *testing.T) {
	session := NewStageSessionEntity()

	if session.IsPaused() {
		t.Error("новая сессия не должна быть на паузе")
	}

	session.SetPaused(true)
	if !session.IsPaused() {
		t.Error("SetPaused(true) не включил паузу")
	}

	session.Reset()
	if session.IsPaused() {
		t.Error("Reset должен снимать паузу")
	}
}

// Каждый танк учитывается счётчиком уничтоженных ровно один раз
func TestStageSessionEntity_TrackDestroyedEnemy(t *testing.T) {
	session := NewStageSessionEntity()
	session.totalEnemies = 20
	tankValue := types.NewDefaultTankEntity(
		types.TankRoleEnemy,
		types.DirectionUp,
	)
	tank := &tankValue

	if !session.TrackDestroyedEnemy(tank) {
		t.Error("первый учёт танка должен вернуть true")
	}
	if session.TrackDestroyedEnemy(tank) {
		t.Error("повторный учёт того же танка должен вернуть false")
	}
	if got := session.GetTotalEnemies() - session.GetRemainingEnemies(); got != 1 {
		t.Errorf("уничтожено %d, ожидался 1", got)
	}
	if session.TrackDestroyedEnemy(nil) {
		t.Error("nil-танк не учитывается")
	}

	// Сброс трекинга не трогает счётчик, но позволяет учесть танк заново
	session.ClearDestroyedEnemiesTracking()
	if got := session.GetTotalEnemies() - session.GetRemainingEnemies(); got != 1 {
		t.Errorf("после сброса трекинга уничтожено %d, ожидался 1", got)
	}
	if !session.TrackDestroyedEnemy(tank) {
		t.Error("после сброса трекинга танк учитывается заново")
	}

	// Полный Reset обнуляет и счётчик, и трекинг
	session.Reset()
	if got := session.GetTotalEnemies() - session.GetRemainingEnemies(); got != 0 {
		t.Errorf("после Reset уничтожено %d, ожидалось 0", got)
	}
}

// The enemy total comes from the level waves
func TestStageSessionEntity_SetUpLevel(t *testing.T) {
	session := NewStageSessionEntity()
	level := types.NewLevelEntity(1, "TEST", 4, 0, []types.WaveSpec{
		{Tanks: make([]types.WaveTank, 3)},
		{Tanks: make([]types.WaveTank, 2)},
	}, false, nil)

	session.SetUpLevel(level)

	if got := session.GetTotalEnemies(); got != 5 {
		t.Errorf("total enemies %d, want 5", got)
	}
	if got := session.GetMaxActiveEnemies(); got != 4 {
		t.Errorf("max active %d, want 4", got)
	}
}

// Deaths count toward lost lives; a failed respawn gives the life back
func TestStageSessionEntity_PlayerDeaths(t *testing.T) {
	session := NewStageSessionEntity()

	session.DecrementPlayerLives(types.PlayerTankNumPlayer1)
	session.DecrementPlayerLives(types.PlayerTankNumPlayer1)
	session.RestorePlayerLife(types.PlayerTankNumPlayer1)

	if got := session.GetPlayerDeaths(); got != 1 {
		t.Errorf("deaths %d, want 1", got)
	}
	if got := session.GetPlayerLives(types.PlayerTankNumPlayer1); got != 2 {
		t.Errorf("lives %d, want 2", got)
	}

	session.Reset()
	if session.GetPlayerDeaths() != 0 || session.GetStageTicks() != 0 {
		t.Error("Reset must clear deaths and stage time")
	}
}

// The spawner history keeps the last two spawners, newest first
func TestStageSessionEntity_RecentSpawners(t *testing.T) {
	session := NewStageSessionEntity()

	session.RegisterEnemySpawned(2, 90)
	session.RegisterEnemySpawned(0, 90)

	if got := session.GetRecentSpawners(); got != [2]int{0, 2} {
		t.Errorf("recent spawners %v, want [0 2]", got)
	}
	if session.CanSpawnNextEnemy() {
		t.Error("the spawn pause must block the next spawn")
	}
}
