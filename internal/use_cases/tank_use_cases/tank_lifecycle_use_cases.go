package tank_use_cases

import (
	"fmt"

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

// SpawnEnemy спавнит врага заданного уровня; точка спавна выбирается
// циклическим перебором по порядковому номеру спавна, как в NES.
// Возвращает (nil, nil), если точка занята и ignoreBlocked == false.
func (uc *TankLifecycleUseCases) SpawnEnemy(
	spawnIndex uint,
	ignoreBlocked bool,
	level uint,
) (*types.TankEntity, error) {
	if len(uc.spawnLayout.EnemySpawners) == 0 {
		return nil, fmt.Errorf("enemy spawners missing")
	}

	selectedIndex := int(spawnIndex) % len(uc.spawnLayout.EnemySpawners)
	spawnPosition := uc.spawnLayout.EnemySpawners[selectedIndex]

	if !ignoreBlocked && uc.isSpawnerBlocked(spawnPosition) {
		return nil, nil
	}

	tank, err := uc.spawnTank(
		types.DirectionDown,
		spawnPosition,
		types.TankRoleEnemy,
		level,
	)
	if err != nil {
		return nil, err
	}

	uc.tanksRepository.AddEnemy(&tank)

	return &tank, nil
}

// SpawnPlayer1 спавнит первого игрока с указанным уровнем звёзд
// (уровень переживает переход между этапами, но не гибель танка)
func (uc *TankLifecycleUseCases) SpawnPlayer1(
	level uint,
) (*types.TankEntity, error) {
	if uc.isSpawnerBlocked(uc.spawnLayout.Player1Spawner) {
		return nil, nil
	}
	tank, err := uc.spawnTank(
		types.DirectionUp,
		uc.spawnLayout.Player1Spawner,
		types.TankRolePlayer1,
		level,
	)
	if err != nil {
		return nil, err
	}

	tankPtr := new(types.TankEntity)
	*tankPtr = tank
	uc.tanksRepository.SetPlayer(types.PlayerTankNumPlayer1, tankPtr)
	return tankPtr, nil
}

// SpawnPlayer2 спавнит второго игрока с указанным уровнем звёзд
func (uc *TankLifecycleUseCases) SpawnPlayer2(
	level uint,
) (*types.TankEntity, error) {
	if uc.isSpawnerBlocked(uc.spawnLayout.Player2Spawner) {
		return nil, nil
	}
	tank, err := uc.spawnTank(
		types.DirectionUp,
		uc.spawnLayout.Player2Spawner,
		types.TankRolePlayer2,
		level,
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

func (uc *TankLifecycleUseCases) RemoveEnemy(tank *types.TankEntity) {
	uc.tanksRepository.RemoveEnemy(tank)
}

// spawnShieldTicks — щит игрока после появления (~3 секунды)
const spawnShieldTicks = 180

func (uc *TankLifecycleUseCases) finishSpawnAnimation(
	tank *types.TankEntity,
) {
	uc.renderUseCases.UpdateTankAnimation(tank)
	tank.State = types.TankStateStopped
	if !tank.IsEnemy() {
		tank.SetShieldTicks(spawnShieldTicks)
	}
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
