package input_adapters

import (
	"math"
	"math/rand"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

var _ interfaces.IAiInputAdapter = (*AiInputAdapter)(nil)

// Фазы AI по времени этапа: свободный обход, охота на игрока,
// атака штаба (~60 и ~120 секунд при 60 TPS)
const (
	aiPhaseWanderTicks = 3600
	aiPhaseHuntTicks   = 7200
)

// Случайный интервал между выстрелами врага в тиках
const (
	aiShotDelayMin  = 30
	aiShotDelaySpan = 61
)

type AiInputAdapter struct {
	tankActions        interfaces.ITankActionsUseCases
	aiUseCases         interfaces.IAIUseCases
	tankCommonUseCases interfaces.ITankCommonUseCases
	stageSession       *session_entities.StageSessionEntity

	tanks          []*types.TankEntity
	updateInterval int
	tickCounter    int
	lastShotTick   map[*types.TankEntity]int
	shotDelay      map[*types.TankEntity]int
	hqPosition     types.Position
}

func NewAiInputAdapter(
	tankActions interfaces.ITankActionsUseCases,
	tank *types.TankEntity,
	updateInterval int,
	aiUseCases interfaces.IAIUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	stageSession *session_entities.StageSessionEntity,
	hqPosition types.Position,
) (*AiInputAdapter, error) {
	adapter := &AiInputAdapter{
		tankActions:        tankActions,
		aiUseCases:         aiUseCases,
		tankCommonUseCases: tankCommonUseCases,
		stageSession:       stageSession,
		tanks:              make([]*types.TankEntity, 0, 1),
		updateInterval:     updateInterval,
		tickCounter:        0,
		lastShotTick:       make(map[*types.TankEntity]int),
		shotDelay:          make(map[*types.TankEntity]int),
		hqPosition:         hqPosition,
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

	a.pruneExplodedTanks()

	// Решения AI принимаются раз в updateInterval тиков;
	// контроль торможения работает каждый тик
	makeDecisions := a.updateInterval <= 0 ||
		a.tickCounter%a.updateInterval == 0

	for _, tank := range a.tanks {
		if tank == nil || !tank.IsActive() {
			continue
		}

		if tank.State == types.TankStateMoving {
			a.checkAndSetBraking(tank)
		}

		if makeDecisions && tank.IsStopped() {
			a.updateAI(tank)
		}
	}
}

// pruneExplodedTanks убирает взорванные танки из списка и карт
// кулдаунов, иначе они копятся до конца уровня
func (a *AiInputAdapter) pruneExplodedTanks() {
	for i := len(a.tanks) - 1; i >= 0; i-- {
		tank := a.tanks[i]
		if tank == nil || tank.State == types.TankStateExploded {
			a.tanks = append(a.tanks[:i], a.tanks[i+1:]...)
			if tank != nil {
				delete(a.lastShotTick, tank)
				delete(a.shotDelay, tank)
			}
		}
	}
}

func (a *AiInputAdapter) updateAI(tank *types.TankEntity) {
	if a.aiUseCases == nil {
		return
	}

	decision, err := a.aiUseCases.ExecuteAI(tank, a.buildContext(tank))
	if err != nil {
		return
	}

	a.tankActions.ApplyDecision(tank, decision)

	if a.canShoot(tank) {
		_ = a.tankActions.Shoot(tank)
		a.lastShotTick[tank] = a.tickCounter
		a.shotDelay[tank] = randomShotDelay()
	}
}

// buildContext определяет фазу AI по времени этапа и цель:
// на охоте — ближайший игрок, при атаке — штаб
func (a *AiInputAdapter) buildContext(
	tank *types.TankEntity,
) types.EnemyAIContext {
	stageTicks := uint(0)
	if a.stageSession != nil {
		stageTicks = a.stageSession.GetStageTicks()
	}

	switch {
	case stageTicks < aiPhaseWanderTicks:
		return types.EnemyAIContext{Phase: types.EnemyAIPhaseWander}
	case stageTicks < aiPhaseHuntTicks:
		if player := a.nearestPlayer(tank); player != nil {
			return types.EnemyAIContext{
				Phase:     types.EnemyAIPhaseHunt,
				TargetX:   player.Position.X,
				TargetY:   player.Position.Y,
				HasTarget: true,
			}
		}
		return types.EnemyAIContext{Phase: types.EnemyAIPhaseHunt}
	default:
		return types.EnemyAIContext{
			Phase:     types.EnemyAIPhaseSiege,
			TargetX:   a.hqPosition.X,
			TargetY:   a.hqPosition.Y,
			HasTarget: true,
		}
	}
}

// nearestPlayer — ближайший активный танк игрока
func (a *AiInputAdapter) nearestPlayer(
	tank *types.TankEntity,
) *types.TankEntity {
	if a.tankCommonUseCases == nil {
		return nil
	}

	var nearest *types.TankEntity
	nearestDistance := math.MaxFloat64
	for _, player := range a.tankCommonUseCases.GetAllPlayerTanks() {
		if player == nil || !player.IsActive() {
			continue
		}
		distance := math.Abs(player.Position.X-tank.Position.X) +
			math.Abs(player.Position.Y-tank.Position.Y)
		if distance < nearestDistance {
			nearestDistance = distance
			nearest = player
		}
	}
	return nearest
}

func randomShotDelay() int {
	return aiShotDelayMin + rand.Intn(aiShotDelaySpan)
}

func (a *AiInputAdapter) canShoot(tank *types.TankEntity) bool {
	delay, ok := a.shotDelay[tank]
	if !ok {
		delay = randomShotDelay()
		a.shotDelay[tank] = delay
	}

	lastShotTick, ok := a.lastShotTick[tank]
	if !ok {
		lastShotTick = a.tickCounter - delay
	}

	return a.tickCounter-lastShotTick >= delay
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
	a.shotDelay[tank] = randomShotDelay()
	a.lastShotTick[tank] = a.tickCounter
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

	delete(a.lastShotTick, tank)
	delete(a.shotDelay, tank)
}
