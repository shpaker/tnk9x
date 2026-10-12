package use_cases_test

import (
	"errors"
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
	"github.com/shpaker/tnk9x/internal/use_cases/tank_use_cases"
)

var errTileNotFound = errors.New("tile not found")

type bulletTestEnv struct {
	registry    *testutil.FakeTilesetRegistry
	bulletsRepo *game.BulletsRepository
	specsUC     *use_cases.SpecsUseCases
	bulletUC    *use_cases.BulletUseCases
}

func newBulletTestEnv() *bulletTestEnv {
	registry := &testutil.FakeTilesetRegistry{}
	bulletsRepo := game.NewBulletsRepository()
	tilesUC := use_cases.NewTilesUseCases(
		registry,
		types.TilesetTypeBullet,
		nil,
		nil,
	)

	return &bulletTestEnv{
		registry:    registry,
		bulletsRepo: bulletsRepo,
		specsUC:     use_cases.NewSpecsUseCases(),
		bulletUC: use_cases.NewBulletUseCases(
			bulletsRepo,
			tilesUC,
			16,
			false,
		),
	}
}

func (env *bulletTestEnv) newShooter(
	direction types.Direction,
	x, y float64,
	level uint,
) *types.TankEntity {
	tankValue := types.NewDefaultTankEntity(types.TankRolePlayer1, direction)
	tank := &tankValue
	tank.Position = types.Position{X: x, Y: y}
	tank.State = types.TankStateActive
	tank.SetSpecs(env.specsUC.GetTankSpecs(false, level))
	return tank
}

func TestBulletUseCases_ShootBullet_MuzzlePositions(t *testing.T) {
	tests := []struct {
		name      string
		direction types.Direction
		wantX     float64
		wantY     float64
	}{
		{"up", types.DirectionUp, 106, 56},
		{"down", types.DirectionDown, 106, 68},
		{"left", types.DirectionLeft, 96, 66},
		{"right", types.DirectionRight, 108, 66},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newBulletTestEnv()
			tank := env.newShooter(tt.direction, 100, 60, 0)

			if _, err := env.bulletUC.ShootBullet(tank); err != nil {
				t.Fatalf("shot failed: %v", err)
			}

			bullets := env.bulletUC.GetBullets()
			if len(bullets) != 1 {
				t.Fatalf("want 1 bullet, got %d", len(bullets))
			}

			bullet := bullets[0]
			want := types.Position{X: tt.wantX, Y: tt.wantY}
			if bullet.Position != want {
				t.Errorf(
					"bullet position %v, want %v",
					bullet.Position,
					want,
				)
			}
			if bullet.Direction != tt.direction {
				t.Errorf("bullet direction %v", bullet.Direction)
			}
			if bullet.GetSize() != (types.Size{Width: 4, Height: 4}) {
				t.Errorf("bullet size %v", bullet.GetSize())
			}
			if bullet.GetAltitude() != types.SURFACE {
				t.Errorf("bullet altitude %v", bullet.GetAltitude())
			}
			if bullet.GetOwner() != tank {
				t.Errorf("bullet owner is not the shooter")
			}
			if bullet.GetSpecs() != tank.GetSpecs() {
				t.Errorf("bullet specs are not the tank specs")
			}
			if id, err := bullet.GetImageID(); err != nil || id != "bullet" {
				t.Errorf("bullet image: id=%q err=%v", id, err)
			}
		})
	}
}

// The second bullet is silently rejected; the shot does not count.
func TestBulletUseCases_ShootBullet_OwnerLimit(t *testing.T) {
	env := newBulletTestEnv()

	tank := env.newShooter(types.DirectionUp, 100, 60, 0)
	if fired, err := env.bulletUC.ShootBullet(tank); err != nil || !fired {
		t.Fatalf("first shot: fired=%v err=%v", fired, err)
	}
	if fired, err := env.bulletUC.ShootBullet(tank); err != nil || fired {
		t.Fatalf("second shot: fired=%v err=%v", fired, err)
	}
	if got := len(env.bulletUC.GetBullets()); got != 1 {
		t.Fatalf("want 1 bullet at limit 1, got %d", got)
	}

	env = newBulletTestEnv()
	tank = env.newShooter(types.DirectionUp, 100, 60, 3)
	_, _ = env.bulletUC.ShootBullet(tank)
	_, _ = env.bulletUC.ShootBullet(tank)
	_, _ = env.bulletUC.ShootBullet(tank)
	if got := len(env.bulletUC.GetBullets()); got != 2 {
		t.Fatalf("want 2 bullets at limit 2, got %d", got)
	}
}

