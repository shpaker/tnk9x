package game

import (
	"fmt"

	"github.com/shpaker/tnk25/internal/types"
)

type BulletsRepository struct {
	bullets []*types.BulletEntity
}

func NewBulletsRepository() *BulletsRepository {
	return &BulletsRepository{
		bullets: make([]*types.BulletEntity, 0),
	}
}

func (br *BulletsRepository) AddBullet(bullet *types.BulletEntity) error {
	if bullet == nil || bullet.GetOwner() == nil {
		return fmt.Errorf("bullet owner is nil")
	}

	owner := bullet.GetOwner()

	// Получаем лимит пуль из спецификаций пули
	bulletLimit := uint(1) // Значение по умолчанию
	if bullet.GetSpecs() != nil {
		bulletLimit = bullet.GetSpecs().GetBulletsLimit()
	}

	// Подсчитываем количество активных пуль у владельца
	activeBulletsCount := uint(0)
	for _, existingBullet := range br.bullets {
		if existingBullet != nil &&
			existingBullet.GetOwner() == owner {
			activeBulletsCount++
		}
	}

	// Проверяем, не превышен ли лимит
	if activeBulletsCount >= bulletLimit {
		return fmt.Errorf("tank bullet limit exceeded")
	}

	br.bullets = append(br.bullets, bullet)
	return nil
}

func (br *BulletsRepository) GetAllBullets() []*types.BulletEntity {
	return br.bullets
}

func (br *BulletsRepository) RemoveBullet(index int) error {
	if index < 0 || index >= len(br.bullets) {
		return fmt.Errorf(
			"bullet index %d out of range [0, %d)",
			index,
			len(br.bullets),
		)
	}
	br.bullets = append(br.bullets[:index], br.bullets[index+1:]...)
	return nil
}
