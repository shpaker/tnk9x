package game_test

import (
	"testing"

	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/types"
)

func TestNewGameRepositoriesRegistry(t *testing.T) {
	gameRepo := game.NewGameRepositoriesRegistry()

	// Проверяем, что все репозитории созданы
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

	// Создаем тестовый танк
	tank := &types.TankEntity{
		Position: types.Position{X: 0, Y: 0},
	}

	// Создаем пулю с минимальными данными
	bullet := types.NewBulletEntity(
		types.Position{X: 10, Y: 20},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		nil,
		100.0,
		types.DirectionUp,
		tank,
	)

	// Добавляем пулю
	err := gameRepo.GetBulletsRepository().AddBullet(bullet)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Проверяем, что пуля добавлена
	bullets := gameRepo.GetBulletsRepository().GetAllBullets()
	if len(bullets) != 1 {
		t.Errorf("Expected 1 bullet, got %d", len(bullets))
	}

	// Проверяем методы интерфейса
	bulletsRepo := gameRepo.GetBulletsRepository()
	if bulletsRepo == nil {
		t.Error("BulletsRepository() should not return nil")
	}
}

func TestGameRepositoriesRegistryAnimations(t *testing.T) {
	gameRepo := game.NewGameRepositoriesRegistry()

	// Проверяем методы интерфейса
	animationsRepo := gameRepo.GetAnimationsRepository()
	if animationsRepo == nil {
		t.Error("AnimationsRepository() should not return nil")
	}
}

func TestGameRepositoriesRegistryTanks(t *testing.T) {
	gameRepo := game.NewGameRepositoriesRegistry()

	// Создаем танк игрока
	playerTank := &types.TankEntity{
		Position:  types.Position{X: 100, Y: 200},
		Direction: types.DirectionUp,
		Speed:     0,
	}

	// Добавляем танк игрока
	gameRepo.GetTanksRepository().
		SetPlayer(types.PlayerTankNumPlayer1, playerTank)

	// Проверяем, что танк добавлен
	player := gameRepo.GetTanksRepository().
		GetPlayer(types.PlayerTankNumPlayer1)
	if player == nil {
		t.Error("Player tank should not be nil")
	}

	// Создаем врага
	enemyTank := &types.TankEntity{
		Position:  types.Position{X: 300, Y: 400},
		Direction: types.DirectionDown,
		Speed:     0,
	}

	// Добавляем врага
	gameRepo.GetTanksRepository().AddEnemy(enemyTank)

	// Проверяем получение всех врагов
	enemies := gameRepo.GetTanksRepository().GetAllEnemies()
	if len(enemies) != 1 {
		t.Errorf("Expected 1 enemy, got %d", len(enemies))
	}

	// Проверяем методы интерфейса
	tanksRepo := gameRepo.GetTanksRepository()
	if tanksRepo == nil {
		t.Error("TanksRepository() should not return nil")
	}
}
