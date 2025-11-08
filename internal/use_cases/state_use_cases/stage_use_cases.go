package stateusecases

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

type StageUseCases struct {
	isPaused bool

	tankLifecycleUseCases interfaces.ITankLifecycleUseCases
	tankCommonUseCases    interfaces.ITankCommonUseCases
	bulletUseCases        interfaces.IBulletUseCases
	collisionUseCases     interfaces.ICollisionUseCases
	enemyInputAdapter     interfaces.IAiInputAdapter
	hqUseCases            interfaces.IHQUseCases
	hqEntity              *types.HQEntity
}

func NewStageUseCases(
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	bulletUseCases interfaces.IBulletUseCases,
	collisionUseCases interfaces.ICollisionUseCases,
	enemyInputAdapter interfaces.IAiInputAdapter,
	hqUseCases interfaces.IHQUseCases,
	hqEntity *types.HQEntity,
) StageUseCases {
	return StageUseCases{
		isPaused:              false,
		tankLifecycleUseCases: tankLifecycleUseCases,
		tankCommonUseCases:    tankCommonUseCases,
		bulletUseCases:        bulletUseCases,
		collisionUseCases:     collisionUseCases,
		enemyInputAdapter:     enemyInputAdapter,
		hqUseCases:            hqUseCases,
		hqEntity:              hqEntity,
	}
}

func (uc *StageUseCases) PauseStageState() {
	uc.isPaused = true
}

func (uc *StageUseCases) ResumeStageState() {
	uc.isPaused = false
}

func (uc *StageUseCases) SpawnPlayerTank() *types.TankEntity {
	if uc.tankLifecycleUseCases == nil {
		return nil
	}

	playerTank, err := uc.tankLifecycleUseCases.SpawnPlayer1()
	if err != nil || playerTank == nil {
		return nil
	}

	return playerTank
}

func (uc *StageUseCases) SpawnEnemyTanks() []*types.TankEntity {
	if uc.tankLifecycleUseCases == nil {
		return nil
	}

	enemies, err := uc.tankLifecycleUseCases.OnStageSetUpEnemiesSpawn()
	if err != nil {
		return nil
	}

	for _, enemy := range enemies {
		if enemy == nil {
			continue
		}
		if uc.enemyInputAdapter != nil {
			uc.enemyInputAdapter.AddTank(enemy)
		}
	}

	return enemies[:]
}

func (uc *StageUseCases) UpdateGameObjects(dt float64) {
	if uc.isPaused {
		return
	}

	if uc.tankCommonUseCases != nil {
		if err := uc.tankCommonUseCases.UpdateAllTanks(dt); err != nil {
			_ = err
		}
	}

	if uc.bulletUseCases != nil {
		_ = uc.bulletUseCases.UpdateBullets(dt)
	}

	if uc.collisionUseCases != nil {
		uc.collisionUseCases.UpdateCollisions()
	}

	if uc.enemyInputAdapter != nil {
		uc.enemyInputAdapter.Update(dt)
	}

	if uc.hqUseCases != nil {
		uc.hqUseCases.IsExplosionFinished(uc.hqEntity)
	}
}

func (uc *StageUseCases) TogglePause() {
	uc.isPaused = !uc.isPaused
}

func (uc *StageUseCases) IsPaused() bool {
	return uc.isPaused
}
