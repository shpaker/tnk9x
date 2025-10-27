package game

import (
	"testing"

	"github.com/shpaker/gonflict/internal/types"
)

// MockImageIdGetter для тестирования
type MockImageIdGetter struct {
	id string
}

func (m *MockImageIdGetter) GetImageId() (string, error) {
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

	// Создаем тестовую пулю с ImageGetter
	bullet := types.BulletEntity{
		ImageGetter:   &MockImageIdGetter{id: "bullet"},
		Position: types.Position{X: 100, Y: 100},
		Speed:         200.0,
		Direction:     types.DirectionUp,
	}

	repo.AddBullet(bullet)

	bullets := repo.GetAllBullets()
	if len(bullets) != 1 {
		t.Errorf("Ожидалось 1 пуля, получено %d", len(bullets))
	}

	// Проверяем, что ImageGetter работает корректно
	if bullets[0].ImageGetter == nil {
		t.Error("ImageGetter не должен быть nil")
	}

	imageId, err := bullets[0].ImageGetter.GetImageId()
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}
	if imageId != "bullet" {
		t.Errorf("Ожидался ID 'bullet', получен '%s'", imageId)
	}
}

func TestRemoveBullet(t *testing.T) {
	repo := NewBulletsRepository()

	// Создаем тестовые пули с ImageGetter
	bullet1 := types.BulletEntity{
		ImageGetter:   &MockImageIdGetter{id: "bullet"},
		Position: types.Position{X: 100, Y: 100},
		Speed:         200.0,
		Direction:     types.DirectionUp,
	}
	bullet2 := types.BulletEntity{
		ImageGetter:   &MockImageIdGetter{id: "bullet"},
		Position: types.Position{X: 200, Y: 200},
		Speed:         200.0,
		Direction:     types.DirectionDown,
	}

	repo.AddBullet(bullet1)
	repo.AddBullet(bullet2)

	bullets := repo.GetAllBullets()
	if len(bullets) != 2 {
		t.Errorf("Ожидалось 2 пули, получено %d", len(bullets))
	}

	// Удаляем первую пулю
	err := repo.RemoveBullet(0)
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
