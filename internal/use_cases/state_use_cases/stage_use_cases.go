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

// SpawnPlayerTank спавнит танк игрока с указанной ролью
func (uc *StageUseCases) SpawnPlayerTank(
	role types.TankRole,
) *types.TankEntity {
	if uc.stageSession == nil {
		return nil
	}

	// Проверяем, не проиграл ли игрок
	num := types.RoleToPlayerTankNum(role)
	if uc.stageSession.IsPlayerDefeated(num) {
		return nil
	}

	if uc.tankLifecycleUseCases == nil {
		return nil
	}

	var playerTank *types.TankEntity
	var err error

	switch role {
	case types.TankRolePlayer1:
		playerTank, err = uc.tankLifecycleUseCases.SpawnPlayer1()
	case types.TankRolePlayer2:
		playerTank, err = uc.tankLifecycleUseCases.SpawnPlayer2()
	default:
		return nil
	}

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

// TryRespawnPlayersTanks пытается возродить всех игроков, если их танки уничтожены и есть оставшиеся жизни
// Возвращает респавненных игроков (player1, player2) или (nil, nil) если никого не удалось респавнить
func (uc *StageUseCases) TryRespawnPlayersTanks() (*types.TankEntity, *types.TankEntity) {
	if uc.stageSession == nil {
		return nil, nil
	}

	// Определяем количество игроков
	playerCount := int(uc.stageSession.GetPlayerCount())
	if playerCount < 1 {
		playerCount = 1
	}
	if playerCount > 2 {
		playerCount = 2
	}

	playersTanks := uc.GetPlayersTanks()
	var respawned1, respawned2 *types.TankEntity

	// Пытаемся респавнить игроков в зависимости от выбранного количества
	for i := 0; i < playerCount; i++ {
		num := types.PlayerTankNum(i)
		playerTank := playersTanks[i]

		if playerTank != nil && playerTank.State == types.TankStateExploded &&
			!uc.stageSession.IsPlayerDefeated(num) {
			uc.stageSession.DecrementPlayerLives(num)
			role := types.PlayerTankNumToRole(num)
			respawned := uc.SpawnPlayerTank(role)

			if respawned == nil {
				// Возвращаем жизнь, если возродить не удалось И жизни еще остались
				if !uc.stageSession.IsPlayerDefeated(num) {
					uc.stageSession.SetPlayerLives(
						num,
						uc.stageSession.GetPlayerLives(num)+1,
					)
				}
			} else {
				// Сохраняем респавненного игрока для возврата
				if num == types.PlayerTankNumPlayer1 {
					respawned1 = respawned
				} else if num == types.PlayerTankNumPlayer2 {
					respawned2 = respawned
				}
			}
		}
	}

	return respawned1, respawned2
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

	// Проверяем, что хотя бы один игрок не проиграл (через итерацию)
	hasActivePlayer := false
	for i := 0; i < 2; i++ {
		num := types.PlayerTankNum(i)
		if !uc.stageSession.IsPlayerDefeated(num) {
			hasActivePlayer = true
			break
		}
	}

	if !hasActivePlayer {
		return false
	}

	if uc.hqUseCases != nil {
		return !uc.hqUseCases.IsDestroyed()
	}

	return true
}

// IsStageLost возвращает true, если выполнено любое условие поражения (игрок погиб или база уничтожена)
func (uc *StageUseCases) IsStageLost() bool {
	if uc.hqUseCases != nil && uc.hqUseCases.IsDestroyed() {
		return true
	}

	if uc.stageSession == nil {
		return false
	}

	// Проверяем, что все игроки проиграли через итерацию
	allPlayersDefeated := true
	for i := 0; i < 2; i++ {
		num := types.PlayerTankNum(i)
		if !uc.stageSession.IsPlayerDefeated(num) {
			allPlayersDefeated = false
			break
		}
	}

	return allPlayersDefeated
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
		if tank == nil || !tank.IsEnemy() {
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

// GetPlayersTanks возвращает танки всех игроков через массив
func (uc *StageUseCases) GetPlayersTanks() []*types.TankEntity {
	if uc.tankLifecycleUseCases == nil {
		return make([]*types.TankEntity, 2)
	}

	playersTanks := make([]*types.TankEntity, 2)

	// Получаем танки всех игроков через итерацию
	for i := 0; i < 2; i++ {
		num := types.PlayerTankNum(i)
		playersTanks[i] = uc.tankLifecycleUseCases.GetPlayerTank(num)
	}

	return playersTanks
}

func countActiveEnemies(tanks []*types.TankEntity) int {
	total := 0
	for _, tank := range tanks {
		if tank != nil && tank.IsEnemy() && tank.IsActive() {
			total++
		}
	}
	return total
}
