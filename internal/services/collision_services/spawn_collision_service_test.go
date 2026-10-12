package collision_services

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

type spawnCollisionTestEnv struct {
	service *SpawnCollisionService
}

func newSpawnCollisionTestEnv() *spawnCollisionTestEnv {
	return &spawnCollisionTestEnv{
		service: NewSpawnCollisionService(),
	}
}

func (env *spawnCollisionTestEnv) newTankAt(
	x, y float64,
	state types.TankState,
) *types.TankEntity {
	tankValue := types.NewDefaultTankEntity(
		types.TankRoleEnemy,
		types.DirectionUp,
	)
	tank := &tankValue
	tank.Position = types.Position{X: x, Y: y}
	tank.State = state
	return tank
}

func TestSpawnCollisionService_IsSpawnerBlocked(t *testing.T) {
	spawner := types.Position{X: 2, Y: 3}
	size := types.Size{Width: 16, Height: 16}

	tests := []struct {
		name  string
		tanks func(env *spawnCollisionTestEnv) []*types.TankEntity
		want  bool
	}{
		{
			name: "tank overlaps the spawner cell",
			tanks: func(env *spawnCollisionTestEnv) []*types.TankEntity {
				return []*types.TankEntity{
					env.newTankAt(40, 56, types.TankStateActive),
				}
			},
			want: true,
		},
		{
			name: "tank far from the spawner",
			tanks: func(env *spawnCollisionTestEnv) []*types.TankEntity {
				return []*types.TankEntity{
					env.newTankAt(100, 100, types.TankStateActive),
				}
			},
			want: false,
		},
		{
			name: "tank flush against the cell border",
			tanks: func(env *spawnCollisionTestEnv) []*types.TankEntity {
				return []*types.TankEntity{
					env.newTankAt(48, 48, types.TankStateActive),
				}
			},
			want: false,
		},
		{
			name: "exploding tank does not block the spawner",
			tanks: func(env *spawnCollisionTestEnv) []*types.TankEntity {
				return []*types.TankEntity{
					env.newTankAt(40, 56, types.TankStateExploding),
				}
			},
			want: false,
		},
		{
			name: "exploded tank does not block the spawner",
			tanks: func(env *spawnCollisionTestEnv) []*types.TankEntity {
				return []*types.TankEntity{
					env.newTankAt(40, 56, types.TankStateExploded),
				}
			},
			want: false,
		},
		{
			name: "nil tanks are skipped",
			tanks: func(env *spawnCollisionTestEnv) []*types.TankEntity {
				return []*types.TankEntity{
					nil,
					env.newTankAt(40, 56, types.TankStateActive),
				}
			},
			want: true,
		},
		{
			name: "spawner is free without tanks",
			tanks: func(env *spawnCollisionTestEnv) []*types.TankEntity {
				return nil
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newSpawnCollisionTestEnv()
			got := env.service.IsSpawnerBlocked(
				spawner,
				size,
				tt.tanks(env),
			)
			if got != tt.want {
				t.Errorf(
					"IsSpawnerBlocked() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestSpawnCollisionService_IsSpawnerBlocked_ZeroSize(t *testing.T) {
	env := newSpawnCollisionTestEnv()
	tanks := []*types.TankEntity{
		env.newTankAt(0, 0, types.TankStateActive),
	}

	for _, size := range []types.Size{
		{},
		{Width: 16},
		{Height: 16},
	} {
		if env.service.IsSpawnerBlocked(types.Position{}, size, tanks) {
			t.Errorf("size %v: spawner reported as blocked", size)
		}
	}
}
