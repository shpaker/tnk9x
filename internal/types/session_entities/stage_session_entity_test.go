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
	session := NewStageSessionEntity(nil)

	if session.GetPlayerLives(
		types.PlayerTankNumPlayer1,
	) != defaultRunPlayerLives {
		t.Fatalf(
			"expected default player lives to be %d, got %d",
			defaultRunPlayerLives,
			session.GetPlayerLives(types.PlayerTankNumPlayer1),
		)
	}

	if session.IsPlayerDefeated(types.PlayerTankNumPlayer1) {
		t.Fatalf("expected player to have lives remaining by default")
	}

	if session.GetPlayerInitialLives(
		types.PlayerTankNumPlayer1,
	) != defaultRunPlayerLives {
		t.Fatalf(
			"expected initial player lives to be %d, got %d",
			defaultRunPlayerLives,
			session.GetPlayerInitialLives(types.PlayerTankNumPlayer1),
		)
	}
}

func TestStageSessionEntity_Player1Lives_Decrement(t *testing.T) {
	session := NewStageSessionEntity(nil)
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
	session := NewStageSessionEntity(nil)
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

// Пауза хранится в сессии; ResetStage снимает её
func TestStageSessionEntity_Pause(t *testing.T) {
	session := NewStageSessionEntity(nil)

	if session.IsPaused() {
		t.Error("новая сессия не должна быть на паузе")
	}

	session.SetPaused(true)
	if !session.IsPaused() {
		t.Error("SetPaused(true) не включил паузу")
	}

	session.ResetStage()
	if session.IsPaused() {
		t.Error("ResetStage должен снимать паузу")
	}
}

// ResetStage готовит следующий этап, не трогая жизни забега
func TestStageSessionEntity_ResetStage_KeepsLives(t *testing.T) {
	run := NewRunSessionEntity()
	session := NewStageSessionEntity(run)

	session.DecrementPlayerLives(types.PlayerTankNumPlayer1)
	session.IncrementSpawnedEnemies()
	session.IncrementDestroyedEnemies()

	session.ResetStage()

	if got := session.GetPlayerLives(types.PlayerTankNumPlayer1); got != defaultRunPlayerLives-1 {
		t.Errorf(
			"жизни после ResetStage %d, ожидалось %d",
			got,
			defaultRunPlayerLives-1,
		)
	}
	if session.EnemiesForSpawnCount() != session.GetTotalEnemies() {
		t.Error("счётчик спавна не сброшен")
	}
	if session.GetRemainingEnemies() != session.GetTotalEnemies() {
		t.Error("счётчик уничтоженных не сброшен")
	}
}

// Каждый танк учитывается счётчиком уничтоженных ровно один раз
func TestStageSessionEntity_TrackDestroyedEnemy(t *testing.T) {
	session := NewStageSessionEntity(nil)
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

	// ResetStage обнуляет и счётчик, и трекинг
	session.ResetStage()
	if got := session.GetTotalEnemies() - session.GetRemainingEnemies(); got != 0 {
		t.Errorf("после ResetStage уничтожено %d, ожидалось 0", got)
	}
}
