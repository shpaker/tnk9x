package stateusecases

import (
	"github.com/shpaker/tnk25/internal/interfaces"
	"github.com/shpaker/tnk25/internal/types"
	"github.com/shpaker/tnk25/internal/use_cases"

	"github.com/shpaker/tnk25/internal/types/session_entities"
)

type StageUseCases struct {
	isPaused bool

	tankLifecycleUseCases interfaces.ITankLifecycleUseCases
	tankCommonUseCases    interfaces.ITankCommonUseCases
	bulletUseCases        interfaces.IBulletUseCases
	collisionUseCases     interfaces.ICollisionUseCases
	hqUseCases            interfaces.IHQUseCases

	stageSession *session_entities.StageSessionEntity

	destroyedEnemies map[*types.TankEntity]struct{}

	bonusesRepository interfaces.IBonusesRepository
	mapUseCases       interfaces.IMapUseCases
	bonusUseCases     *use_cases.BonusUseCases
}

func NewStageUseCases(
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	bulletUseCases interfaces.IBulletUseCases,
	collisionUseCases interfaces.ICollisionUseCases,
	hqUseCases interfaces.IHQUseCases,
	stageSession *session_entities.StageSessionEntity,
	enemyRespawnDelay uint,
	bonusesRepository interfaces.IBonusesRepository,
	mapUseCases interfaces.IMapUseCases,
	bonusUseCases *use_cases.BonusUseCases,
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
		bonusesRepository:     bonusesRepository,
		mapUseCases:           mapUseCases,
		bonusUseCases:         bonusUseCases,
	}
}

func (uc *StageUseCases) PauseStageState() {
	uc.isPaused = true
}

func (uc *StageUseCases) ResumeStageState() {
	uc.isPaused = false
}

func (uc *StageUseCases) SpawnPlayerTank(
	role types.TankRole,
) *types.TankEntity {
	if uc.stageSession == nil {
		return nil
	}

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

	// Очищаем карту уничтоженных врагов при спавне начальных врагов
	uc.destroyedEnemies = make(map[*types.TankEntity]struct{})

	enemies, err := uc.tankLifecycleUseCases.OnStageSetUpEnemiesSpawn()
	if err != nil {
		return nil
	}

	result := make([]*types.TankEntity, 0, len(enemies))
	for _, enemy := range enemies {
		if enemy == nil {
			continue
		}
		enemyNumber := uc.getNextEnemyNumber()
		if uc.shouldHaveBonus(enemyNumber) {
			enemy.SetWithBonus(true)
		}
		uc.registerEnemySpawned()
		result = append(result, enemy)
	}

	return result
}

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

func (uc *StageUseCases) TogglePause() {
	uc.isPaused = !uc.isPaused
}

func (uc *StageUseCases) IsPaused() bool {
	return uc.isPaused
}

