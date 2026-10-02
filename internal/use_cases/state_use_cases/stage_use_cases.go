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

// SpawnPlayerTank реализует IStageUseCases: возрождение танка игрока
// всегда с нулевой прокачкой
func (uc *StageUseCases) SpawnPlayerTank(
	role types.TankRole,
) *types.TankEntity {
	return uc.spawnPlayerTank(role, 0)
}

func (uc *StageUseCases) spawnPlayerTank(
	role types.TankRole,
	level uint,
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
		playerTank, err = uc.tankLifecycleUseCases.SpawnPlayer1(level)
	case types.TankRolePlayer2:
		playerTank, err = uc.tankLifecycleUseCases.SpawnPlayer2(level)
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
// возрождение после гибели по-прежнему идёт через анимацию появления.
// Прокачка — из снимка прошлого уровня при переносе
func (uc *StageUseCases) PlacePlayerTank(
	role types.TankRole,
) *types.TankEntity {
	if uc.stageSession == nil {
		return nil
	}
	num := types.RoleToPlayerTankNum(role)
	tank := uc.spawnPlayerTank(role, uc.stageSession.GetStartTier(num))
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
		CarriedOver:  uc.stageSession.IsCarryOver(),
	}
}

// SaveCarryOver реализует IStageUseCases: жизни и прокачка танков
// игроков на момент победы
func (uc *StageUseCases) SaveCarryOver() {
	if uc.stageSession == nil {
		return
	}
	playersTanks := uc.GetPlayersTanks()
	for i := 0; i < int(uc.stageSession.GetPlayerCount()) && i < len(playersTanks); i++ {
		num := types.PlayerTankNum(i)
		lives := uc.stageSession.GetPlayerLives(num)
		var tier uint
		tank := playersTanks[i]
		switch {
		case tank == nil:
		case tank.IsDestroyed():
			// Жизнь списывается при возрождении, которого уже не будет
			if lives > 0 {
				lives--
			}
		case tank.GetSpecs() != nil:
			tier = tank.GetSpecs().GetLevel()
		}
		uc.stageSession.SaveCarryOver(num, lives, tier)
	}
}

// BoostCarryOver реализует IStageUseCases
func (uc *StageUseCases) BoostCarryOver() {
	uc.stageSession.BoostCarryOver()
}

// CanRevivePlayers реализует IStageUseCases: второй шанс — только
// при поражении потерей жизней. Штаб должен быть цел, а не просто
// не разрушен: взрывающийся на паузе штаб так и остался бы
// взрывающимся. Без оставшихся врагов возрождение дало бы победу
func (uc *StageUseCases) CanRevivePlayers() bool {
	if uc.stageSession.IsReviveUsed() || !uc.IsStageLost() {
		return false
	}
	hq := uc.hqUseCases.GetHQ()
	return hq != nil && hq.IsIntact() &&
		uc.stageSession.GetRemainingEnemies() > 0
}

// RevivePlayers реализует IStageUseCases
func (uc *StageUseCases) RevivePlayers() {
	uc.stageSession.RevivePlayers()
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

		// Бонус не должен лечь на штаб
		if hq := uc.hqUseCases.GetHQ(); hq != nil &&
			overlaps(bonusCandidate, hq) {
			continue
		}

		// Проверяем коллизии с блоками
		blocks := uc.mapUseCases.GetBlocks()
		hasBlockCollision := false
		for _, block := range blocks {
			if block == nil {
				continue
			}
			if overlaps(bonusCandidate, block) {
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

// overlaps проверяет пересечение прямоугольников бонуса и препятствия
func overlaps(bonus *types.BonusEntity, obstacle types.IEntityCollider) bool {
	bonusPos := bonus.GetPosition()
	bonusSize := bonus.GetSize()
	obstaclePos := obstacle.GetPosition()
	obstacleSize := obstacle.GetSize()

	return bonusPos.X < obstaclePos.X+float64(obstacleSize.Width) &&
		bonusPos.X+float64(bonusSize.Width) > obstaclePos.X &&
		bonusPos.Y < obstaclePos.Y+float64(obstacleSize.Height) &&
		bonusPos.Y+float64(bonusSize.Height) > obstaclePos.Y
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
