package game

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/models"
)

type BulletsRepository struct {
	bullets []models.Bullet
}

func NewBulletsRepository() *BulletsRepository {
	return &BulletsRepository{
		bullets: make([]models.Bullet, 0),
	}
}

// AddBullet добавляет пулю в репозиторий
func (br *BulletsRepository) AddBullet(bullet models.Bullet) {
	br.bullets = append(br.bullets, bullet)
}

// GetAllBullets возвращает все пули
func (br *BulletsRepository) GetAllBullets() []models.Bullet {
	return br.bullets
}

// RemoveBullet удаляет пулю по индексу
func (br *BulletsRepository) RemoveBullet(index int) error {
	if index < 0 || index >= len(br.bullets) {
		return fmt.Errorf("индекс пули %d вне диапазона [0, %d)", index, len(br.bullets))
	}
	br.bullets = append(br.bullets[:index], br.bullets[index+1:]...)
	return nil
}

// ClearAllBullets очищает все пули
func (br *BulletsRepository) ClearAllBullets() {
	br.bullets = make([]models.Bullet, 0)
}

// GetBulletsCount возвращает количество пуль
func (br *BulletsRepository) GetBulletsCount() int {
	return len(br.bullets)
}
