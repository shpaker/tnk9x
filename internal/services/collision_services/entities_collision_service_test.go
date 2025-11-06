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

		bullet := types.NewBulletEntity(
			types.Position{X: 15, Y: 15}, // Пуля внутри танка
			types.Size{Width: 4, Height: 4},
			types.SURFACE,
			nil,
			0,
			types.DirectionUp,
			nil,
		)

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

		bullet := types.NewBulletEntity(
			types.Position{X: 50, Y: 50}, // Пуля далеко от танка
			types.Size{Width: 4, Height: 4},
			types.SURFACE,
			nil,
			0,
			types.DirectionUp,
			nil,
		)

		result := service.CheckColliders(tank, bullet)
		if result {
			t.Errorf("не ожидалась коллизия, но получили true")
		}
	})
}

func TestEntitiesCollisionService_ResolveCollisionPosition(t *testing.T) {
	service := NewEntitiesCollisionService()

	t.Run("корректировка позиции", func(t *testing.T) {
		entity := &types.TankEntity{
			Position: types.Position{X: 10, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		obstacle := &types.BlockEntity{
			Position: types.Position{X: 20, Y: 20},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		result := service.ResolveCollisionPosition(
			entity,
			obstacle,
			types.DirectionRight,
		)
		if result.X != 4.0 {
			t.Errorf("ожидалась X=4.0, но получили X=%f", result.X)
		}
		if result.Y != 10.0 {
			t.Errorf("ожидалась Y=10.0, но получили Y=%f", result.Y)
		}
	})

	t.Run("корректировка для другого направления", func(t *testing.T) {
		entity := &types.TankEntity{
			Position: types.Position{X: 10, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		obstacle := &types.BlockEntity{
			Position: types.Position{X: 20, Y: 20},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		result := service.ResolveCollisionPosition(
			entity,
			obstacle,
			types.DirectionUp,
		)
		// При движении вверх позиция Y должна быть скорректирована
		if result.Y != 36.0 {
			t.Errorf("ожидалась Y=36.0, но получили Y=%f", result.Y)
		}
		if result.X != 10.0 {
			t.Errorf("ожидалась X=10.0, но получили X=%f", result.X)
		}
	})
}
