package input_adapters

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IInputAdapterWithTank = (*CompositeInputAdapter)(nil)

// CompositeInputAdapter объединяет несколько источников ввода одного
// игрока (клавиатура + тач): события применяются в порядке адаптеров,
// при конфликте побеждает последний
type CompositeInputAdapter struct {
	adapters []interfaces.IInputAdapter
}

func NewCompositeInputAdapter(
	adapters ...interfaces.IInputAdapter,
) *CompositeInputAdapter {
	return &CompositeInputAdapter{adapters: adapters}
}

func (a *CompositeInputAdapter) Update(dt float64) {
	for _, adapter := range a.adapters {
		adapter.Update(dt)
	}
}

func (a *CompositeInputAdapter) SetPlayerTank(tank *types.TankEntity) {
	for _, adapter := range a.adapters {
		withTank, ok := adapter.(interfaces.IInputAdapterWithTank)
		if ok {
			withTank.SetPlayerTank(tank)
		}
	}
}