func TestBulletUseCases_ShootBullet_InactiveTank(t *testing.T) {
	env := newBulletTestEnv()
	tank := env.newShooter(types.DirectionUp, 100, 60, 0)
	tank.State = types.TankStateSpawning

	if fired, err := env.bulletUC.ShootBullet(tank); err != nil || fired {
		t.Fatalf("inactive tank: fired=%v err=%v", fired, err)
	}
	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("inactive tank fired: %d bullets", got)
	}
	if got := len(env.registry.Requested); got != 0 {
		t.Errorf("bullet tile requested for inactive tank: %d", got)
	}
}

func TestBulletUseCases_ShootBullet_TileError(t *testing.T) {
	env := newBulletTestEnv()
	env.registry.Err = errTileNotFound
	tank := env.newShooter(types.DirectionUp, 100, 60, 0)

	if _, err := env.bulletUC.ShootBullet(tank); err == nil {
		t.Fatal("want tile creation error")
	}
	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("bullet added despite tile error: %d", got)
	}
}

func TestBulletUseCases_UpdateBullets_Movement(t *testing.T) {
	tests := []struct {
		name      string
		direction types.Direction
		level     uint
		dt        float64
		wantDX    float64
		wantDY    float64
	}{
		{"up", types.DirectionUp, 0, 0.5, 0, -60},
		{"down", types.DirectionDown, 0, 0.5, 0, 60},
		{"left", types.DirectionLeft, 0, 0.5, -60, 0},
		{"right", types.DirectionRight, 0, 0.5, 60, 0},
		{"up faster", types.DirectionUp, 1, 0.5, 0, -75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newBulletTestEnv()
			tank := env.newShooter(tt.direction, 100, 60, tt.level)
			if _, err := env.bulletUC.ShootBullet(tank); err != nil {
				t.Fatalf("shot: %v", err)
			}

			bullet := env.bulletUC.GetBullets()[0]
			start := bullet.Position

			if err := env.bulletUC.UpdateBullets(tt.dt); err != nil {
				t.Fatalf("update: %v", err)
			}

			want := types.Position{
				X: start.X + tt.wantDX,
				Y: start.Y + tt.wantDY,
			}
			if bullet.Position != want {
				t.Errorf(
					"position %v, want %v",
					bullet.Position,
					want,
				)
			}
		})
	}
}

func TestBulletUseCases_RemoveBullet(t *testing.T) {
	env := newBulletTestEnv()
	tank := env.newShooter(types.DirectionUp, 100, 60, 0)
	if _, err := env.bulletUC.ShootBullet(tank); err != nil {
		t.Fatalf("shot: %v", err)
	}
	bullet := env.bulletUC.GetBullets()[0]

	if err := env.bulletUC.RemoveBullet(bullet); err != nil {
		t.Fatalf("removal: %v", err)
	}
	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("bullet not removed: %d", got)
	}
	if err := env.bulletUC.RemoveBullet(bullet); err == nil {
		t.Error("second removal returned no error")
	}
}

// While a bullet is in flight, pressing again gives no sound or flash.
func TestTankActionsUseCases_Shoot_NoEffectsWhileBulletFlies(t *testing.T) {
	env := newBulletTestEnv()
	soundUC := use_cases.NewSoundUseCases(game.NewSoundEventsRepository())
	effects := &testutil.FakeVisualEffectsUseCases{}
	tankActions := tank_use_cases.NewTankActionsUseCases(
		env.bulletUC,
		&testutil.FakeTankCommonUseCases{},
		nil,
		soundUC,
		effects,
	)
	tank := env.newShooter(types.DirectionUp, 100, 60, 0)

	for range 3 {
		if err := tankActions.Shoot(tank); err != nil {
			t.Fatalf("shot: %v", err)
		}
	}

	if got := len(effects.Events); got != 1 {
		t.Errorf("shot effects %d, want 1", got)
	}
	if got := countSounds(soundUC.GetEvents(), types.SoundIDFire); got != 1 {
		t.Errorf("shot sounds %d, want 1", got)
	}
}

// With cooldown a repeat shot needs a pause, even if the last bullet is gone.
func TestBulletUseCases_ShootBullet_Reload(t *testing.T) {
	env := newBulletTestEnv()
	tilesUC := use_cases.NewTilesUseCases(
		env.registry,
		types.TilesetTypeBullet,
		nil,
		nil,
	)
	bulletUC := use_cases.NewBulletUseCases(env.bulletsRepo, tilesUC, 16, true)
	tank := env.newShooter(types.DirectionUp, 100, 60, 0)

	if fired, _ := bulletUC.ShootBullet(tank); !fired {
		t.Fatal("first shot must succeed")
	}
	for _, bullet := range bulletUC.GetBullets() {
		_ = bulletUC.RemoveBullet(bullet)
	}
	if fired, _ := bulletUC.ShootBullet(tank); fired {
		t.Fatal("no shot expected during cooldown")
	}

	tank.UpdateReload(tank.GetSpecs().GetReloadTime())
	if fired, _ := bulletUC.ShootBullet(tank); !fired {
		t.Error("shot must succeed after cooldown")
	}
}
