package input_adapters

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

// recordingInputAdapter — фиксирует вызовы Update
type recordingInputAdapter struct {
	updates int
}

func (r *recordingInputAdapter) Update(float64) { r.updates++ }

// recordingInputAdapterWithTank — фиксирует и Update, и SetPlayerTank
type recordingInputAdapterWithTank struct {
	recordingInputAdapter
	tank *types.TankEntity
}

func (r *recordingInputAdapterWithTank) SetPlayerTank(
	tank *types.TankEntity,
) {
	r.tank = tank
}

func TestCompositeInputAdapter_FanOut(t *testing.T) {
	plain := &recordingInputAdapter{}
	withTank := &recordingInputAdapterWithTank{}
	composite := NewCompositeInputAdapter(plain, withTank)

	composite.Update(0.016)
	if plain.updates != 1 || withTank.updates != 1 {
		t.Error("Update должен доходить до всех адаптеров")
	}

	tank := &types.TankEntity{}
	composite.SetPlayerTank(tank)
	if withTank.tank != tank {
		t.Error("SetPlayerTank должен доходить до адаптеров с танком")
	}
}
