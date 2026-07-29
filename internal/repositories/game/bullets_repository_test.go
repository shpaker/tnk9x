package game

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
)

func TestNewBulletsRepository(t *testing.T) {
	repo := NewBulletsRepository()

	if repo == nil {
		t.Fatal("NewBulletsRepository вернул nil")
	}

	bullets := repo.GetAllBullets()
	if len(bullets) != 0 {
		t.Errorf("Ожидалось 0 пуль, получено %d", len(bullets))
	}
}

func TestAddAndGetBullets(t *testing.T) {
	repo := NewBulletsRepository()

	tank := types.NewDefaultTankEntity(types.TankRolePlayer1, types.DirectionUp)
	tank.Position = types.Position{X: 0, Y: 0}
	tankPtr := &tank

	// Создаем specs для теста с лимитом 1 пуля
	specs := types.NewSpecsEntity(
		0, // Уровень 0
		32.0,
		false, // Пули не усиленные
		200.0, // Скорость пули
		1,     // Лимит пуль: 1
	)
	bullet := types.NewBulletEntity(
		types.Position{X: 100, Y: 100},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&testutil.FakeImageProvider{ImageID: "bullet"},
		types.DirectionUp,
		specs,
		tankPtr,
	)

	err := repo.AddBullet(bullet)
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}

	bullets := repo.GetAllBullets()
	if len(bullets) != 1 {
		t.Errorf("Ожидалось 1 пуля, получено %d", len(bullets))
	}

	if bullets[0] == nil || bullets[0].Image == nil {
		t.Error("Image не должен быть nil")
	}

	imageID, err := bullets[0].Image.GetImageID()
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}
	if imageID != "bullet" {
		t.Errorf("Ожидался ID 'bullet', получен '%s'", imageID)
	}
}

func TestAddBulletWithoutOwner(t *testing.T) {
	repo := NewBulletsRepository()

	bullet := types.NewBulletEntity(
		types.Position{X: 100, Y: 100},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&testutil.FakeImageProvider{ImageID: "bullet"},
		types.DirectionUp,
		nil,
		nil,
	)

	err := repo.AddBullet(bullet)
	if err == nil {
		t.Error("Ожидалась ошибка для пули без owner")
	}
}

func TestAddBulletDuplicateOwner(t *testing.T) {
	repo := NewBulletsRepository()

	tank := types.NewDefaultTankEntity(types.TankRolePlayer1, types.DirectionUp)
	tank.Position = types.Position{X: 0, Y: 0}
	tankPtr := &tank

	// Создаем specs для теста с лимитом 1 пуля
	specs := types.NewSpecsEntity(
		0, // Уровень 0
		32.0,
		false, // Пули не усиленные
		200.0, // Скорость пули
		1,     // Лимит пуль: 1
	)
	bullet1 := types.NewBulletEntity(
		types.Position{X: 100, Y: 100},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&testutil.FakeImageProvider{ImageID: "bullet"},
		types.DirectionUp,
		specs,
		tankPtr,
	)

	err := repo.AddBullet(bullet1)
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}

	bullet2 := types.NewBulletEntity(
		types.Position{X: 200, Y: 200},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&testutil.FakeImageProvider{ImageID: "bullet"},
		types.DirectionDown,
		specs,
		tankPtr,
	)

	err = repo.AddBullet(bullet2)
	if err == nil {
		t.Error("Ожидалась ошибка для второй пули от того же owner")
	}

	bullets := repo.GetAllBullets()
	if len(bullets) != 1 {
		t.Errorf("Ожидалось 1 пуля, получено %d", len(bullets))
	}
}

func TestRemoveBullet(t *testing.T) {
	repo := NewBulletsRepository()

	tank1 := types.NewDefaultTankEntity(
		types.TankRolePlayer1,
		types.DirectionUp,
	)
	tank1.Position = types.Position{X: 0, Y: 0}
	tank1Ptr := &tank1

	tank2 := types.NewDefaultTankEntity(
		types.TankRolePlayer2,
		types.DirectionUp,
	)
	tank2.Position = types.Position{X: 10, Y: 10}
	tank2Ptr := &tank2

	specs := types.NewSpecsEntity(
		0, // Уровень 0
		32.0,
		false, // Пули не усиленные
		200.0, // Скорость пули
		1,     // Лимит пуль: 1
	)
	bullet1 := types.NewBulletEntity(
		types.Position{X: 100, Y: 100},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&testutil.FakeImageProvider{ImageID: "bullet"},
		types.DirectionUp,
		specs,
		tank1Ptr,
	)
	bullet2 := types.NewBulletEntity(
		types.Position{X: 200, Y: 200},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&testutil.FakeImageProvider{ImageID: "bullet"},
		types.DirectionDown,
		specs,
		tank2Ptr,
	)

	err := repo.AddBullet(bullet1)
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}
	err = repo.AddBullet(bullet2)
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}

	bullets := repo.GetAllBullets()
	if len(bullets) != 2 {
		t.Errorf("Ожидалось 2 пули, получено %d", len(bullets))
	}

	err = repo.RemoveBullet(bullet1)
	if err != nil {
		t.Errorf("Неожиданная ошибка при удалении: %v", err)
	}

	bullets = repo.GetAllBullets()
	if len(bullets) != 1 {
		t.Errorf("Ожидалось 1 пуля после удаления, получено %d", len(bullets))
	}
	if bullets[0] != bullet2 {
		t.Error("Ожидалось, что останется bullet2")
	}

	err = repo.RemoveBullet(bullet1)
	if err == nil {
		t.Error("Ожидалась ошибка при повторном удалении той же пули")
	}

	err = repo.RemoveBullet(nil)
	if err == nil {
		t.Error("Ожидалась ошибка для nil-пули")
	}
}
