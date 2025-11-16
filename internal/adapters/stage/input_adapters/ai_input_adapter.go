package input_adapters

import (
	"math"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

var _ interfaces.IAiInputAdapter = (*AiInputAdapter)(nil)

type AiInputAdapter struct {
	tankActions    interfaces.ITankActionsUseCases
	tanks          []*types.TankEntity
	updateInterval int
	tickCounter    int
	aiUseCases     *use_cases.AIUseCases
	lastShotTick   map[*types.TankEntity]int
	shootCooldown  int
}

func NewAiInputAdapter(
	tankActions interfaces.ITankActionsUseCases,
	tank *types.TankEntity,
	updateInterval int,
	aiUseCases *use_cases.AIUseCases,
) (*AiInputAdapter, error) {
	adapter := &AiInputAdapter{
		tankActions:    tankActions,
		tanks:          make([]*types.TankEntity, 0, 1),
		updateInterval: updateInterval,
		tickCounter:    0,
		aiUseCases:     aiUseCases,
		lastShotTick:   make(map[*types.TankEntity]int),
		shootCooldown:  20,
	}

	if tank != nil {
		adapter.AddTank(tank)
	}

	return adapter, nil
}

func (a *AiInputAdapter) Update(dt float64) {
	a.tickCounter++

	if len(a.tanks) == 0 {
		return
	}

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

func (a *AiInputAdapter) updateAI(tank *types.TankEntity) {
	if a.aiUseCases != nil {
		decision, err := a.aiUseCases.ExecuteAI(tank)
		if err == nil && tank != nil {

			a.tankActions.ApplyDecision(tank, decision)

			if a.canShoot(tank) {
				_ = a.tankActions.Shoot(tank)
				a.lastShotTick[tank] = a.tickCounter
			}
		}
	}
}

func (a *AiInputAdapter) canShoot(tank *types.TankEntity) bool {
	lastShotTick, ok := a.lastShotTick[tank]
	if !ok {
		lastShotTick = -a.shootCooldown
	}

	ticksSinceLastShot := a.tickCounter - lastShotTick
	return ticksSinceLastShot >= a.shootCooldown
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
	if a.lastShotTick == nil {
		a.lastShotTick = make(map[*types.TankEntity]int)
	}
	a.lastShotTick[tank] = a.tickCounter - a.shootCooldown
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

	if a.lastShotTick != nil {
		delete(a.lastShotTick, tank)
	}
}
