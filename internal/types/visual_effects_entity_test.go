package types_test

import (
	"math"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

// Фара встаёт по стволу сразу, затем доворачивается по кратчайшему пути
func TestTankEntity_TurnHeadlight(t *testing.T) {
	tank := types.NewDefaultTankEntity(types.TankRolePlayer1, types.DirectionLeft)
	tank.TurnHeadlight(0.5)
	if got := tank.GetHeadlightAngle(); got != types.DirectionLeft.Angle() {
		t.Fatalf("угол %v, ожидался %v", got, types.DirectionLeft.Angle())
	}

	// Влево (π) → вниз (π/2): кратчайший путь через 3π/4, а не через 0
	tank.Direction = types.DirectionDown
	tank.TurnHeadlight(0.5)
	if got := tank.GetHeadlightAngle(); math.Abs(got-3*math.Pi/4) > 1e-9 {
		t.Errorf("угол %v, ожидался 3π/4", got)
	}

	for range 20 {
		tank.TurnHeadlight(0.5)
	}
	if got := tank.GetHeadlightAngle(); got != types.DirectionDown.Angle() {
		t.Errorf("угол %v, ожидался %v", got, types.DirectionDown.Angle())
	}
}

func TestTankEntity_Recoil(t *testing.T) {
	tank := types.NewDefaultTankEntity(types.TankRolePlayer1, types.DirectionUp)
	tank.StartRecoil(2)
	tank.TickRecoil()
	if !tank.IsRecoiling() {
		t.Fatal("отдача закончилась раньше срока")
	}
	tank.TickRecoil()
	tank.TickRecoil()
	if tank.IsRecoiling() {
		t.Error("отдача не закончилась")
	}
}

func TestDirection_Vector(t *testing.T) {
	cases := map[types.Direction]types.Position{
		types.DirectionUp:    {X: 0, Y: -1},
		types.DirectionDown:  {X: 0, Y: 1},
		types.DirectionLeft:  {X: -1, Y: 0},
		types.DirectionRight: {X: 1, Y: 0},
	}
	for direction, want := range cases {
		if got := direction.Vector(); got != want {
			t.Errorf("%v: вектор %v, ожидался %v", direction, got, want)
		}
	}
}

// Тряска не выше 1 и затухает до нуля
func TestScreenShakeEntity(t *testing.T) {
	var shake types.ScreenShakeEntity
	shake.AddTrauma(0.8)
	shake.AddTrauma(0.8)
	if shake.GetTrauma() != 1 {
		t.Errorf("травма %v, ожидалась 1", shake.GetTrauma())
	}
	shake.Decay(0.7)
	shake.Decay(0.7)
	if shake.GetTrauma() != 0 || shake.GetTicks() != 2 {
		t.Errorf("травма %v, тиков %d", shake.GetTrauma(), shake.GetTicks())
	}
}

func TestParticleEntity_Fade(t *testing.T) {
	particle := types.ParticleEntity{Life: 5, MaxLife: 10}
	if particle.Fade() != 0.5 {
		t.Errorf("доля жизни %v, ожидалась 0.5", particle.Fade())
	}
	if (&types.ParticleEntity{}).Fade() != 0 {
		t.Error("частица без жизни должна быть погасшей")
	}
}

// Поле зрения — передняя полусфера и вплотную со всех сторон
func TestViewerEntity_FieldAt(t *testing.T) {
	viewer := types.ViewerEntity{Direction: types.Position{X: 0, Y: -1}}
	if viewer.FieldAt(types.Position{Y: -100}) != 1 {
		t.Error("точка впереди вне поля зрения")
	}
	if viewer.FieldAt(types.Position{Y: 100}) != 0 {
		t.Error("точка сзади в поле зрения")
	}
	if viewer.FieldAt(types.Position{Y: 20}) != 1 {
		t.Error("точка вплотную сзади не замечена")
	}
}

// Фара светит вперёд, не назад, и не слепит у самого танка
func TestLightEntity_IlluminationAt(t *testing.T) {
	cone := types.LightEntity{
		Radius:    130,
		Intensity: 1,
		Direction: types.Position{X: 0, Y: -1},
		ConeCos:   0.8,
	}
	far := cone.IlluminationAt(types.Position{Y: -60})
	near := cone.IlluminationAt(types.Position{Y: -5})
	if far <= near {
		t.Errorf("у танка %v, впереди %v: фара должна набирать силу", near, far)
	}
	if cone.IlluminationAt(types.Position{Y: 60}) != 0 {
		t.Error("фара светит назад")
	}

	omni := types.LightEntity{Radius: 40, Intensity: 1}
	if omni.IlluminationAt(types.Position{X: 10}) <= 0 ||
		omni.IlluminationAt(types.Position{X: 50}) != 0 {
		t.Error("всенаправленный свет вне радиуса или не светит внутри")
	}
}
