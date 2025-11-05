package game_test

import (
	"testing"

	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/types"
)

func TestNewGameRepositoriesRegistry(t *testing.T) {
	gameRepo := game.NewGameRepositoriesRegistry()

	// Проверяем, что все репозитории созданы
	if gameRepo.BulletsRepository() == nil {
		t.Error("Bullets repository should not be nil")
	}
	if gameRepo.AnimationsRepository() == nil {
		t.Error("Animations repository should not be nil")
	}
	if gameRepo.TanksRepository() == nil {
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
	bullet := types.BulletEntity{
		Position:  types.Position{X: 10, Y: 20},
		Direction: types.DirectionUp,
		Speed:     100,
		Altitude:  types.SURFACE,
		Owner:     tank,
	}

	// Добавляем пулю
	err := gameRepo.BulletsRepository().AddBullet(bullet)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Проверяем, что пуля добавлена
	bullets := gameRepo.BulletsRepository().GetAllBullets()
	if len(bullets) != 1 {
		t.Errorf("Expected 1 bullet, got %d", len(bullets))
	}

	// Проверяем методы интерфейса
	bulletsRepo := gameRepo.BulletsRepository()
	if bulletsRepo == nil {
		t.Error("BulletsRepository() should not return nil")
	}
}

func TestGameRepositoriesRegistryAnimations(t *testing.T) {
	gameRepo := game.NewGameRepositoriesRegistry()

	// Проверяем методы интерфейса
	animationsRepo := gameRepo.AnimationsRepository()
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
	gameRepo.TanksRepository().SetPlayer(playerTank)

	// Проверяем, что танк добавлен
	player := gameRepo.TanksRepository().GetPlayer()
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
	gameRepo.TanksRepository().AddEnemy(enemyTank)

	// Проверяем получение всех врагов
	enemies := gameRepo.TanksRepository().GetAllEnemies()
	if len(enemies) != 1 {
		t.Errorf("Expected 1 enemy, got %d", len(enemies))
	}

	// Проверяем методы интерфейса
	tanksRepo := gameRepo.TanksRepository()
	if tanksRepo == nil {
		t.Error("TanksRepository() should not return nil")
	}
}
