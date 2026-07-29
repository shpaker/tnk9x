package tank_use_cases

import (
	"fmt"
	"math/rand"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ITankLifecycleUseCases = (*TankLifecycleUseCases)(nil)

type TankLifecycleUseCases struct {
	tilesUseCases         interfaces.ITilesUseCases
	renderUseCases        interfaces.IRenderUseCases
	tankCommonUseCases    interfaces.ITankCommonUseCases
	tanksRepository       interfaces.ITanksRepository
	spawnCollisionService interfaces.ISpawnCollisionService
	specsUseCases         interfaces.ISpecsUseCases
	spawnLayout           types.SpawnLayout
}

func NewTankLifecycleUseCases(
	tilesUseCases interfaces.ITilesUseCases,
	renderUseCases interfaces.IRenderUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	tanksRepository interfaces.ITanksRepository,
	spawnCollisionService interfaces.ISpawnCollisionService,
	specsUseCases interfaces.ISpecsUseCases,
	spawnLayout types.SpawnLayout,
) *TankLifecycleUseCases {
	return &TankLifecycleUseCases{
		tilesUseCases:         tilesUseCases,
		renderUseCases:        renderUseCases,
		tankCommonUseCases:    tankCommonUseCases,
		tanksRepository:       tanksRepository,
		spawnCollisionService: spawnCollisionService,
		specsUseCases:         specsUseCases,
		spawnLayout:           spawnLayout,
	}
}

func (uc *TankLifecycleUseCases) OnStageSetUpEnemiesSpawn() ([3]*types.TankEntity, error) {
	var spawnedEnemies [3]*types.TankEntity

	if len(uc.spawnLayout.EnemySpawners) == 0 {
		return spawnedEnemies, nil
	}

	// Первые три танка всегда 0 уровня
	// Используем большое значение remainingEnemies чтобы получить уровень 0
	remainingEnemies := uint(20) // Максимальное значение для первых трех танков

	for index := 0; index < len(uc.spawnLayout.EnemySpawners) &&
		index < len(spawnedEnemies); index++ {
		spawned, err := uc.SpawnEnemyWithLevel(&index, true, remainingEnemies)
		if err != nil {
			return spawnedEnemies, err
		}
		spawnedEnemies[index] = spawned
	}

	return spawnedEnemies, nil
}

func (uc *TankLifecycleUseCases) SpawnEnemyWithLevel(
	index *int,
	ignoreRespawnDelay bool,
	remainingEnemies uint,
) (*types.TankEntity, error) {
	selectedIndex := 0
	if index != nil {
		selectedIndex = *index
	} else if len(uc.spawnLayout.EnemySpawners) > 0 {
		selectedIndex = rand.Intn(len(uc.spawnLayout.EnemySpawners))
	} else {
		return nil, fmt.Errorf("enemy spawners missing")
	}

	if selectedIndex >= len(uc.spawnLayout.EnemySpawners) {
		return nil, fmt.Errorf("enemy spawner index out of range")
	}

	spawnPosition := uc.spawnLayout.EnemySpawners[selectedIndex]

	spawnerBlocked := uc.isSpawnerBlocked(spawnPosition)

	if !ignoreRespawnDelay && spawnerBlocked {
		return nil, nil
	}

	// Определяем уровень врага на основе количества оставшихся врагов
	enemyLevel := uc.specsUseCases.GetEnemyLevelByRemainingCount(
		remainingEnemies,
	)

	tank, err := uc.spawnTank(
		types.DirectionUp,
		spawnPosition,
		types.TankRoleEnemy,
		enemyLevel,
	)
	if err != nil {
		return nil, err
	}

	uc.tanksRepository.AddEnemy(&tank)

	return &tank, nil
}

func (uc *TankLifecycleUseCases) SpawnPlayer1() (*types.TankEntity, error) {
	if uc.isSpawnerBlocked(uc.spawnLayout.Player1Spawner) {
		return nil, nil
	}
	tank, err := uc.spawnTank(
		types.DirectionUp,
		uc.spawnLayout.Player1Spawner,
		types.TankRolePlayer1,
		0, // Игроки всегда начинают с уровня 0
	)
	if err != nil {
		return nil, err
	}

	tankPtr := new(types.TankEntity)
	*tankPtr = tank
	uc.tanksRepository.SetPlayer(types.PlayerTankNumPlayer1, tankPtr)
	return tankPtr, nil
}

func (uc *TankLifecycleUseCases) SpawnPlayer2() (*types.TankEntity, error) {
	if uc.isSpawnerBlocked(uc.spawnLayout.Player2Spawner) {
		return nil, nil
	}
	tank, err := uc.spawnTank(
		types.DirectionUp,
		uc.spawnLayout.Player2Spawner,
		types.TankRolePlayer2,
		0, // Игроки всегда начинают с уровня 0
	)
	if err != nil {
		return nil, err
	}

	tankPtr := new(types.TankEntity)
	*tankPtr = tank
	uc.tanksRepository.SetPlayer(types.PlayerTankNumPlayer2, tankPtr)
	return tankPtr, nil
}

// isSpawnerBlocked проверяет перекрытие спавнера живыми танками
func (uc *TankLifecycleUseCases) isSpawnerBlocked(
	position types.Position,
) bool {
	return uc.spawnCollisionService.IsSpawnerBlocked(
		position,
		uc.spawnLayout.BaseSize,
		uc.tankCommonUseCases.GetAllTanks(),
	)
}

func (uc *TankLifecycleUseCases) GetPlayerTank(
	num types.PlayerTankNum,
) *types.TankEntity {
	return uc.tanksRepository.GetPlayer(num)
}

func (uc *TankLifecycleUseCases) SetPlayerTank(
	num types.PlayerTankNum,
	tank *types.TankEntity,
) {
	uc.tanksRepository.SetPlayer(num, tank)
}

func (uc *TankLifecycleUseCases) spawnTank(
	direction types.Direction,
	spawnAt types.Position,
	role types.TankRole,
	level uint,
) (types.TankEntity, error) {
	tank := types.NewDefaultTankEntity(role, direction)
	tank.Size = uc.spawnLayout.BaseSize
	tank.Position = types.Position{
		X: spawnAt.X * float64(uc.spawnLayout.BaseSize.Width),
		Y: spawnAt.Y * float64(uc.spawnLayout.BaseSize.Height),
	}

	// Устанавливаем спецификации танка с указанным уровнем
	specs := uc.specsUseCases.GetTankSpecs(
		role == types.TankRoleEnemy,
		level,
	)
	if specs != nil {
		tank.SetSpecs(specs)
		// Для тяжёлого танка (враг уровня 3) устанавливаем 4 попадания
		if role == types.TankRoleEnemy && level == 3 {
			tank.SetHitPoints(4)
		} else {
			// Для остальных танков 1 попадание
			tank.SetHitPoints(1)
		}
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

func (uc *TankLifecycleUseCases) finishSpawnAnimation(
	tank *types.TankEntity,
) {
	uc.renderUseCases.UpdateTankAnimation(tank)
	tank.State = types.TankStateStopped
}

func (uc *TankLifecycleUseCases) UpdateAllTanksLifecycle() error {
	allTanks := uc.tankCommonUseCases.GetAllTanks()
	for _, tank := range allTanks {
		if tank != nil {
			uc.updateTankSpawn(tank)
			uc.updateTankExplosion(tank)
		}
	}
	return nil
}

func (uc *TankLifecycleUseCases) updateTankSpawn(tank *types.TankEntity) {
	if tank.State != types.TankStateSpawning {
		return
	}

	if uc.renderUseCases.IsTankSpawnAnimationFinished(tank) {
		uc.finishSpawnAnimation(tank)
	}
}

func (uc *TankLifecycleUseCases) updateTankExplosion(tank *types.TankEntity) {
	if tank.State != types.TankStateExploding {
		return
	}

	if uc.renderUseCases.IsTankExplosionAnimationFinished(tank) {
		tank.State = types.TankStateExploded
	}
}
