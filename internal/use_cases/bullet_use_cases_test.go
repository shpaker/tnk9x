package use_cases_test

import (
	"errors"
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
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
		bulletUC:    use_cases.NewBulletUseCases(bulletsRepo, tilesUC, 16),
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
	tank.State = types.TankStateStopped
	tank.SetSpecs(env.specsUC.GetTankSpecs(false, level))
	return tank
}

// Позиции вылета пули по направлениям (танк 16px в точке 100,60)
func TestBulletUseCases_ShootBullet_MuzzlePositions(t *testing.T) {
	tests := []struct {
		name      string
		direction types.Direction
		wantX     float64
		wantY     float64
	}{
		{"вверх", types.DirectionUp, 106, 56},
		{"вниз", types.DirectionDown, 106, 68},
		{"влево", types.DirectionLeft, 96, 66},
		{"вправо", types.DirectionRight, 108, 66},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newBulletTestEnv()
			tank := env.newShooter(tt.direction, 100, 60, 0)

			if err := env.bulletUC.ShootBullet(tank); err != nil {
				t.Fatalf("выстрел не удался: %v", err)
			}

			bullets := env.bulletUC.GetBullets()
			if len(bullets) != 1 {
				t.Fatalf("ожидалась 1 пуля, получено %d", len(bullets))
			}

			bullet := bullets[0]
			want := types.Position{X: tt.wantX, Y: tt.wantY}
			if bullet.Position != want {
				t.Errorf(
					"позиция пули %v, ожидалась %v",
					bullet.Position,
					want,
				)
			}
			if bullet.Direction != tt.direction {
				t.Errorf("направление пули %v", bullet.Direction)
			}
			if bullet.GetSize() != (types.Size{Width: 4, Height: 4}) {
				t.Errorf("размер пули %v", bullet.GetSize())
			}
			if bullet.GetAltitude() != types.SURFACE {
				t.Errorf("высота пули %v", bullet.GetAltitude())
			}
			if bullet.GetOwner() != tank {
				t.Errorf("владелец пули не танк-стрелок")
			}
			if bullet.GetSpecs() != tank.GetSpecs() {
				t.Errorf("спецификации пули не от танка")
			}
			if id, err := bullet.GetImageID(); err != nil || id != "bullet" {
				t.Errorf("изображение пули: id=%q err=%v", id, err)
			}
		})
	}
}

// Лимит пуль одного владельца: вторая пуля молча не добавляется
func TestBulletUseCases_ShootBullet_OwnerLimit(t *testing.T) {
	env := newBulletTestEnv()

	// Уровень 0: лимит 1 пуля
	tank := env.newShooter(types.DirectionUp, 100, 60, 0)
	if err := env.bulletUC.ShootBullet(tank); err != nil {
		t.Fatalf("первый выстрел: %v", err)
	}
	if err := env.bulletUC.ShootBullet(tank); err != nil {
		t.Fatalf("повторный выстрел вернул ошибку: %v", err)
	}
	if got := len(env.bulletUC.GetBullets()); got != 1 {
		t.Fatalf("ожидалась 1 пуля при лимите 1, получено %d", got)
	}

	// Уровень 3: лимит 2 пули
	env = newBulletTestEnv()
	tank = env.newShooter(types.DirectionUp, 100, 60, 3)
	_ = env.bulletUC.ShootBullet(tank)
	_ = env.bulletUC.ShootBullet(tank)
	_ = env.bulletUC.ShootBullet(tank)
	if got := len(env.bulletUC.GetBullets()); got != 2 {
		t.Fatalf("ожидалось 2 пули при лимите 2, получено %d", got)
	}
}

func TestBulletUseCases_ShootBullet_InactiveTank(t *testing.T) {
	env := newBulletTestEnv()
	tank := env.newShooter(types.DirectionUp, 100, 60, 0)
	tank.State = types.TankStateSpawning

	if err := env.bulletUC.ShootBullet(tank); err != nil {
		t.Fatalf("неактивный танк вернул ошибку: %v", err)
	}
	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("неактивный танк выстрелил: %d пуль", got)
	}
	if got := len(env.registry.Requested); got != 0 {
		t.Errorf("тайл пули запрошен для неактивного танка: %d", got)
	}
}

func TestBulletUseCases_ShootBullet_TileError(t *testing.T) {
	env := newBulletTestEnv()
	env.registry.Err = errTileNotFound
	tank := env.newShooter(types.DirectionUp, 100, 60, 0)

	if err := env.bulletUC.ShootBullet(tank); err == nil {
		t.Fatal("ожидалась ошибка создания тайла")
	}
	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("пуля добавлена при ошибке тайла: %d", got)
	}
}

// Движение пуль по направлениям со скоростью из спецификаций
func TestBulletUseCases_UpdateBullets_Movement(t *testing.T) {
	tests := []struct {
		name      string
		direction types.Direction
		level     uint
		dt        float64
		wantDX    float64
		wantDY    float64
	}{
		// Уровень 0: скорость пули 120
		{"вверх", types.DirectionUp, 0, 0.5, 0, -60},
		{"вниз", types.DirectionDown, 0, 0.5, 0, 60},
		{"влево", types.DirectionLeft, 0, 0.5, -60, 0},
		{"вправо", types.DirectionRight, 0, 0.5, 60, 0},
		// Уровень 1: скорость пули 150
		{"вверх быстрее", types.DirectionUp, 1, 0.5, 0, -75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newBulletTestEnv()
			tank := env.newShooter(tt.direction, 100, 60, tt.level)
			if err := env.bulletUC.ShootBullet(tank); err != nil {
				t.Fatalf("выстрел: %v", err)
			}

			bullet := env.bulletUC.GetBullets()[0]
			start := bullet.Position

			if err := env.bulletUC.UpdateBullets(tt.dt); err != nil {
				t.Fatalf("обновление: %v", err)
			}

			want := types.Position{
				X: start.X + tt.wantDX,
				Y: start.Y + tt.wantDY,
			}
			if bullet.Position != want {
				t.Errorf(
					"позиция %v, ожидалась %v",
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
	if err := env.bulletUC.ShootBullet(tank); err != nil {
		t.Fatalf("выстрел: %v", err)
	}
	bullet := env.bulletUC.GetBullets()[0]

	if err := env.bulletUC.RemoveBullet(bullet); err != nil {
		t.Fatalf("удаление: %v", err)
	}
	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("пуля не удалена: %d", got)
	}
	if err := env.bulletUC.RemoveBullet(bullet); err == nil {
		t.Error("повторное удаление не вернуло ошибку")
	}
}