func (uc *StageUseCases) TryRespawnPlayersTanks() (*types.TankEntity, *types.TankEntity) {
	if uc.stageSession == nil {
		return nil, nil
	}

	playerCount := int(uc.stageSession.GetPlayerCount())
	if playerCount < 1 {
		playerCount = 1
	}
	if playerCount > 2 {
		playerCount = 2
	}

	playersTanks := uc.GetPlayersTanks()
	var respawned1, respawned2 *types.TankEntity

	for i := 0; i < playerCount; i++ {
		num := types.PlayerTankNum(i)
		playerTank := playersTanks[i]

		if playerTank != nil && playerTank.State == types.TankStateExploded &&
			!uc.stageSession.IsPlayerDefeated(num) {
			uc.stageSession.DecrementPlayerLives(num)
			role := types.PlayerTankNumToRole(num)
			respawned := uc.SpawnPlayerTank(role)

			if respawned == nil {
				if !uc.stageSession.IsPlayerDefeated(num) {
					uc.stageSession.SetPlayerLives(
						num,
						uc.stageSession.GetPlayerLives(num)+1,
					)
				}
			} else {
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

	// Получаем количество оставшихся врагов для определения уровня
	remainingEnemies := uint(0)
	if uc.stageSession != nil {
		remainingEnemies = uc.stageSession.GetRemainingEnemies()
	}

	spawned, err := uc.tankLifecycleUseCases.SpawnEnemyWithLevel(
		nil,
		false,
		remainingEnemies,
	)
	if err != nil || spawned == nil {
		return nil
	}

	enemyNumber := uc.getNextEnemyNumber()
	if uc.shouldHaveBonus(enemyNumber) {
		spawned.SetWithBonus(true)
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

func (uc *StageUseCases) getNextEnemyNumber() uint {
	if uc.stageSession != nil {
		return uc.stageSession.GetNextEnemyNumber()
	}
	return 0
}

// shouldHaveBonus проверяет, должен ли враг с данным номером иметь бонус
// Логика: первый враг с бонусом = 4, затем каждый следующий = предыдущий + (4+i),
// где i увеличивается на 1 после каждого спауна с бонусом
// Последовательность: 4, 4+5=9, 9+6=15, 15+7=22, ...
func (uc *StageUseCases) shouldHaveBonus(enemyNumber uint) bool {
	if enemyNumber < 4 {
		return false
	}
	if enemyNumber == 4 {
		return true
	}

	// Вычисляем последовательность номеров с бонусом
	// Первый: 4
	// Второй: 4 + 5 = 9 (где 5 = 4+1)
	// Третий: 9 + 6 = 15 (где 6 = 4+2)
	// Четвертый: 15 + 7 = 22 (где 7 = 4+3)
	// И так далее...
	currentBonusNumber := uint(4)
	i := uint(1)
	for currentBonusNumber < enemyNumber {
		currentBonusNumber = currentBonusNumber + (4 + i)
		i++
	}
	return currentBonusNumber == enemyNumber
}

func (uc *StageUseCases) IsStageWon() bool {
	if uc.stageSession == nil {
		return false
	}

	if !uc.stageSession.AreAllEnemiesDefeated() {
		return false
	}

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

func (uc *StageUseCases) IsStageLost() bool {
	if uc.hqUseCases != nil && uc.hqUseCases.IsDestroyed() {
		return true
	}

	if uc.stageSession == nil {
		return false
	}

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

			// Если враг с бонусом уничтожен, удаляем все бонусы без owner'а и спавним новый
			if tank.GetWithBonus() {
				uc.handleEnemyWithBonusDestroyed()
			}
		}
	}
}

func (uc *StageUseCases) handleEnemyWithBonusDestroyed() {
	// Удаляем все бонусы без owner'а
	if uc.bonusesRepository != nil {
		uc.bonusesRepository.RemoveBonusesWithoutOwner()
	}

	// Спавним новый рандомный бонус используя BonusUseCases.SpawnRandomBonusEntity
	if uc.mapUseCases == nil || uc.bonusUseCases == nil {
		return
	}

	// Получаем размер базового тайла для проверки коллизий
	baseSizePx := uint(16)
	bonusSize := types.Size{
		Width:  int(baseSizePx),
		Height: int(baseSizePx),
	}

	const maxAttempts = 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Получаем случайную позицию для спавна бонуса
		position := uc.mapUseCases.GetRandomBonusSpawnPosition()

		// Создаем временную сущность бонуса для проверки коллизий
		bonusCandidate := types.NewBonusEntity(
			types.BonusTypeGrenade, // Тип не важен для проверки коллизий
			position,
			bonusSize,
			nil, // Изображение не нужно для проверки коллизий
		)

		// Проверяем коллизии с танками используя collisionUseCases
		hasTankCollision := false
		if uc.collisionUseCases != nil {
			hasTankCollision = uc.collisionUseCases.IsSpawnerBlocked(
				types.Position{
					X: position.X / float64(bonusSize.Width),
					Y: position.Y / float64(bonusSize.Height),
				},
				bonusSize,
			)
		}

		if hasTankCollision {
			continue
		}

		// Проверяем коллизии с блоками
		blocks := uc.mapUseCases.GetBlocks()
		hasBlockCollision := false
		// Используем collisionUseCases для проверки коллизий с блоками
		// Для этого нужно проверить каждый блок
		for _, block := range blocks {
			if block == nil {
				continue
			}
			// Используем простую проверку пересечения прямоугольников
			if uc.checkBonusBlockCollision(bonusCandidate, block) {
				hasBlockCollision = true
				break
			}
		}

		if !hasBlockCollision {
			// Используем BonusUseCases для создания бонуса
			bonus := uc.bonusUseCases.SpawnRandomBonusEntity(position)
			if bonus != nil && uc.bonusesRepository != nil {
				uc.bonusesRepository.AddBonus(bonus)
				return
			}
		}
	}
}

// checkBonusBlockCollision проверяет коллизию бонуса с блоком
func (uc *StageUseCases) checkBonusBlockCollision(
	bonus *types.BonusEntity,
	block *types.BlockEntity,
) bool {
	if bonus == nil || block == nil {
		return false
	}

	bonusPos := bonus.GetPosition()
	bonusSize := bonus.GetSize()
	blockPos := block.GetPosition()
	blockSize := block.GetSize()

	// Простая проверка пересечения прямоугольников
	return bonusPos.X < blockPos.X+float64(blockSize.Width) &&
		bonusPos.X+float64(bonusSize.Width) > blockPos.X &&
		bonusPos.Y < blockPos.Y+float64(blockSize.Height) &&
		bonusPos.Y+float64(bonusSize.Height) > blockPos.Y
}

func (uc *StageUseCases) GetPlayersTanks() []*types.TankEntity {
	if uc.tankLifecycleUseCases == nil {
		return make([]*types.TankEntity, 2)
	}

	playersTanks := make([]*types.TankEntity, 2)

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
