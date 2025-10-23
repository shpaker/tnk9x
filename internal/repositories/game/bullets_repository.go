package game

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/types"
)

type BulletsRepository struct {
	bullets []types.BulletEntity
}

func NewBulletsRepository() *BulletsRepository {
	return &BulletsRepository{
		bullets: make([]types.BulletEntity, 0),
	}
}

// AddBullet добавляет пулю в репозиторий
func (br *BulletsRepository) AddBullet(bullet types.BulletEntity) {
	br.bullets = append(br.bullets, bullet)
}

// GetAllBullets возвращает все пули
func (br *BulletsRepository) GetAllBullets() []types.BulletEntity {
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
