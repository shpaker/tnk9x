package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/models"
	"github.com/shpaker/gonflict/internal/types"
)

func TestNewBulletsRepository(t *testing.T) {
	repo := NewBulletsRepository()

	if repo == nil {
		t.Fatal("NewBulletsRepository вернул nil")
	}

	if repo.GetBulletsCount() != 0 {
		t.Errorf("Ожидалось 0 пуль, получено %d", repo.GetBulletsCount())
	}
}

func TestAddAndGetBullets(t *testing.T) {
	repo := NewBulletsRepository()

	// Создаем тестовую пулю
	bullet := models.Bullet{
		Image:         ebiten.NewImage(4, 4),
		WorldPosition: types.Position{X: 100, Y: 100},
		Speed:         200.0,
		Direction:     types.DirectionUp,
	}

	repo.AddBullet(bullet)

	if repo.GetBulletsCount() != 1 {
		t.Errorf("Ожидалось 1 пуля, получено %d", repo.GetBulletsCount())
	}

	bullets := repo.GetAllBullets()
	if len(bullets) != 1 {
		t.Errorf("Ожидалось 1 пуля в списке, получено %d", len(bullets))
	}
}

func TestRemoveBullet(t *testing.T) {
	repo := NewBulletsRepository()

	// Создаем тестовые пули
	bullet1 := models.Bullet{
		Image:         ebiten.NewImage(4, 4),
		WorldPosition: types.Position{X: 100, Y: 100},
		Speed:         200.0,
		Direction:     types.DirectionUp,
	}
	bullet2 := models.Bullet{
		Image:         ebiten.NewImage(4, 4),
		WorldPosition: types.Position{X: 200, Y: 200},
		Speed:         200.0,
		Direction:     types.DirectionDown,
	}

	repo.AddBullet(bullet1)
	repo.AddBullet(bullet2)

	if repo.GetBulletsCount() != 2 {
		t.Errorf("Ожидалось 2 пули, получено %d", repo.GetBulletsCount())
	}

	// Удаляем первую пулю
	err := repo.RemoveBullet(0)
	if err != nil {
		t.Errorf("Неожиданная ошибка при удалении: %v", err)
	}

	if repo.GetBulletsCount() != 1 {
		t.Errorf("Ожидалось 1 пуля после удаления, получено %d", repo.GetBulletsCount())
	}

	// Тест невалидного индекса
	err = repo.RemoveBullet(5)
	if err == nil {
		t.Error("Ожидалась ошибка для невалидного индекса")
	}
}

func TestClearAllBullets(t *testing.T) {
	repo := NewBulletsRepository()

	// Создаем тестовые пули
	bullet1 := models.Bullet{
		Image:         ebiten.NewImage(4, 4),
		WorldPosition: types.Position{X: 100, Y: 100},
		Speed:         200.0,
		Direction:     types.DirectionUp,
	}
	bullet2 := models.Bullet{
		Image:         ebiten.NewImage(4, 4),
		WorldPosition: types.Position{X: 200, Y: 200},
		Speed:         200.0,
		Direction:     types.DirectionDown,
	}

	repo.AddBullet(bullet1)
	repo.AddBullet(bullet2)

	if repo.GetBulletsCount() != 2 {
		t.Errorf("Ожидалось 2 пули, получено %d", repo.GetBulletsCount())
	}

	repo.ClearAllBullets()

	if repo.GetBulletsCount() != 0 {
		t.Errorf("Ожидалось 0 пуль после очистки, получено %d", repo.GetBulletsCount())
	}
}
