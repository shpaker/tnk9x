package game

import (
	"testing"

	"github.com/shpaker/gonflict/internal/types"
)

// MockImageProvider для тестирования
type MockImageProvider struct {
	id string
}

func (m *MockImageProvider) GetImageID() (string, error) {
	return m.id, nil
}

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

	// Создаем тестовый танк
	tank := &types.TankEntity{
		Position: types.Position{X: 0, Y: 0},
	}

	// Создаем тестовую пулю с Image и Owner
	bullet := *types.NewBulletEntity(
		types.Position{X: 100, Y: 100},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&MockImageProvider{id: "bullet"},
		200.0,
		types.DirectionUp,
		tank,
	)

	err := repo.AddBullet(bullet)
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}

	bullets := repo.GetAllBullets()
	if len(bullets) != 1 {
		t.Errorf("Ожидалось 1 пуля, получено %d", len(bullets))
	}

	// Проверяем, что Image работает корректно
	if bullets[0].Image == nil {
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

	// Создаем пулю без owner
	bullet := *types.NewBulletEntity(
		types.Position{X: 100, Y: 100},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&MockImageProvider{id: "bullet"},
		200.0,
		types.DirectionUp,
		nil, // Owner остается nil
	)

	err := repo.AddBullet(bullet)
	if err == nil {
		t.Error("Ожидалась ошибка для пули без owner")
	}
}

func TestAddBulletDuplicateOwner(t *testing.T) {
	repo := NewBulletsRepository()

	// Создаем тестовый танк
	tank := &types.TankEntity{
		Position: types.Position{X: 0, Y: 0},
	}

	// Создаем первую пулю
	bullet1 := *types.NewBulletEntity(
		types.Position{X: 100, Y: 100},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&MockImageProvider{id: "bullet"},
		200.0,
		types.DirectionUp,
		tank,
	)

	err := repo.AddBullet(bullet1)
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}

	// Пытаемся добавить вторую пулю от того же owner
	bullet2 := *types.NewBulletEntity(
		types.Position{X: 200, Y: 200},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&MockImageProvider{id: "bullet"},
		200.0,
		types.DirectionDown,
		tank,
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

	// Создаем тестовые танки
	tank1 := &types.TankEntity{
		Position: types.Position{X: 0, Y: 0},
	}
	tank2 := &types.TankEntity{
		Position: types.Position{X: 10, Y: 10},
	}

	// Создаем тестовые пули с Image и Owner
	bullet1 := *types.NewBulletEntity(
		types.Position{X: 100, Y: 100},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&MockImageProvider{id: "bullet"},
		200.0,
		types.DirectionUp,
		tank1,
	)
	bullet2 := *types.NewBulletEntity(
		types.Position{X: 200, Y: 200},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&MockImageProvider{id: "bullet"},
		200.0,
		types.DirectionDown,
		tank2,
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

	// Удаляем первую пулю
	err = repo.RemoveBullet(0)
	if err != nil {
		t.Errorf("Неожиданная ошибка при удалении: %v", err)
	}

	bullets = repo.GetAllBullets()
	if len(bullets) != 1 {
		t.Errorf("Ожидалось 1 пуля после удаления, получено %d", len(bullets))
	}

	// Тест невалидного индекса
	err = repo.RemoveBullet(5)
	if err == nil {
		t.Error("Ожидалась ошибка для невалидного индекса")
	}
}
