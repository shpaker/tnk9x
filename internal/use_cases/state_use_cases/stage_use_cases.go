package state_use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"

	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

var _ interfaces.IStageUseCases = (*StageUseCases)(nil)

type StageUseCases struct {
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases
	tankCommonUseCases    interfaces.ITankCommonUseCases
	bulletUseCases        interfaces.IBulletUseCases
	collisionUseCases     interfaces.ICollisionUseCases
	hqUseCases            interfaces.IHQUseCases

	stageSession *session_entities.StageSessionEntity

	bonusesRepository interfaces.IBonusesRepository
	fortressUseCases  interfaces.IFortressUseCases
	soundUseCases     interfaces.ISoundUseCases
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
	fortressUseCases interfaces.IFortressUseCases,
	soundUseCases interfaces.ISoundUseCases,
) *StageUseCases {
	if enemyRespawnDelay == 0 {
		enemyRespawnDelay = 3 * 60
	}
	if stageSession != nil {
		stageSession.SetEnemyRespawnDelay(enemyRespawnDelay)
	}
	return &StageUseCases{
		tankLifecycleUseCases: tankLifecycleUseCases,
		tankCommonUseCases:    tankCommonUseCases,
		bulletUseCases:        bulletUseCases,
		collisionUseCases:     collisionUseCases,
		hqUseCases:            hqUseCases,
		stageSession:          stageSession,
		bonusesRepository:     bonusesRepository,
		fortressUseCases:      fortressUseCases,
		soundUseCases:         soundUseCases,
	}
}

func (uc *StageUseCases) PauseStageState() {
	if uc.stageSession != nil {
		uc.stageSession.SetPaused(true)
	}
}

func (uc *StageUseCases) ResumeStageState() {
	if uc.stageSession != nil {
		uc.stageSession.SetPaused(false)
	}
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

	// Уровень звёзд переживает переход между этапами
	starLevel := uint(0)
	if run := uc.stageSession.RunSession(); run != nil {
		starLevel = run.GetStarLevel(num)
	}

	var playerTank *types.TankEntity
	var err error

	switch role {
	case types.TankRolePlayer1:
		playerTank, err = uc.tankLifecycleUseCases.SpawnPlayer1(starLevel)
	case types.TankRolePlayer2:
		playerTank, err = uc.tankLifecycleUseCases.SpawnPlayer2(starLevel)
	default:
		return nil
	}

	if err != nil || playerTank == nil {
		return nil
	}

	return playerTank
}

// initialEnemiesCount — сколько врагов появляется на старте этапа
const initialEnemiesCount = 3

func (uc *StageUseCases) SpawnInitialEnemyTanks() []*types.TankEntity {
	if uc.tankLifecycleUseCases == nil || uc.stageSession == nil {
		return nil
	}

	// Сбрасываем учёт уничтоженных врагов при спавне начальных врагов
	uc.stageSession.ClearDestroyedEnemiesTracking()

	result := make([]*types.TankEntity, 0, initialEnemiesCount)
	for i := 0; i < initialEnemiesCount; i++ {
		if uc.stageSession.EnemiesForSpawnCount() == 0 {
			break
		}

		enemy, err := uc.tankLifecycleUseCases.SpawnEnemy(
			uc.stageSession.NextEnemySpawnIndex(),
			true,
			uc.stageSession.NextEnemyTier(),
		)
		if err != nil || enemy == nil {
			continue
		}

		enemyNumber := uc.getNextEnemyNumber()
		if uc.shouldHaveBonus(enemyNumber) {
			enemy.SetWithBonus(true)
			uc.clearFieldBonuses()
		}
		uc.registerEnemySpawned()
		result = append(result, enemy)
	}

	return result
}

func (uc *StageUseCases) UpdateGameObjects(dt float64) {
	if uc.IsPaused() {
		return
	}

	uc.updateEnemySpawnCountdown()

	if uc.stageSession != nil {
		uc.stageSession.IncrementStageTicks()
		uc.stageSession.TickEnemyFreeze()
	}
	if uc.fortressUseCases != nil {
		uc.fortressUseCases.Update()
	}

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

func (uc *StageUseCases) AreEnemiesFrozen() bool {
	return uc.stageSession != nil && uc.stageSession.AreEnemiesFrozen()
}

func (uc *StageUseCases) TogglePause() {
	if uc.stageSession != nil {
		uc.stageSession.SetPaused(!uc.stageSession.IsPaused())
	}
}

func (uc *StageUseCases) IsPaused() bool {
	return uc.stageSession != nil && uc.stageSession.IsPaused()
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
			// Гибель танка сбрасывает накопленные звёзды
			if run := uc.stageSession.RunSession(); run != nil {
				run.SetStarLevel(num, 0)
			}
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

	spawned, err := uc.tankLifecycleUseCases.SpawnEnemy(
		uc.stageSession.NextEnemySpawnIndex(),
		false,
		uc.stageSession.NextEnemyTier(),
	)
	if err != nil || spawned == nil {
		return nil
	}

	enemyNumber := uc.getNextEnemyNumber()
	if uc.shouldHaveBonus(enemyNumber) {
		spawned.SetWithBonus(true)
		uc.clearFieldBonuses()
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

// shouldHaveBonus — мигающими с бонусом становятся враги №4, №11 и №18,
// как в оригинальной Battle City
func (uc *StageUseCases) shouldHaveBonus(enemyNumber uint) bool {
	return enemyNumber == 4 || enemyNumber == 11 || enemyNumber == 18
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

	for _, tank := range enemies {
		if tank == nil || !tank.IsEnemy() {
			continue
		}

		if tank.State == types.TankStateExploded {
			if !uc.stageSession.TrackDestroyedEnemy(tank) {
				continue
			}

			uc.awardEnemyKill(tank)

			// Учтённый взорванный враг больше не нужен: убираем из
			// репозитория, чтобы не накапливать мёртвые сущности
			if uc.tankLifecycleUseCases != nil {
				uc.tankLifecycleUseCases.RemoveEnemy(tank)
			}
		}
	}
}

// awardEnemyKill начисляет очки автору добивающего выстрела;
// враги без атрибуции (граната) очков не приносят, как в NES
func (uc *StageUseCases) awardEnemyKill(tank *types.TankEntity) {
	if uc.stageSession == nil {
		return
	}
	run := uc.stageSession.RunSession()
	if run == nil {
		return
	}

	role, hasDestroyer := tank.GetDestroyedBy()
	if !hasDestroyer {
		return
	}

	tier := uint(0)
	if specs := tank.GetSpecs(); specs != nil {
		tier = specs.GetLevel()
	}

	num := types.RoleToPlayerTankNum(role)
	if run.AddEnemyKill(num, tier) && uc.soundUseCases != nil {
		// Дополнительная жизнь за порог очков
		uc.soundUseCases.RequestSound(types.SoundIDBonus, false)
	}
}

// clearFieldBonuses убирает лежащий на поле бонус: появление нового
// мигающего танка отменяет предыдущий бонус, как в оригинале
func (uc *StageUseCases) clearFieldBonuses() {
	if uc.bonusesRepository != nil {
		uc.bonusesRepository.RemoveBonusesWithoutOwner()
	}
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
