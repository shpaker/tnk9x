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
		service: NewSpawnCollisionService(NewEntitiesCollisionService()),
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

// Позиция спавнера задана в клетках и масштабируется размером танка:
// клетка {2, 3} при 16px — прямоугольник (32, 48)-(48, 64)
func TestSpawnCollisionService_IsSpawnerBlocked(t *testing.T) {
	spawner := types.Position{X: 2, Y: 3}
	size := types.Size{Width: 16, Height: 16}

	tests := []struct {
		name  string
		tanks func(env *spawnCollisionTestEnv) []*types.TankEntity
		want  bool
	}{
		{
			name: "танк перекрывает клетку спавнера",
			tanks: func(env *spawnCollisionTestEnv) []*types.TankEntity {
				return []*types.TankEntity{
					env.newTankAt(40, 56, types.TankStateStopped),
				}
			},
			want: true,
		},
		{
			name: "танк далеко от спавнера",
			tanks: func(env *spawnCollisionTestEnv) []*types.TankEntity {
				return []*types.TankEntity{
					env.newTankAt(100, 100, types.TankStateStopped),
				}
			},
			want: false,
		},
		{
			name: "танк вплотную к границе клетки",
			tanks: func(env *spawnCollisionTestEnv) []*types.TankEntity {
				return []*types.TankEntity{
					env.newTankAt(48, 48, types.TankStateStopped),
				}
			},
			want: false,
		},
		{
			name: "взрывающийся танк не блокирует спавнер",
			tanks: func(env *spawnCollisionTestEnv) []*types.TankEntity {
				return []*types.TankEntity{
					env.newTankAt(40, 56, types.TankStateExploding),
				}
			},
			want: false,
		},
		{
			name: "взорванный танк не блокирует спавнер",
			tanks: func(env *spawnCollisionTestEnv) []*types.TankEntity {
				return []*types.TankEntity{
					env.newTankAt(40, 56, types.TankStateExploded),
				}
			},
			want: false,
		},
		{
			name: "nil-танки пропускаются",
			tanks: func(env *spawnCollisionTestEnv) []*types.TankEntity {
				return []*types.TankEntity{
					nil,
					env.newTankAt(40, 56, types.TankStateStopped),
				}
			},
			want: true,
		},
		{
			name: "без танков спавнер свободен",
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
					"IsSpawnerBlocked() = %v, ожидалось %v",
					got,
					tt.want,
				)
			}
		})
	}
}

// Нулевой размер отключает проверку блокировки
func TestSpawnCollisionService_IsSpawnerBlocked_ZeroSize(t *testing.T) {
	env := newSpawnCollisionTestEnv()
	tanks := []*types.TankEntity{
		env.newTankAt(0, 0, types.TankStateStopped),
	}

	for _, size := range []types.Size{
		{},
		{Width: 16},
		{Height: 16},
	} {
		if env.service.IsSpawnerBlocked(types.Position{}, size, tanks) {
			t.Errorf("размер %v: спавнер считается заблокированным", size)
		}
	}
}
