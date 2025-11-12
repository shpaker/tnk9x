package collision_services

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

func TestBulletCollisionService_CheckBulletBlockCollision(t *testing.T) {
	entitiesService := NewEntitiesCollisionService()
	service := NewBulletCollisionService(16, entitiesService)

	t.Run("коллизия", func(t *testing.T) {
		bullet := types.NewBulletEntity(
			types.Position{X: 10, Y: 10},
			types.Size{Width: 4, Height: 4},
			types.SURFACE,
			nil,
			0,
			types.DirectionUp,
			nil,
		)

		block := types.NewBlockEntity(
			"brick",
			12.0,
			12.0,
			16,
			&MockImageProvider{id: "brick"},
		)

		result := service.CheckBulletBlockCollision(bullet, block)
		if !result {
			t.Errorf("ожидалась коллизия, но получили false")
		}
	})

	t.Run("нет коллизии", func(t *testing.T) {
		bullet := types.NewBulletEntity(
			types.Position{X: 10, Y: 10},
			types.Size{Width: 4, Height: 4},
			types.SURFACE,
			nil,
			0,
			types.DirectionUp,
			nil,
		)

		block := types.NewBlockEntity(
			"brick",
			50.0,
			50.0,
			16,
			&MockImageProvider{id: "brick"},
		)

		result := service.CheckBulletBlockCollision(bullet, block)
		if result {
			t.Errorf("не ожидалась коллизия, но получили true")
		}
	})
}

func TestBulletCollisionService_CheckBulletTankCollision(t *testing.T) {
	entitiesService := NewEntitiesCollisionService()
	service := NewBulletCollisionService(16, entitiesService)

	t.Run("коллизия", func(t *testing.T) {
		owner := types.NewDefaultTankEntity(
			types.TankRolePlayer1,
			types.DirectionUp,
		)
		owner.Position = types.Position{X: 0, Y: 0}
		owner.Size = types.Size{Width: 16, Height: 16}

		bullet := types.NewBulletEntity(
			types.Position{X: 12, Y: 12},
			types.Size{Width: 4, Height: 4},
			types.SURFACE,
			nil,
			0,
			types.DirectionUp,
			&owner,
		)

		tank := types.NewDefaultTankEntity(
			types.TankRoleEnemy,
			types.DirectionUp,
		)
		tank.Position = types.Position{X: 10, Y: 10}
		tank.Size = types.Size{Width: 16, Height: 16}
		tank.Altitude = types.SURFACE

		result := service.CheckBulletTankCollision(bullet, &tank)
		if !result {
			t.Errorf("ожидалась коллизия, но получили false")
		}
	})

	t.Run("нет коллизии", func(t *testing.T) {
		owner := types.NewDefaultTankEntity(
			types.TankRolePlayer1,
			types.DirectionUp,
		)
		owner.Position = types.Position{X: 0, Y: 0}
		owner.Size = types.Size{Width: 16, Height: 16}

		bullet := types.NewBulletEntity(
			types.Position{X: 10, Y: 10},
			types.Size{Width: 4, Height: 4},
			types.SURFACE,
			nil,
			0,
			types.DirectionUp,
			&owner,
		)

		tank := types.NewDefaultTankEntity(
			types.TankRoleEnemy,
			types.DirectionUp,
		)
		tank.Position = types.Position{X: 50, Y: 50}
		tank.Size = types.Size{Width: 16, Height: 16}

		result := service.CheckBulletTankCollision(bullet, &tank)
		if result {
			t.Errorf("не ожидалась коллизия, но получили true")
		}
	})
}

func TestBulletCollisionService_CheckBulletHQCollision(t *testing.T) {
	entitiesService := NewEntitiesCollisionService()
	service := NewBulletCollisionService(16, entitiesService)

	t.Run("коллизия", func(t *testing.T) {
		owner := types.NewDefaultTankEntity(
			types.TankRolePlayer1,
			types.DirectionUp,
		)
		owner.Position = types.Position{X: 0, Y: 0}
		owner.Size = types.Size{Width: 16, Height: 16}

		bullet := types.NewBulletEntity(
			types.Position{X: 10, Y: 10},
			types.Size{Width: 4, Height: 4},
			types.SURFACE,
			nil,
			0,
			types.DirectionUp,
			&owner,
		)

		hq := &types.HQEntity{
			Position: types.Position{X: 12, Y: 12},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		result := service.CheckBulletHQCollision(bullet, hq)
		if !result {
			t.Errorf("ожидалась коллизия, но получили false")
		}
	})

	t.Run("нет коллизии", func(t *testing.T) {
		owner := types.NewDefaultTankEntity(
			types.TankRolePlayer1,
			types.DirectionUp,
		)
		owner.Position = types.Position{X: 0, Y: 0}
		owner.Size = types.Size{Width: 16, Height: 16}

		bullet := types.NewBulletEntity(
			types.Position{X: 10, Y: 10},
			types.Size{Width: 4, Height: 4},
			types.SURFACE,
			nil,
			0,
			types.DirectionUp,
			&owner,
		)

		hq := &types.HQEntity{
			Position: types.Position{X: 50, Y: 50},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		result := service.CheckBulletHQCollision(bullet, hq)
		if result {
			t.Errorf("не ожидалась коллизия, но получили true")
		}
	})
}
