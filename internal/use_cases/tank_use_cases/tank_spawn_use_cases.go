package tank_use_cases

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// Убеждаемся, что TankSpawnUseCases реализует интерфейс ITankSpawnUseCases
var _ interfaces.ITankSpawnUseCases = (*TankSpawnUseCases)(nil)

// TankSpawnUseCases отвечает за спавн танков на этапе подготовки уровня
type TankSpawnUseCases struct {
	tanksRepository       interfaces.ITanksRepository
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases
	enemySpawnPositions   [3]types.Position
}

// NewTankSpawnUseCases создает новый экземпляр TankSpawnUseCases
func NewTankSpawnUseCases(
	tanksRepository interfaces.ITanksRepository,
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	enemySpawnPositions [3]types.Position,
) *TankSpawnUseCases {
	return &TankSpawnUseCases{
		tanksRepository:       tanksRepository,
		tankLifecycleUseCases: tankLifecycleUseCases,
		enemySpawnPositions:   enemySpawnPositions,
	}
}

// StageSetUp подготавливает танки уровня: игрока и трех врагов
func (uc *TankSpawnUseCases) StageSetUp() error {
	if uc.tankLifecycleUseCases == nil {
		return nil
	}

	if err := uc.spawnPlayerTank(); err != nil {
		return err
	}

	if err := uc.spawnEnemyTanks(); err != nil {
		return err
	}

	return nil
}

// spawnPlayerTank спавнит танк игрока, если он существует
func (uc *TankSpawnUseCases) spawnPlayerTank() error {
	if uc.tanksRepository == nil {
		return nil
	}

	playerTank := uc.tanksRepository.GetPlayer()
	if playerTank == nil {
		return nil
	}

	return uc.tankLifecycleUseCases.Spawn(playerTank)
}

// spawnEnemyTanks спавнит до трех вражеских танков
func (uc *TankSpawnUseCases) spawnEnemyTanks() error {
	if uc.tanksRepository == nil {
		return nil
	}

	enemyTanks := uc.tanksRepository.GetAllEnemies()
	for index, enemyTank := range enemyTanks {
		if index >= 3 {
			break
		}

		if enemyTank == nil {
			continue
		}

		enemyTank.Position = uc.enemySpawnPositions[index]

		if err := uc.tankLifecycleUseCases.Spawn(enemyTank); err != nil {
			return err
		}
	}

	return nil
}
