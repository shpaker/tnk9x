package collision_services

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
)

func TestWallCollisionService_CheckTankWallCollision(t *testing.T) {
	entitiesService := NewEntitiesCollisionService()
	service := NewWallCollisionService(entitiesService)

	t.Run("коллизия", func(t *testing.T) {
		tank := &types.TankEntity{
			Position: types.Position{X: 10, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		block := types.NewBlockEntity(
			"brick",
			12.0,
			12.0,
			16,
			&testutil.FakeImageProvider{ImageID: "brick"},
		)

		result := service.CheckTankWallCollision(tank, block)
		if !result {
			t.Errorf("ожидалась коллизия, но получили false")
		}
	})

	t.Run("нет коллизии", func(t *testing.T) {
		tank := &types.TankEntity{
			Position: types.Position{X: 10, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		block := types.NewBlockEntity(
			"brick",
			50.0,
			50.0,
			16,
			&testutil.FakeImageProvider{ImageID: "brick"},
		)

		result := service.CheckTankWallCollision(tank, block)
		if result {
			t.Errorf("не ожидалась коллизия, но получили true")
		}
	})
}
