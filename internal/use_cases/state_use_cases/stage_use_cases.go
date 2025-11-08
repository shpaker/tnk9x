package stateusecases

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"

	"github.com/shpaker/gonflict/internal/types/session_entities"
)

type StageUseCases struct {
	// Служебное состояние
	isPaused bool

	// Use Cases
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases
	tankCommonUseCases    interfaces.ITankCommonUseCases
	bulletUseCases        interfaces.IBulletUseCases
	collisionUseCases     interfaces.ICollisionUseCases
	hqUseCases            interfaces.IHQUseCases

	// Сущности
	hqEntity     *types.HQEntity
	stageSession *session_entities.StageSessionEntity
}

func NewStageUseCases(
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	bulletUseCases interfaces.IBulletUseCases,
	collisionUseCases interfaces.ICollisionUseCases,
	hqUseCases interfaces.IHQUseCases,
	hqEntity *types.HQEntity,
	stageSession *session_entities.StageSessionEntity,
	enemyRespawnDelay uint,
) StageUseCases {
	if enemyRespawnDelay == 0 {
		enemyRespawnDelay = 3 * 60
	}
	if stageSession != nil {
		stageSession.SetEnemyRespawnDelay(enemyRespawnDelay)
	}
	return StageUseCases{
		isPaused:              false,
		tankLifecycleUseCases: tankLifecycleUseCases,
		tankCommonUseCases:    tankCommonUseCases,
		bulletUseCases:        bulletUseCases,
		collisionUseCases:     collisionUseCases,
		hqUseCases:            hqUseCases,
		hqEntity:              hqEntity,
		stageSession:          stageSession,
	}
}

// --- Управление паузой ---

func (uc *StageUseCases) PauseStageState() {
	uc.isPaused = true
}

func (uc *StageUseCases) ResumeStageState() {
	uc.isPaused = false
}

// --- Спавн сущностей ---

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

func (uc *StageUseCases) SpawnInitialEnemyTanks() []*types.TankEntity {
	if uc.tankLifecycleUseCases == nil {
		return nil
	}

	if uc.stageSession != nil {
		uc.stageSession.Reset()
	}

	enemies, err := uc.tankLifecycleUseCases.OnStageSetUpEnemiesSpawn()
	if err != nil {
		return nil
	}

	result := make([]*types.TankEntity, 0, len(enemies))
	for _, enemy := range enemies {
		if enemy == nil {
			continue
		}
		uc.registerEnemySpawned()
		result = append(result, enemy)
	}

	return result
}

// --- Основной игровой цикл ---

func (uc *StageUseCases) UpdateGameObjects(dt float64) {
	if uc.isPaused {
		return
	}

	uc.updateEnemySpawnCountdown()

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

	if uc.hqUseCases != nil {
		uc.hqUseCases.IsExplosionFinished(uc.hqEntity)
	}
}

// --- Дополнительные операции ---

func (uc *StageUseCases) TogglePause() {
	uc.isPaused = !uc.isPaused
}

func (uc *StageUseCases) IsPaused() bool {
	return uc.isPaused
}

// TrySpawnEnemy пытается создать нового врага при соблюдении условий респавна
func (uc *StageUseCases) TrySpawnEnemy() *types.TankEntity {
	if uc.tankLifecycleUseCases == nil {
		return nil
	}

	if uc.stageSession != nil &&
		uc.tankCommonUseCases != nil &&
		int(uc.stageSession.MaxActiveEnemies()) <= countActiveEnemies(
			uc.tankCommonUseCases.GetAllTanks(),
		) {
		return nil
	}

	if !uc.canSpawnEnemy() {
		return nil
	}

	spawned, err := uc.tankLifecycleUseCases.SpawnEnemy(nil, false)
	if err != nil || spawned == nil {
		return nil
	}

	uc.registerEnemySpawned()
	return spawned
}

func (uc *StageUseCases) updateEnemySpawnCountdown() {
	if uc.stageSession != nil {
		uc.stageSession.UpdateEnemySpawnCountdown()
	}
}

func (uc *StageUseCases) canSpawnEnemy() bool {
	if uc.stageSession == nil {
		return false
	}
	return uc.stageSession.CanSpawnNextEnemy()
}

func (uc *StageUseCases) registerEnemySpawned() {
	if uc.stageSession != nil {
		uc.stageSession.RegisterEnemySpawned()
	}
}

func countActiveEnemies(tanks []*types.TankEntity) int {
	total := 0
	for _, tank := range tanks {
		if tank != nil && tank.IsEnemy && tank.IsActive() {
			total++
		}
	}
	return total
}
