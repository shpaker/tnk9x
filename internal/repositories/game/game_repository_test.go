package game_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/types"
)

func TestNewGameRepositoriesRegistry(t *testing.T) {
	gameRepo := game.NewGameRepositoriesRegistry()

	if gameRepo.GetBulletsRepository() == nil {
		t.Error("Bullets repository should not be nil")
	}
	if gameRepo.GetAnimationsRepository() == nil {
		t.Error("Animations repository should not be nil")
	}
	if gameRepo.GetTanksRepository() == nil {
		t.Error("Tanks repository should not be nil")
	}
}

func TestGameRepositoriesRegistryBullets(t *testing.T) {
	gameRepo := game.NewGameRepositoriesRegistry()

	tank := types.NewDefaultTankEntity(types.TankRolePlayer1, types.DirectionUp)
	tank.Position = types.Position{X: 0, Y: 0}
	tankPtr := &tank

	specs := types.NewSpecsEntity(
		0, // Уровень 0
		32.0,
		false, // Пули не усиленные
		100.0, // Скорость пули
		1,     // Лимит пуль: 1
	)
	bullet := types.NewBulletEntity(
		types.Position{X: 10, Y: 20},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		nil,
		types.DirectionUp,
		specs,
		tankPtr,
	)

	err := gameRepo.GetBulletsRepository().AddBullet(bullet)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	bullets := gameRepo.GetBulletsRepository().GetAllBullets()
	if len(bullets) != 1 {
		t.Errorf("Expected 1 bullet, got %d", len(bullets))
	}

	bulletsRepo := gameRepo.GetBulletsRepository()
	if bulletsRepo == nil {
		t.Error("BulletsRepository() should not return nil")
	}
}

func TestGameRepositoriesRegistryAnimations(t *testing.T) {
	gameRepo := game.NewGameRepositoriesRegistry()

	animationsRepo := gameRepo.GetAnimationsRepository()
	if animationsRepo == nil {
		t.Error("AnimationsRepository() should not return nil")
	}
}

func TestGameRepositoriesRegistryTanks(t *testing.T) {
	gameRepo := game.NewGameRepositoriesRegistry()

	playerTank := &types.TankEntity{
		Position:  types.Position{X: 100, Y: 200},
		Direction: types.DirectionUp,
	}

	gameRepo.GetTanksRepository().
		SetPlayer(types.PlayerTankNumPlayer1, playerTank)

	player := gameRepo.GetTanksRepository().
		GetPlayer(types.PlayerTankNumPlayer1)
	if player == nil {
		t.Error("Player tank should not be nil")
	}

	enemyTank := &types.TankEntity{
		Position:  types.Position{X: 300, Y: 400},
		Direction: types.DirectionDown,
	}

	gameRepo.GetTanksRepository().AddEnemy(enemyTank)

	enemies := gameRepo.GetTanksRepository().GetAllEnemies()
	if len(enemies) != 1 {
		t.Errorf("Expected 1 enemy, got %d", len(enemies))
	}

	tanksRepo := gameRepo.GetTanksRepository()
	if tanksRepo == nil {
		t.Error("TanksRepository() should not return nil")
	}
}
