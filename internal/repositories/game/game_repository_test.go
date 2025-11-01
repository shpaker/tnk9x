package game_test

import (
	"testing"

	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/types"
)

func TestNewGameRepositoriesRegistry(t *testing.T) {
	gameRepo := game.NewGameRepositoriesRegistry()

	// Проверяем, что все репозитории созданы
	if gameRepo.BlocksRepository() == nil {
		t.Error("Blocks repository should not be nil")
	}
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

func TestGameRepositoriesRegistryBlocks(t *testing.T) {
	gameRepo := game.NewGameRepositoriesRegistry()

	// Создаем блок с минимальными данными
	block := types.BlockEntity{
		Position: types.Position{X: 10, Y: 20},
		Altitude: types.SURFACE,
		Data: &types.BlockData{
			Name:     types.Brick,
			Position: types.Position{X: 10, Y: 20},
		},
	}

	// Добавляем блок
	gameRepo.BlocksRepository().AddBlock(block)

	// Проверяем, что блок добавлен
	blocks := gameRepo.BlocksRepository().GetAllBlocks()
	if len(*blocks) != 1 {
		t.Errorf("Expected 1 block, got %d", len(*blocks))
	}

	// Проверяем методы интерфейса
	blocksRepo := gameRepo.BlocksRepository()
	if blocksRepo == nil {
		t.Error("BlocksRepository() should not return nil")
	}
}

func TestGameRepositoriesRegistryBullets(t *testing.T) {
	gameRepo := game.NewGameRepositoriesRegistry()

	// Создаем пулю с минимальными данными
	bullet := types.BulletEntity{
		Position:  types.Position{X: 10, Y: 20},
		Direction: types.DirectionUp,
		Speed:     100,
		Altitude:  types.SURFACE,
	}

	// Добавляем пулю
	gameRepo.BulletsRepository().AddBullet(bullet)

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

func TestGetGameContext(t *testing.T) {
	gameRepo := game.NewGameRepositoriesRegistry()

	// Добавляем игрока
	playerTank := &types.TankEntity{
		Position:  types.Position{X: 100, Y: 200},
		Direction: types.DirectionUp,
		Speed:     0,
	}
	gameRepo.TanksRepository().SetPlayer(playerTank)

	// Добавляем врага
	enemyTank := &types.TankEntity{
		Position:  types.Position{X: 300, Y: 400},
		Direction: types.DirectionDown,
		Speed:     0,
	}
	gameRepo.TanksRepository().AddEnemy(enemyTank)

	// Добавляем пулю
	bullet := types.BulletEntity{
		Position:  types.Position{X: 150, Y: 250},
		Direction: types.DirectionUp,
		Speed:     100,
		Altitude:  types.SURFACE,
	}
	gameRepo.BulletsRepository().AddBullet(bullet)

	// Добавляем блок
	block := types.BlockEntity{
		Position: types.Position{X: 50, Y: 50},
		Altitude: types.SURFACE,
		Data: &types.BlockData{
			Name:     types.Brick,
			Position: types.Position{X: 50, Y: 50},
		},
	}
	gameRepo.BlocksRepository().AddBlock(block)

	// Получаем контекст игры
	context := gameRepo.GetGameContext()

	// Проверяем, что контекст не nil
	if context == nil {
		t.Fatal("GameContext should not be nil")
	}

	// Проверяем игрока
	if context.Player == nil {
		t.Fatal("Player should not be nil")
	}
	if context.Player.Position.X != 100 || context.Player.Position.Y != 200 {
		t.Error("Player position mismatch")
	}

	// Проверяем врагов
	if len(context.Enemies) != 1 {
		t.Errorf("Expected 1 enemy, got %d", len(context.Enemies))
	}
	if context.Enemies[0].Position.X != 300 ||
		context.Enemies[0].Position.Y != 400 {
		t.Error("Enemy position mismatch")
	}

	// Проверяем пули
	if len(context.Bullets) != 1 {
		t.Errorf("Expected 1 bullet, got %d", len(context.Bullets))
	}
	if context.Bullets[0].Position.X != 150 ||
		context.Bullets[0].Position.Y != 250 {
		t.Error("Bullet position mismatch")
	}

	// Проверяем блоки
	if len(context.Blocks) != 1 {
		t.Errorf("Expected 1 block, got %d", len(context.Blocks))
	}
	if context.Blocks[0].Position.X != 50 ||
		context.Blocks[0].Position.Y != 50 {
		t.Error("Block position mismatch")
	}
}
