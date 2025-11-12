package game

import (
	"testing"

	"github.com/shpaker/gonflict/internal/types"
)

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

	tank := &types.TankEntity{
		Position: types.Position{X: 0, Y: 0},
	}

	bullet := types.NewBulletEntity(
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
		&MockImageProvider{id: "bullet"},
		200.0,
		types.DirectionUp,
		nil,
	)

	err := repo.AddBullet(bullet)
	if err == nil {
		t.Error("Ожидалась ошибка для пули без owner")
	}
}

func TestAddBulletDuplicateOwner(t *testing.T) {
	repo := NewBulletsRepository()

	tank := &types.TankEntity{
		Position: types.Position{X: 0, Y: 0},
	}

	bullet1 := types.NewBulletEntity(
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

	bullet2 := types.NewBulletEntity(
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

	tank1 := &types.TankEntity{
		Position: types.Position{X: 0, Y: 0},
	}
	tank2 := &types.TankEntity{
		Position: types.Position{X: 10, Y: 10},
	}

	bullet1 := types.NewBulletEntity(
		types.Position{X: 100, Y: 100},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&MockImageProvider{id: "bullet"},
		200.0,
		types.DirectionUp,
		tank1,
	)
	bullet2 := types.NewBulletEntity(
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

	err = repo.RemoveBullet(0)
	if err != nil {
		t.Errorf("Неожиданная ошибка при удалении: %v", err)
	}

	bullets = repo.GetAllBullets()
	if len(bullets) != 1 {
		t.Errorf("Ожидалось 1 пуля после удаления, получено %d", len(bullets))
	}

	err = repo.RemoveBullet(5)
	if err == nil {
		t.Error("Ожидалась ошибка для невалидного индекса")
	}
}
