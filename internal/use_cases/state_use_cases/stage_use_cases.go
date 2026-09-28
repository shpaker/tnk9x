package state_use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"

	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

var _ interfaces.IStageUseCases = (*StageUseCases)(nil)

// initialEnemiesCount — сколько врагов выходит сразу при старте уровня
const initialEnemiesCount = 3

// coopExtraActiveEnemies — прибавка к лимиту активных врагов
// при игре вдвоём, как в NES
const coopExtraActiveEnemies = 2

type StageUseCases struct {
	tankLifecycleUseCases       interfaces.ITankLifecycleUseCases
	waveUseCases                interfaces.IWaveUseCases
	enemySpawnSelectionUseCases interfaces.IEnemySpawnSelectionUseCases
	tankCommonUseCases          interfaces.ITankCommonUseCases
	bulletUseCases              interfaces.IBulletUseCases
	collisionUseCases           interfaces.ICollisionUseCases
	hqUseCases                  interfaces.IHQUseCases

	stageSession *session_entities.StageSessionEntity

	bonusesRepository interfaces.IBonusesRepository
	mapUseCases       interfaces.IMapUseCases
	bonusUseCases     interfaces.IBonusUseCases
}

func NewStageUseCases(
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	waveUseCases interfaces.IWaveUseCases,
	enemySpawnSelectionUseCases interfaces.IEnemySpawnSelectionUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	bulletUseCases interfaces.IBulletUseCases,
	collisionUseCases interfaces.ICollisionUseCases,
	hqUseCases interfaces.IHQUseCases,
	stageSession *session_entities.StageSessionEntity,
	bonusesRepository interfaces.IBonusesRepository,
	mapUseCases interfaces.IMapUseCases,
	bonusUseCases interfaces.IBonusUseCases,
) *StageUseCases {
	return &StageUseCases{
		tankLifecycleUseCases:       tankLifecycleUseCases,
		waveUseCases:                waveUseCases,
		enemySpawnSelectionUseCases: enemySpawnSelectionUseCases,
		tankCommonUseCases:          tankCommonUseCases,
		bulletUseCases:              bulletUseCases,
		collisionUseCases:           collisionUseCases,
		hqUseCases:                  hqUseCases,
		stageSession:                stageSession,
		bonusesRepository:           bonusesRepository,
		mapUseCases:                 mapUseCases,
		bonusUseCases:               bonusUseCases,
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

// PlacePlayerTank реализует IStageUseCases: на старте уровня танк
// игрока сразу стоит на карте — со светом и зрением с первого кадра;
// возрождение после гибели по-прежнему идёт через анимацию появления
func (uc *StageUseCases) PlacePlayerTank(
	role types.TankRole,
) *types.TankEntity {
	tank := uc.SpawnPlayerTank(role)
	if tank == nil {
		return nil
	}
	uc.tankLifecycleUseCases.CompleteSpawn(tank)
	return tank
}

// SpawnInitialEnemyTanks выводит первых врагов сценария сразу
// на разные спаунеры по порядку, как в NES
func (uc *StageUseCases) SpawnInitialEnemyTanks() []*types.TankEntity {
	if uc.stageSession == nil {
		return nil
	}

	// Сбрасываем учёт уничтоженных врагов при спавне начальных врагов
	uc.stageSession.ClearDestroyedEnemiesTracking()

	count := min(
		initialEnemiesCount,
		int(uc.maxActiveEnemies()),
		uc.enemySpawnSelectionUseCases.GetSpawnersCount(),
	)

	result := make([]*types.TankEntity, 0, count)
	for spawnerIndex := 0; spawnerIndex < count; spawnerIndex++ {
		spawned := uc.spawnWaveTank(spawnerIndex)
		if spawned == nil {
			break
		}
		result = append(result, spawned)
	}

	return result
}

func (uc *StageUseCases) UpdateGameObjects(dt float64) {
	if uc.IsPaused() {
		return
	}

	uc.updateEnemySpawnCountdown()
	if uc.stageSession != nil {
		uc.stageSession.UpdateStageTicks()
	}

	if uc.bulletUseCases != nil {
		_ = uc.bulletUseCases.UpdateBullets(dt)
	}

	if uc.collisionUseCases != nil {
		uc.collisionUseCases.UpdateCollisions()
	}

	if uc.bonusUseCases != nil {
		uc.bonusUseCases.UpdateEffects()
	}

	uc.trackDestroyedEnemies()

	if uc.hqUseCases != nil {
		uc.hqUseCases.IsExplosionFinished(uc.hqUseCases.GetHQ())
	}
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
			role := types.PlayerTankNumToRole(num)
			respawned := uc.SpawnPlayerTank(role)

			if respawned == nil {
				// Спаунер занят: жизнь вернётся и спишется
				// при следующей попытке
				if !uc.stageSession.IsPlayerDefeated(num) {
					uc.stageSession.RestorePlayerLife(num)
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
	if uc.stageSession == nil {
		return nil
	}

	if uc.tankCommonUseCases != nil &&
		int(uc.maxActiveEnemies()) <= countActiveEnemies(
			uc.tankCommonUseCases.GetAllTanks(),
		) {
		return nil
	}

	if !uc.stageSession.CanSpawnNextEnemy() {
		return nil
	}

	spawnerIndex, ok := uc.enemySpawnSelectionUseCases.SelectSpawner(
		uc.stageSession,
	)
	if !ok {
		return nil
	}

	return uc.spawnWaveTank(spawnerIndex)
}

// spawnWaveTank выводит следующий танк сценария на спаунер;
// nil — волна ещё не началась или спаун не удался
func (uc *StageUseCases) spawnWaveTank(spawnerIndex int) *types.TankEntity {
	tank, ok := uc.waveUseCases.NextTank(uc.stageSession)
	if !ok {
		return nil
	}

	spawned, err := uc.tankLifecycleUseCases.SpawnEnemy(
		spawnerIndex,
		tank.Level,
	)
	if err != nil || spawned == nil {
		return nil
	}

	spawned.SetWithBonus(tank.HasBonus)
	uc.waveUseCases.CommitSpawn(uc.stageSession, spawnerIndex)
	return spawned
}

// maxActiveEnemies — лимит активных врагов уровня с учётом кооператива
func (uc *StageUseCases) maxActiveEnemies() uint {
	limit := uc.stageSession.GetMaxActiveEnemies()
	if uc.stageSession.GetPlayerCount() > 1 {
		limit += coopExtraActiveEnemies
	}
	return limit
}

func (uc *StageUseCases) updateEnemySpawnCountdown() {
	if uc.stageSession != nil {
		uc.stageSession.UpdateEnemySpawnCountdown()
	}
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

// GetStageResult — итог уровня: исход, потерянные жизни и время
func (uc *StageUseCases) GetStageResult() types.StageResult {
	if uc.stageSession == nil {
		return types.StageResult{}
	}
	return types.StageResult{
		Won:          uc.IsStageWon(),
		LivesLost:    uc.stageSession.GetPlayerDeaths(),
		ElapsedTicks: uc.stageSession.GetStageTicks(),
	}
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
