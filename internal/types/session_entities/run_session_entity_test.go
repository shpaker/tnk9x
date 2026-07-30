package session_entities

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

// Новый забег: жизни по умолчанию, второй игрок без жизней в 1P
func TestRunSessionEntity_ResetRun(t *testing.T) {
	run := NewRunSessionEntity()

	if got := run.GetPlayerLives(types.PlayerTankNumPlayer1); got != defaultRunPlayerLives {
		t.Errorf("жизни P1 %d, ожидалось %d", got, defaultRunPlayerLives)
	}
	if got := run.GetPlayerLives(types.PlayerTankNumPlayer2); got != 0 {
		t.Errorf("жизни P2 в 1P %d, ожидалось 0", got)
	}
	if run.GetStage() != 1 {
		t.Errorf("этап %d, ожидался 1", run.GetStage())
	}

	run.ResetRun(2, 5)
	if got := run.GetPlayerLives(types.PlayerTankNumPlayer2); got != defaultRunPlayerLives {
		t.Errorf("жизни P2 в 2P %d, ожидалось %d", got, defaultRunPlayerLives)
	}
	if run.GetStage() != 5 {
		t.Errorf("этап %d, ожидался 5", run.GetStage())
	}
}

// ResetRun сбрасывает очки и звёзды, но сохраняет HI-SCORE
func TestRunSessionEntity_ResetRun_KeepsHiScore(t *testing.T) {
	run := NewRunSessionEntity()

	run.AddEnemyKill(types.PlayerTankNumPlayer1, 3)
	run.SetStarLevel(types.PlayerTankNumPlayer1, 2)
	hiScore := run.GetHiScore()
	if hiScore == 0 {
		t.Fatal("HI-SCORE не обновился после начисления")
	}

	run.ResetRun(1, 1)
	if run.GetScore(types.PlayerTankNumPlayer1) != 0 {
		t.Error("очки не сброшены новым забегом")
	}
	if run.GetStarLevel(types.PlayerTankNumPlayer1) != 0 {
		t.Error("звёзды не сброшены новым забегом")
	}
	if run.GetHiScore() != hiScore {
		t.Error("HI-SCORE потерян при новом забеге")
	}
}

// Очки за врагов: 100/200/300/400 по уровням, учёт в пер-этапном подсчёте
func TestRunSessionEntity_AddEnemyKill(t *testing.T) {
	run := NewRunSessionEntity()

	for tier := uint(0); tier < enemyTiersCount; tier++ {
		run.AddEnemyKill(types.PlayerTankNumPlayer1, tier)
	}

	wantScore := uint(100 + 200 + 300 + 400)
	if got := run.GetScore(types.PlayerTankNumPlayer1); got != wantScore {
		t.Errorf("счёт %d, ожидалось %d", got, wantScore)
	}

	kills := run.GetStageKills(types.PlayerTankNumPlayer1)
	for tier, count := range kills {
		if count != 1 {
			t.Errorf("убийств уровня %d: %d, ожидалось 1", tier, count)
		}
	}

	run.ResetStageTallies()
	kills = run.GetStageKills(types.PlayerTankNumPlayer1)
	for tier, count := range kills {
		if count != 0 {
			t.Errorf(
				"после сброса итогов убийств уровня %d: %d",
				tier,
				count,
			)
		}
	}
	if run.GetScore(types.PlayerTankNumPlayer1) != wantScore {
		t.Error("сброс пер-этапных итогов не должен трогать счёт")
	}
}

// Дополнительная жизнь за 20000 очков выдаётся ровно один раз
func TestRunSessionEntity_ExtraLifeAtThreshold(t *testing.T) {
	run := NewRunSessionEntity()
	num := types.PlayerTankNumPlayer1

	granted := 0
	// 50 тяжёлых танков по 400 = 20000, дальше ещё 10 сверх порога
	for i := 0; i < 60; i++ {
		if run.AddEnemyKill(num, 3) {
			granted++
		}
	}

	if granted != 1 {
		t.Errorf("доп. жизнь выдана %d раз, ожидался 1", granted)
	}
	if got := run.GetPlayerLives(num); got != defaultRunPlayerLives+1 {
		t.Errorf(
			"жизней %d, ожидалось %d",
			got,
			defaultRunPlayerLives+1,
		)
	}
}

// Бонус приносит 500 очков
func TestRunSessionEntity_AddBonusPoints(t *testing.T) {
	run := NewRunSessionEntity()

	run.AddBonusPoints(types.PlayerTankNumPlayer1)
	if got := run.GetScore(types.PlayerTankNumPlayer1); got != bonusPickupScore {
		t.Errorf("счёт %d, ожидалось %d", got, bonusPickupScore)
	}
}
