package collision_services

import (
	"testing"

	"github.com/shpaker/gonflict/internal/types"
)

func TestEntitiesCollisionService_CheckColliders(t *testing.T) {
	service := NewEntitiesCollisionService()

	t.Run("коллизия", func(t *testing.T) {
		// Создаем два объекта, которые пересекаются
		tank := &types.TankEntity{
			Position: types.Position{X: 10, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		bullet := &types.BulletEntity{
			Position: types.Position{X: 15, Y: 15}, // Пуля внутри танка
			Size:     types.Size{Width: 4, Height: 4},
			Altitude: types.SURFACE,
		}

		result := service.CheckColliders(tank, bullet)
		if !result {
			t.Errorf("ожидалась коллизия, но получили false")
		}
	})

	t.Run("нет коллизии", func(t *testing.T) {
		// Создаем два объекта, которые не пересекаются
		tank := &types.TankEntity{
			Position: types.Position{X: 10, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		bullet := &types.BulletEntity{
			Position: types.Position{X: 50, Y: 50}, // Пуля далеко от танка
			Size:     types.Size{Width: 4, Height: 4},
			Altitude: types.SURFACE,
		}

		result := service.CheckColliders(tank, bullet)
		if result {
			t.Errorf("не ожидалась коллизия, но получили true")
		}
	})
}
