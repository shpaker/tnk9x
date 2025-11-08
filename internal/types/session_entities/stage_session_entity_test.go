package session_entities

import "testing"

func TestStageSessionEntity_IsCompleted(t *testing.T) {
	session := &StageSessionEntity{
		totalEnemies:     2,
		spawnedEnemies:   2,
		destroyedEnemies: 2,
	}

	if !session.IsCompleted() {
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
