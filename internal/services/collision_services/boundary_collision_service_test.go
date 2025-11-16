package collision_services

import (
	"testing"

	"github.com/shpaker/tnk25/internal/types"
)

func TestBoundaryCollisionService_CheckLeftBoundaryCollision(t *testing.T) {
	service := NewBoundaryCollisionService(100, 16)

	t.Run("коллизия", func(t *testing.T) {
		tank := &types.TankEntity{
			Position: types.Position{X: -5, Y: 50},
			Size:     types.Size{Width: 16, Height: 16},
		}

		result := service.CheckLeftBoundaryCollision(tank)
		if !result {
			t.Errorf("ожидалась коллизия, но получили false")
		}
	})

	t.Run("нет коллизии", func(t *testing.T) {
		tank := &types.TankEntity{
			Position: types.Position{X: 10, Y: 50},
			Size:     types.Size{Width: 16, Height: 16},
		}

		result := service.CheckLeftBoundaryCollision(tank)
		if result {
			t.Errorf("не ожидалась коллизия, но получили true")
		}
	})
}

func TestBoundaryCollisionService_CheckRightBoundaryCollision(
	t *testing.T,
) {
	service := NewBoundaryCollisionService(100, 16)

	t.Run("коллизия", func(t *testing.T) {
		tank := &types.TankEntity{
			Position: types.Position{X: 90, Y: 50},
			Size:     types.Size{Width: 16, Height: 16},
		}

		result := service.CheckRightBoundaryCollision(tank)
		if !result {
			t.Errorf("ожидалась коллизия, но получили false")
		}
	})

	t.Run("нет коллизии", func(t *testing.T) {
		tank := &types.TankEntity{
			Position: types.Position{X: 50, Y: 50},
			Size:     types.Size{Width: 16, Height: 16},
		}

		result := service.CheckRightBoundaryCollision(tank)
		if result {
			t.Errorf("не ожидалась коллизия, но получили true")
		}
	})
}

func TestBoundaryCollisionService_CheckTopBoundaryCollision(t *testing.T) {
	service := NewBoundaryCollisionService(100, 16)

	t.Run("коллизия", func(t *testing.T) {
		tank := &types.TankEntity{
			Position: types.Position{X: 50, Y: -5},
			Size:     types.Size{Width: 16, Height: 16},
		}

		result := service.CheckTopBoundaryCollision(tank)
		if !result {
			t.Errorf("ожидалась коллизия, но получили false")
		}
	})

	t.Run("нет коллизии", func(t *testing.T) {
		tank := &types.TankEntity{
			Position: types.Position{X: 50, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
		}

		result := service.CheckTopBoundaryCollision(tank)
		if result {
			t.Errorf("не ожидалась коллизия, но получили true")
		}
	})
}

func TestBoundaryCollisionService_CheckBottomBoundaryCollision(
	t *testing.T,
) {
	service := NewBoundaryCollisionService(100, 16)

	t.Run("коллизия", func(t *testing.T) {
		tank := &types.TankEntity{
			Position: types.Position{X: 50, Y: 90},
			Size:     types.Size{Width: 16, Height: 16},
		}

		result := service.CheckBottomBoundaryCollision(tank)
		if !result {
			t.Errorf("ожидалась коллизия, но получили false")
		}
	})

	t.Run("нет коллизии", func(t *testing.T) {
		tank := &types.TankEntity{
			Position: types.Position{X: 50, Y: 50},
			Size:     types.Size{Width: 16, Height: 16},
		}

		result := service.CheckBottomBoundaryCollision(tank)
		if result {
			t.Errorf("не ожидалась коллизия, но получили true")
		}
	})
}
