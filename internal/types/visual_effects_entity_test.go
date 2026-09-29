package types_test

import (
	"math"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

// Фара встаёт по стволу сразу, затем доворачивается по кратчайшему пути
func TestTankEntity_TurnHeadlight(t *testing.T) {
	tank := types.NewDefaultTankEntity(
		types.TankRolePlayer1,
		types.DirectionLeft,
	)
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

// Фаза силового поля чередуется каждые ShieldFlickerTicks тиков
func TestTankEntity_ShieldPhase(t *testing.T) {
	tank := types.NewDefaultTankEntity(types.TankRolePlayer1, types.DirectionUp)
	tank.ActivateShield(4)
	first := tank.GetShieldPhase()
	for range types.ShieldFlickerTicks {
		tank.UpdateShieldCountdown()
	}
	if tank.GetShieldPhase() == first {
		t.Error("фаза поля не сменилась")
	}
}
