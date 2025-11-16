package collision_services

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

func TestEntitiesCollisionService_CheckColliders(t *testing.T) {
	service := NewEntitiesCollisionService()

	t.Run("коллизия", func(t *testing.T) {
		tank := &types.TankEntity{
			Position: types.Position{X: 10, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		bullet := types.NewBulletEntity(
			types.Position{X: 15, Y: 15},
			types.Size{Width: 4, Height: 4},
			types.SURFACE,
			nil,
			types.DirectionUp,
			nil,
			nil,
		)

		result := service.CheckColliders(tank, bullet)
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

		bullet := types.NewBulletEntity(
			types.Position{X: 50, Y: 50},
			types.Size{Width: 4, Height: 4},
			types.SURFACE,
			nil,
			types.DirectionUp,
			nil,
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
			Position: types.Position{
				X: 30,
				Y: 10,
			},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		result, err := service.ResolveCollisionPosition(
			entity,
			obstacle,
			types.DirectionRight,
		)
		if err != nil {
			t.Errorf("не ожидалась ошибка, но получили: %v", err)
		}
		if result.X != 14.0 {
			t.Errorf("ожидалась X=14.0, но получили X=%f", result.X)
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
			Position: types.Position{X: 10, Y: 0},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		result, err := service.ResolveCollisionPosition(
			entity,
			obstacle,
			types.DirectionUp,
		)
		if err != nil {
			t.Errorf("не ожидалась ошибка, но получили: %v", err)
		}

		if result.Y != 16.0 {
			t.Errorf("ожидалась Y=16.0, но получили Y=%f", result.Y)
		}
		if result.X != 10.0 {
			t.Errorf("ожидалась X=10.0, но получили X=%f", result.X)
		}
	})

	t.Run("препятствие не по направлению движения", func(t *testing.T) {
		entity := &types.TankEntity{
			Position: types.Position{X: 10, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		obstacle := &types.BlockEntity{
			Position: types.Position{
				X: 0,
				Y: 10,
			},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		_, err := service.ResolveCollisionPosition(
			entity,
			obstacle,
			types.DirectionRight,
		)
		if err == nil {
			t.Errorf("ожидалась ошибка, но получили nil")
		}
	})
}

func TestEntitiesCollisionService_IsObstacleInDirection(t *testing.T) {
	service := NewEntitiesCollisionService()

	t.Run("препятствие по направлению движения", func(t *testing.T) {
		entity := &types.TankEntity{
			Position: types.Position{X: 10, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		obstacle := &types.BlockEntity{
			Position: types.Position{
				X: 30,
				Y: 10,
			},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		result := service.isObstacleInDirection(
			entity,
			obstacle,
			types.DirectionRight,
		)
		if !result {
			t.Errorf("ожидалось true, но получили false")
		}
	})

	t.Run("препятствие не по направлению движения", func(t *testing.T) {
		entity := &types.TankEntity{
			Position: types.Position{X: 10, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		obstacle := &types.BlockEntity{
			Position: types.Position{
				X: 0,
				Y: 10,
			},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		result := service.isObstacleInDirection(
			entity,
			obstacle,
			types.DirectionRight,
		)
		if result {
			t.Errorf("ожидалось false, но получили true")
		}
	})

	t.Run("препятствие по направлению движения вверх", func(t *testing.T) {
		entity := &types.TankEntity{
			Position: types.Position{X: 10, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		obstacle := &types.BlockEntity{
			Position: types.Position{X: 10, Y: 0},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		result := service.isObstacleInDirection(
			entity,
			obstacle,
			types.DirectionUp,
		)
		if !result {
			t.Errorf("ожидалось true, но получили false")
		}
	})

	t.Run("препятствие не по направлению движения вверх", func(t *testing.T) {
		entity := &types.TankEntity{
			Position: types.Position{X: 10, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		obstacle := &types.BlockEntity{
			Position: types.Position{X: 10, Y: 30},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		result := service.isObstacleInDirection(
			entity,
			obstacle,
			types.DirectionUp,
		)
		if result {
			t.Errorf("ожидалось false, но получили true")
		}
	})
}

func TestEntitiesCollisionService_CalculateCorrectedPosition(t *testing.T) {
	service := NewEntitiesCollisionService()

	t.Run("корректировка позиции вправо", func(t *testing.T) {
		entity := &types.TankEntity{
			Position: types.Position{X: 10, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		obstacle := &types.BlockEntity{
			Position: types.Position{
				X: 30,
				Y: 10,
			},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		result := service.calculateCorrectedPosition(
			entity,
			obstacle,
			types.DirectionRight,
		)
		if result.X != 14.0 {
			t.Errorf("ожидалась X=14.0, но получили X=%f", result.X)
		}
		if result.Y != 10.0 {
			t.Errorf("ожидалась Y=10.0, но получили Y=%f", result.Y)
		}
	})

	t.Run("корректировка позиции вверх", func(t *testing.T) {
		entity := &types.TankEntity{
			Position: types.Position{X: 10, Y: 10},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		obstacle := &types.BlockEntity{
			Position: types.Position{X: 10, Y: 0},
			Size:     types.Size{Width: 16, Height: 16},
			Altitude: types.SURFACE,
		}

		result := service.calculateCorrectedPosition(
			entity,
			obstacle,
			types.DirectionUp,
		)
		if result.Y != 16.0 {
			t.Errorf("ожидалась Y=16.0, но получили Y=%f", result.Y)
		}
		if result.X != 10.0 {
			t.Errorf("ожидалась X=10.0, но получили X=%f", result.X)
		}
	})

	t.Run(
		"не изменяет неправильную координату при движении вправо",
		func(t *testing.T) {
			entity := &types.TankEntity{
				Position: types.Position{X: 10, Y: 10},
				Size:     types.Size{Width: 16, Height: 16},
				Altitude: types.SURFACE,
			}

			obstacle := &types.BlockEntity{
				Position: types.Position{
					X: 30,
					Y: 10,
				},
				Size:     types.Size{Width: 16, Height: 16},
				Altitude: types.SURFACE,
			}

			result := service.calculateCorrectedPosition(
				entity,
				obstacle,
				types.DirectionRight,
			)

			if result.Y != 10.0 {
				t.Errorf(
					"Y координата не должна изменяться при движении вправо, ожидалась Y=10.0, но получили Y=%f",
					result.Y,
				)
			}

			if result.X == 10.0 {
				t.Errorf(
					"X координата должна быть скорректирована, но осталась X=10.0",
				)
			}
		},
	)

	t.Run(
		"не изменяет неправильную координату при движении вверх",
		func(t *testing.T) {
			entity := &types.TankEntity{
				Position: types.Position{X: 10, Y: 10},
				Size:     types.Size{Width: 16, Height: 16},
				Altitude: types.SURFACE,
			}

			obstacle := &types.BlockEntity{
				Position: types.Position{
					X: 10,
					Y: 0,
				},
				Size:     types.Size{Width: 16, Height: 16},
				Altitude: types.SURFACE,
			}

			result := service.calculateCorrectedPosition(
				entity,
				obstacle,
				types.DirectionUp,
			)

			if result.X != 10.0 {
				t.Errorf(
					"X координата не должна изменяться при движении вверх, ожидалась X=10.0, но получили X=%f",
					result.X,
				)
			}

			if result.Y == 10.0 {
				t.Errorf(
					"Y координата должна быть скорректирована, но осталась Y=10.0",
				)
			}
		},
	)
}
