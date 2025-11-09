package stateusecases

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"

	"github.com/shpaker/gonflict/internal/types/session_entities"
)

// StageUseCases предоставляет бизнес-логику уровня (stage)
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
	stageSession *session_entities.StageSessionEntity

	// Отслеживание состояния врагов
	destroyedEnemies map[*types.TankEntity]struct{}
}

func NewStageUseCases(
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	bulletUseCases interfaces.IBulletUseCases,
	collisionUseCases interfaces.ICollisionUseCases,
	hqUseCases interfaces.IHQUseCases,
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
		stageSession:          stageSession,
		destroyedEnemies:      make(map[*types.TankEntity]struct{}),
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
	if uc.stageSession != nil && uc.stageSession.IsPlayer1Defeated() {
		return nil
	}

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
		uc.destroyedEnemies = make(map[*types.TankEntity]struct{})
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

	if uc.bulletUseCases != nil {
		_ = uc.bulletUseCases.UpdateBullets(dt)
	}

	if uc.collisionUseCases != nil {
		uc.collisionUseCases.UpdateCollisions()
	}

	uc.trackDestroyedEnemies()

	if uc.hqUseCases != nil {
		uc.hqUseCases.IsExplosionFinished(uc.hqUseCases.GetHQ())
	}
}

// --- Дополнительные операции ---

func (uc *StageUseCases) TogglePause() {
	uc.isPaused = !uc.isPaused
}

func (uc *StageUseCases) IsPaused() bool {
	return uc.isPaused
}

// TryRespawnPlayerTank пытается возродить игрока, если его танк уничтожен и есть оставшиеся жизни
func (uc *StageUseCases) TryRespawnPlayerTank() *types.TankEntity {
	if uc.stageSession == nil {
		return nil
	}

	playerTank := uc.getPlayerTank()
	if playerTank == nil || playerTank.State != types.TankStateExploded ||
		uc.stageSession.IsPlayer1Defeated() {
		return nil
	}

	uc.stageSession.DecrementPlayer1Lives()

	respawned := uc.SpawnPlayerTank()
	if respawned == nil {
		// Возвращаем жизнь, если возродить не удалось
		uc.stageSession.SetPlayer1Lives(
			uc.stageSession.GetPlayer1Lives() + 1,
		)
	}

	return respawned
}

// TrySpawnEnemy пытается создать нового врага при соблюдении условий респавна
func (uc *StageUseCases) TrySpawnEnemy() *types.TankEntity {
	if uc.tankLifecycleUseCases == nil {
		return nil
	}

	if uc.stageSession != nil &&
		uc.tankCommonUseCases != nil &&
		int(uc.stageSession.GetMaxActiveEnemies()) <= countActiveEnemies(
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

// IsStageWon возвращает true, если игрок выполнил условия победы (все враги уничтожены)
func (uc *StageUseCases) IsStageWon() bool {
	if uc.stageSession == nil {
		return false
	}

	if !uc.stageSession.AreAllEnemiesDefeated() {
		return false
	}

	if uc.stageSession.IsPlayer1Defeated() {
		return false
	}

	if uc.hqUseCases != nil {
		return !uc.hqUseCases.IsDestroyed()
	}

	return true
}

// IsStageLost возвращает true, если выполнено любое условие поражения (игрок погиб или база уничтожена)
func (uc *StageUseCases) IsStageLost() bool {
	if uc.stageSession != nil && uc.stageSession.IsPlayer1Defeated() {
		return true
	}

	if uc.hqUseCases != nil {
		return uc.hqUseCases.IsDestroyed()
	}

	return false
}

// IsStageFinished возвращает true, если уровень завершён победой или поражением
func (uc *StageUseCases) IsStageFinished() bool {
	return uc.IsStageWon() || uc.IsStageLost()
}

func (uc *StageUseCases) trackDestroyedEnemies() {
	if uc.stageSession == nil || uc.tankCommonUseCases == nil {
		return
	}

	enemies := uc.tankCommonUseCases.GetAllTanks()
	if len(enemies) == 0 {
		return
	}

	if uc.destroyedEnemies == nil {
		uc.destroyedEnemies = make(map[*types.TankEntity]struct{})
	}

	for _, tank := range enemies {
		if tank == nil || !tank.IsEnemy {
			continue
		}

		if tank.State == types.TankStateExploded {
			if _, exists := uc.destroyedEnemies[tank]; exists {
				continue
			}
			uc.stageSession.IncrementDestroyedEnemies()
			uc.destroyedEnemies[tank] = struct{}{}
		}
	}
}

func (uc *StageUseCases) getPlayerTank() *types.TankEntity {
	if uc.tankLifecycleUseCases == nil {
		return nil
	}
	return uc.tankLifecycleUseCases.GetPlayerTank()
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
