package input_adapters

import (
	"math"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IAiInputAdapter = (*AiInputAdapter)(nil)

// AiInputAdapter — источник управления вражескими танками: на каждой
// остановке запрашивает решение AI и исполняет его
type AiInputAdapter struct {
	tanks       []*types.TankEntity
	tickCounter int

	// Use Cases
	tankActions interfaces.ITankActionsUseCases
	aiUseCases  interfaces.IAIUseCases
}

func NewAiInputAdapter(
	tankActions interfaces.ITankActionsUseCases,
	aiUseCases interfaces.IAIUseCases,
) *AiInputAdapter {
	return &AiInputAdapter{
		tanks:       make([]*types.TankEntity, 0),
		tankActions: tankActions,
		aiUseCases:  aiUseCases,
	}
}

func (a *AiInputAdapter) Update(dt float64) {
	a.tickCounter++

	for _, tank := range a.tanks {
		if tank == nil || !tank.IsActive() {
			continue
		}

		if tank.State == types.TankStateMoving {
			a.checkAndSetBraking(tank)
		}

		if tank.IsStopped() {
			a.updateAI(tank)
		}
	}
}

// updateAI исполняет решение скрипта: поворот, движение и выстрел;
// лимит пуль соблюдает BulletUseCases
func (a *AiInputAdapter) updateAI(tank *types.TankEntity) {
	decision, err := a.aiUseCases.ExecuteAI(tank, a.tickCounter)
	if err != nil {
		return
	}

	a.tankActions.ApplyDecision(tank, decision)

	if decision.Shoot {
		_ = a.tankActions.Shoot(tank)
	}
}

func (a *AiInputAdapter) checkAndSetBraking(tank *types.TankEntity) {
	if tank == nil {
		return
	}

	var coord float64
	switch tank.Direction {
	case types.DirectionUp, types.DirectionDown:
		coord = tank.Position.Y
	case types.DirectionLeft, types.DirectionRight:
		coord = tank.Position.X
	default:
		return
	}

	remainder := math.Mod(coord, 8)
	if remainder < 0 {
		remainder += 8
	}

	if remainder <= 2 || remainder >= 6 {
		tank.State = types.TankStateBraking
	}
}

func (a *AiInputAdapter) AddTank(tank *types.TankEntity) {
	if tank == nil {
		return
	}

	for _, current := range a.tanks {
		if current == tank {
			return
		}
	}

	a.tanks = append(a.tanks, tank)
}

func (a *AiInputAdapter) RemoveTank(tank *types.TankEntity) {
	if tank == nil {
		return
	}

	for i, current := range a.tanks {
		if current == tank {
			a.tanks = append(a.tanks[:i], a.tanks[i+1:]...)
			break
		}
	}
}
