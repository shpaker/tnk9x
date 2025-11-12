package session_entities

import (
	"testing"

	"github.com/shpaker/gonflict/internal/types"
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
