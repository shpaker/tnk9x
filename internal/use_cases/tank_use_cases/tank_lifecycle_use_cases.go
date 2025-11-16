package tank_use_cases

import (
	"fmt"
	"math/rand"

	"github.com/shpaker/tnk25/internal/interfaces"
	"github.com/shpaker/tnk25/internal/types"
	"github.com/shpaker/tnk25/internal/use_cases"
)

type TankLifecycleUseCases struct {
	tilesUseCases      *use_cases.TilesUseCases
	renderUseCases     interfaces.IRenderUseCases
	tankCommonUseCases interfaces.ITankCommonUseCases
	tanksRepository    interfaces.ITanksRepository
	collisionUseCases  interfaces.ICollisionUseCases
	specsUseCases      interfaces.ISpecsUseCases
	enemySpawners      []types.Position
	player1Spawner     types.Position
	player2Spawner     types.Position
	baseSize           types.Size
}

func NewTankLifecycleUseCases(
	tilesUseCases *use_cases.TilesUseCases,
	renderUseCases interfaces.IRenderUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	enemyRespawnDelay uint,
	specsUseCases interfaces.ISpecsUseCases,
) *TankLifecycleUseCases {
	return &TankLifecycleUseCases{
		tilesUseCases:      tilesUseCases,
		renderUseCases:     renderUseCases,
		tankCommonUseCases: tankCommonUseCases,
		specsUseCases:      specsUseCases,
		player1Spawner:     types.Position{X: 12, Y: 24},
	}
}

func (uc *TankLifecycleUseCases) SetSpawnConfiguration(
	tanksRepository interfaces.ITanksRepository,
	enemySpawners []types.Position,
	player1Spawner types.Position,
	player2Spawner types.Position,
	baseSize types.Size,
) {
	uc.tanksRepository = tanksRepository
	uc.enemySpawners = enemySpawners
	uc.player1Spawner = player1Spawner
	uc.player2Spawner = player2Spawner
	uc.baseSize = baseSize
}

func (uc *TankLifecycleUseCases) SetCollisionUseCases(
	collisionUseCases interfaces.ICollisionUseCases,
) {
	uc.collisionUseCases = collisionUseCases
}

func (uc *TankLifecycleUseCases) SetTankCommonUseCases(
	tankCommonUseCases interfaces.ITankCommonUseCases,
) {
	uc.tankCommonUseCases = tankCommonUseCases
}

func (uc *TankLifecycleUseCases) OnStageSetUpEnemiesSpawn() ([3]*types.TankEntity, error) {
	var spawnedEnemies [3]*types.TankEntity

	if uc.tanksRepository == nil || len(uc.enemySpawners) == 0 {
		return spawnedEnemies, nil
	}

	// Первые три танка всегда 0 уровня
	// Используем большое значение remainingEnemies чтобы получить уровень 0
	remainingEnemies := uint(20) // Максимальное значение для первых трех танков

	for index := 0; index < len(uc.enemySpawners) && index < len(spawnedEnemies); index++ {
		spawned, err := uc.SpawnEnemyWithLevel(&index, true, remainingEnemies)
		if err != nil {
			return spawnedEnemies, err
		}
		spawnedEnemies[index] = spawned
	}

	return spawnedEnemies, nil
}

func (uc *TankLifecycleUseCases) SpawnEnemy(
	index *int,
	ignoreRespawnDelay bool,
) (*types.TankEntity, error) {
	return uc.SpawnEnemyWithLevel(index, ignoreRespawnDelay, 0)
}

func (uc *TankLifecycleUseCases) SpawnEnemyWithLevel(
	index *int,
	ignoreRespawnDelay bool,
	remainingEnemies uint,
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

	// Определяем уровень врага на основе количества оставшихся врагов
	var enemyLevel uint = 0
	if uc.specsUseCases != nil {
		enemyLevel = uc.specsUseCases.GetEnemyLevelByRemainingCount(
			remainingEnemies,
		)
	}

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
	if uc.tanksRepository == nil {
		return nil, fmt.Errorf("tanks repository missing")
	}
	tank, err := uc.spawnTank(
		types.DirectionUp,
		uc.player1Spawner,
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
	if uc.tanksRepository == nil {
		return nil, fmt.Errorf("tanks repository missing")
	}
	tank, err := uc.spawnTank(
		types.DirectionUp,
		uc.player2Spawner,
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

func (uc *TankLifecycleUseCases) GetPlayerTank(
	num types.PlayerTankNum,
) *types.TankEntity {
	if uc.tanksRepository == nil {
		return nil
	}
	return uc.tanksRepository.GetPlayer(num)
}

func (uc *TankLifecycleUseCases) SetPlayerTank(
	num types.PlayerTankNum,
	tank *types.TankEntity,
) {
	if uc.tanksRepository == nil {
		return
	}
	uc.tanksRepository.SetPlayer(num, tank)
}

func (uc *TankLifecycleUseCases) spawnTank(
	direction types.Direction,
	spawnAt types.Position,
	role types.TankRole,
	level uint,
) (types.TankEntity, error) {
	tank := types.NewDefaultTankEntity(role, direction)
	tank.Size = uc.baseSize
	tank.Position = types.Position{
		X: spawnAt.X * float64(uc.baseSize.Width),
		Y: spawnAt.Y * float64(uc.baseSize.Height),
	}

	// Устанавливаем спецификации танка с указанным уровнем
	if uc.specsUseCases != nil {
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

func (uc *TankLifecycleUseCases) IsSpawnFinished(
	tank *types.TankEntity,
	currentTime float64,
) {
	if tank.State == types.TankStateSpawning {
		if uc.renderUseCases.IsTankSpawnAnimationFinished(tank) {
			uc.finishSpawnAnimation(tank)
		}
	}
}

func (uc *TankLifecycleUseCases) IsExplosionFinished(tank *types.TankEntity) {
	if tank.State == types.TankStateExploding {
		if uc.renderUseCases.IsTankExplosionAnimationFinished(tank) {
			tank.State = types.TankStateExploded
		}
	}
}

func (uc *TankLifecycleUseCases) finishSpawnAnimation(
	tank *types.TankEntity,
) {
	if uc.renderUseCases != nil {
		uc.renderUseCases.UpdateTankAnimation(tank)
	}
	tank.State = types.TankStateStopped
}

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

func (uc *TankLifecycleUseCases) updateTankSpawn(tank *types.TankEntity) {
	if tank.State != types.TankStateSpawning {
		return
	}

	if uc.renderUseCases == nil {
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

	if uc.renderUseCases == nil {
		return
	}

	if uc.renderUseCases.IsTankExplosionAnimationFinished(tank) {
		tank.State = types.TankStateExploded
	}
}
