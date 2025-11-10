package tank_use_cases

import (
	"fmt"
	"math/rand"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// TankLifecycleUseCases отвечает за жизненный цикл танка (спавн, взрыв, анимации)
type TankLifecycleUseCases struct {
	tilesUseCases      *use_cases.TilesUseCases
	renderUseCases     interfaces.ITankRenderUseCases
	tankCommonUseCases interfaces.ITankCommonUseCases
	tanksRepository    interfaces.ITanksRepository
	collisionUseCases  interfaces.ICollisionUseCases
	enemySpawners      []types.Position
	player1Spawner     types.Position
	baseSize           types.Size
}

// ============================================================================
// Конфигурация
// ============================================================================

// NewTankLifecycleUseCases создает новый экземпляр TankLifecycleUseCases
func NewTankLifecycleUseCases(
	tilesUseCases *use_cases.TilesUseCases,
	renderUseCases interfaces.ITankRenderUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	enemyRespawnDelay uint,
) *TankLifecycleUseCases {
	return &TankLifecycleUseCases{
		tilesUseCases:      tilesUseCases,
		renderUseCases:     renderUseCases,
		tankCommonUseCases: tankCommonUseCases,
		player1Spawner:     types.Position{X: 12, Y: 24},
	}
}

// SetSpawnConfiguration настраивает параметры спавна танков
func (uc *TankLifecycleUseCases) SetSpawnConfiguration(
	tanksRepository interfaces.ITanksRepository,
	enemySpawners []types.Position,
	player1Spawner types.Position,
	baseSize types.Size,
) {
	uc.tanksRepository = tanksRepository
	uc.enemySpawners = enemySpawners
	uc.player1Spawner = player1Spawner
	uc.baseSize = baseSize
}

// SetCollisionUseCases устанавливает зависимость от use cases коллизий
func (uc *TankLifecycleUseCases) SetCollisionUseCases(
	collisionUseCases interfaces.ICollisionUseCases,
) {
	uc.collisionUseCases = collisionUseCases
}

// SetTankCommonUseCases обновляет ссылку на общие use cases танков.
func (uc *TankLifecycleUseCases) SetTankCommonUseCases(
	tankCommonUseCases interfaces.ITankCommonUseCases,
) {
	uc.tankCommonUseCases = tankCommonUseCases
}

// ============================================================================
// Спавн игрока и врагов
// ============================================================================

// OnStageSetUpEnemiesSpawn спавнит до трех вражеских танков
// OnStageSetUpEnemiesSpawn спавнит до трех вражеских танков
func (uc *TankLifecycleUseCases) OnStageSetUpEnemiesSpawn() ([3]*types.TankEntity, error) {
	var spawnedEnemies [3]*types.TankEntity

	if uc.tanksRepository == nil || len(uc.enemySpawners) == 0 {
		return spawnedEnemies, nil
	}

	for index := 0; index < len(uc.enemySpawners) && index < len(spawnedEnemies); index++ {
		spawned, err := uc.SpawnEnemy(&index, true)
		if err != nil {
			return spawnedEnemies, err
		}
		spawnedEnemies[index] = spawned
	}

	return spawnedEnemies, nil
}

// SpawnEnemy спавнит конкретного врага по индексу
func (uc *TankLifecycleUseCases) SpawnEnemy(
	index *int,
	ignoreRespawnDelay bool,
) (*types.TankEntity, error) {
	if uc.tanksRepository == nil {
		return nil, fmt.Errorf("tanks repository missing")
	}

	selectedIndex := 0
	if index != nil {
		selectedIndex = *index
	} else if len(uc.enemySpawners) > 0 {
		selectedIndex = rand.Intn(len(uc.enemySpawners))
	} else {
		return nil, fmt.Errorf("enemy spawners missing")
	}

	if selectedIndex >= len(uc.enemySpawners) {
		return nil, fmt.Errorf("enemy spawner index out of range")
	}

	spawnPosition := uc.enemySpawners[selectedIndex]

	spawnerBlocked := false
	if uc.collisionUseCases != nil {
		spawnerBlocked = uc.collisionUseCases.IsSpawnerBlocked(
			spawnPosition,
			uc.baseSize,
		)
	}

	if !ignoreRespawnDelay && spawnerBlocked {
		return nil, nil
	}

	tank, err := uc.spawnTank(types.DirectionUp, spawnPosition, true)
	if err != nil {
		return nil, err
	}

	uc.tanksRepository.AddEnemy(&tank)

	return &tank, nil
}

// SpawnPlayer1 создает и спавнит игрока из первого спавнера
func (uc *TankLifecycleUseCases) SpawnPlayer1() (*types.TankEntity, error) {
	if uc.tanksRepository == nil {
		return nil, fmt.Errorf("tanks repository missing")
	}
	tank, err := uc.spawnTank(types.DirectionUp, uc.player1Spawner, false)
	if err != nil {
		return nil, err
	}
	uc.tanksRepository.SetPlayer(&tank)
	return &tank, nil
}

// GetPlayerTank возвращает танк игрока из репозитория
func (uc *TankLifecycleUseCases) GetPlayerTank() *types.TankEntity {
	if uc.tanksRepository == nil {
		return nil
	}
	return uc.tanksRepository.GetPlayer()
}

// SetPlayerTank устанавливает танк игрока в репозитории
func (uc *TankLifecycleUseCases) SetPlayerTank(tank *types.TankEntity) {
	if uc.tanksRepository == nil {
		return
	}
	uc.tanksRepository.SetPlayer(tank)
}

// ============================================================================
// Базовые операции жизненного цикла
// ============================================================================

// spawnTank создает танк и запускает процесс спавна с анимацией
func (uc *TankLifecycleUseCases) spawnTank(
	direction types.Direction,
	spawnAt types.Position,
	isEnemy bool,
) (types.TankEntity, error) {
	tank := types.NewDefaultTankEntity(isEnemy, direction)
	tank.Size = uc.baseSize
	tank.Position = types.Position{
		X: spawnAt.X * float64(uc.baseSize.Width),
		Y: spawnAt.Y * float64(uc.baseSize.Height),
	}
	if uc.tanksRepository == nil {
		return tank, fmt.Errorf("tanks repository missing")
	}
	spawnAnimation, err := uc.tilesUseCases.CreateSpawnAnimation()
	if err != nil {
		return tank, err
	}
	tank.Image = spawnAnimation
	tank.State = types.TankStateSpawning
	tank.Altitude = types.SURFACE
	uc.tilesUseCases.StartAnimation(spawnAnimation)
	return tank, nil
}

// Explode устанавливает и запускает анимацию взрыва для танка
func (uc *TankLifecycleUseCases) Explode(tank *types.TankEntity) error {
	explosionAnim, err := uc.tilesUseCases.CreateExplosionAnimation()
	if err != nil {
		return err
	}

	tank.Image = explosionAnim
	tank.State = types.TankStateExploding
	tank.Altitude = types.AIR

	uc.tilesUseCases.StartAnimation(explosionAnim)
	return nil
}

// ============================================================================
// Проверки завершения анимаций
// ============================================================================

// IsSpawnFinished проверяет и обновляет процесс спавна танка
func (uc *TankLifecycleUseCases) IsSpawnFinished(
	tank *types.TankEntity,
	currentTime float64,
) {
	if tank.State == types.TankStateSpawning {
		if uc.renderUseCases.IsSpawnAnimationFinished(tank) {
			uc.finishSpawnAnimation(tank)
		}
	}
}

// IsExplosionFinished проверяет завершение анимации взрыва танка
func (uc *TankLifecycleUseCases) IsExplosionFinished(tank *types.TankEntity) {
	if tank.State == types.TankStateExploding {
		if uc.renderUseCases.IsExplosionAnimationFinished(tank) {
			tank.State = types.TankStateExploded
		}
	}
}

// finishSpawnAnimation завершает анимацию спавна и устанавливает анимацию танка
func (uc *TankLifecycleUseCases) finishSpawnAnimation(
	tank *types.TankEntity,
) {
	if uc.renderUseCases != nil {
		uc.renderUseCases.UpdateTankAnimation(tank)
	}
	tank.State = types.TankStateStopped
}

// ============================================================================
// Обновление состояния всех танков
// ============================================================================

// UpdateAllTanksLifecycle обновляет жизненный цикл всех танков (спавн и взрыв)
func (uc *TankLifecycleUseCases) UpdateAllTanksLifecycle() error {
	if uc.tankCommonUseCases == nil {
		return nil
	}
	allTanks := uc.tankCommonUseCases.GetAllTanks()
	for _, tank := range allTanks {
		if tank != nil {
			uc.updateTankSpawn(tank)
			uc.updateTankExplosion(tank)
		}
	}
	return nil
}

// updateTankSpawn проверяет и обновляет процесс спавна танка
func (uc *TankLifecycleUseCases) updateTankSpawn(tank *types.TankEntity) {
	if tank.State != types.TankStateSpawning {
		return
	}

	if uc.renderUseCases == nil {
		return
	}

	if uc.renderUseCases.IsSpawnAnimationFinished(tank) {
		uc.finishSpawnAnimation(tank)
	}
}

// updateTankExplosion проверяет завершение анимации взрыва танка
func (uc *TankLifecycleUseCases) updateTankExplosion(tank *types.TankEntity) {
	if tank.State != types.TankStateExploding {
		return
	}

	if uc.renderUseCases == nil {
		return
	}

	if uc.renderUseCases.IsExplosionAnimationFinished(tank) {
		tank.State = types.TankStateExploded
	}
}
