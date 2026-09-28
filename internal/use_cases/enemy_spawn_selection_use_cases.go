package use_cases

import (
	"math"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

// Множители веса спаунера: недавно использованные и соседние с игроком
// точки выбираются реже, но не исключаются совсем
const (
	lastSpawnerWeight     = 0.25
	previousSpawnerWeight = 0.6
	nearPlayerWeight      = 0.15
)

var _ interfaces.IEnemySpawnSelectionUseCases = (*EnemySpawnSelectionUseCases)(
	nil,
)

// EnemySpawnSelectionUseCases выбирает точку спауна врага взвешенным
// случайным выбором среди свободных спаунеров
type EnemySpawnSelectionUseCases struct {
	// Use Cases
	tankCommonUseCases interfaces.ITankCommonUseCases
	// Services
	spawnCollisionService interfaces.ISpawnCollisionService
	randomService         interfaces.IRandomService
	// Configuration
	spawnLayout types.SpawnLayout
	// playerSafeRadius — радиус вокруг спаунера в клетках танка,
	// внутри которого игрок снижает вес точки
	playerSafeRadius float64
}

func NewEnemySpawnSelectionUseCases(
	tankCommonUseCases interfaces.ITankCommonUseCases,
	spawnCollisionService interfaces.ISpawnCollisionService,
	randomService interfaces.IRandomService,
	spawnLayout types.SpawnLayout,
	playerSafeRadius float64,
) *EnemySpawnSelectionUseCases {
	return &EnemySpawnSelectionUseCases{
		tankCommonUseCases:    tankCommonUseCases,
		spawnCollisionService: spawnCollisionService,
		randomService:         randomService,
		spawnLayout:           spawnLayout,
		playerSafeRadius:      playerSafeRadius,
	}
}

func (uc *EnemySpawnSelectionUseCases) GetSpawnersCount() int {
	return len(uc.spawnLayout.EnemySpawners)
}

// SelectSpawner — индекс свободного спаунера; false — все заняты
func (uc *EnemySpawnSelectionUseCases) SelectSpawner(
	session *session_entities.StageSessionEntity,
) (int, bool) {
	tanks := uc.tankCommonUseCases.GetAllTanks()
	weights := make([]float64, len(uc.spawnLayout.EnemySpawners))
	total := 0.0
	for index := range uc.spawnLayout.EnemySpawners {
		weights[index] = uc.spawnerWeight(index, session, tanks)
		total += weights[index]
	}
	if total <= 0 {
		return 0, false
	}

	target := uc.randomService.Float64() * total
	for index, weight := range weights {
		if weight <= 0 {
			continue
		}
		if target < weight {
			return index, true
		}
		target -= weight
	}

	// Погрешность округления: последний свободный спаунер
	for index := len(weights) - 1; index >= 0; index-- {
		if weights[index] > 0 {
			return index, true
		}
	}
	return 0, false
}

// spawnerWeight — вес спаунера; занятый танками получает ноль
func (uc *EnemySpawnSelectionUseCases) spawnerWeight(
	index int,
	session *session_entities.StageSessionEntity,
	tanks []*types.TankEntity,
) float64 {
	position := uc.spawnLayout.EnemySpawners[index]
	if uc.spawnCollisionService.IsSpawnerBlocked(
		position, uc.spawnLayout.BaseSize, tanks,
	) {
		return 0
	}

	weight := 1.0
	recent := session.GetRecentSpawners()
	if recent[0] == index {
		weight *= lastSpawnerWeight
	} else if recent[1] == index {
		weight *= previousSpawnerWeight
	}
	if uc.isPlayerNear(position, tanks) {
		weight *= nearPlayerWeight
	}
	return weight
}

// isPlayerNear — есть ли живой игрок в радиусе от центра спаунера
func (uc *EnemySpawnSelectionUseCases) isPlayerNear(
	spawner types.Position,
	tanks []*types.TankEntity,
) bool {
	cellWidth := float64(uc.spawnLayout.BaseSize.Width)
	cellHeight := float64(uc.spawnLayout.BaseSize.Height)
	centerX := (spawner.X + 0.5) * cellWidth
	centerY := (spawner.Y + 0.5) * cellHeight
	radius := uc.playerSafeRadius * cellWidth

	for _, tank := range tanks {
		if tank == nil || tank.IsEnemy() || !tank.IsActive() {
			continue
		}
		tankX := tank.Position.X + float64(tank.Size.Width)/2
		tankY := tank.Position.Y + float64(tank.Size.Height)/2
		if math.Hypot(tankX-centerX, tankY-centerY) <= radius {
			return true
		}
	}
	return false
}
