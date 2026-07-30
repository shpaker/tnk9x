package use_cases

import (
	"math/rand"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

var _ interfaces.IBonusUseCases = (*BonusUseCases)(nil)

// Тайминги бонусов, как в оригинале
const (
	// helmetShieldTicks — щит от бонуса «шлем» (~10 секунд)
	helmetShieldTicks = 600

	// timerFreezeTicks — заморозка врагов бонусом «таймер» (~10 секунд)
	timerFreezeTicks = 600
)

// bonusFieldSpawnAttempts — попыток найти свободную клетку для бонуса
const bonusFieldSpawnAttempts = 3

type BonusUseCases struct {
	tankCommonUseCases    interfaces.ITankCommonUseCases
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases
	stageSession          *session_entities.StageSessionEntity
	bonusesRepository     interfaces.IBonusesRepository
	configProvider        interfaces.IConfigProvider
	tilesUseCases         interfaces.ITilesUseCases
	renderUseCases        interfaces.IRenderUseCases
	soundUseCases         interfaces.ISoundUseCases
	mapUseCases           interfaces.IMapUseCases
	spawnCollisionService interfaces.ISpawnCollisionService
	fortressUseCases      interfaces.IFortressUseCases
}

func NewBonusUseCases(
	tankCommonUseCases interfaces.ITankCommonUseCases,
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	stageSession *session_entities.StageSessionEntity,
	bonusesRepository interfaces.IBonusesRepository,
	configProvider interfaces.IConfigProvider,
	tilesUseCases interfaces.ITilesUseCases,
	renderUseCases interfaces.IRenderUseCases,
	soundUseCases interfaces.ISoundUseCases,
	mapUseCases interfaces.IMapUseCases,
	spawnCollisionService interfaces.ISpawnCollisionService,
	fortressUseCases interfaces.IFortressUseCases,
) *BonusUseCases {
	return &BonusUseCases{
		tankCommonUseCases:    tankCommonUseCases,
		tankLifecycleUseCases: tankLifecycleUseCases,
		stageSession:          stageSession,
		bonusesRepository:     bonusesRepository,
		configProvider:        configProvider,
		tilesUseCases:         tilesUseCases,
		renderUseCases:        renderUseCases,
		soundUseCases:         soundUseCases,
		mapUseCases:           mapUseCases,
		spawnCollisionService: spawnCollisionService,
		fortressUseCases:      fortressUseCases,
	}
}

func (uc *BonusUseCases) Apply(
	bonus *types.BonusEntity,
	tank *types.TankEntity,
) {
	if bonus == nil || tank == nil {
		return
	}

	uc.soundUseCases.RequestSound(types.SoundIDBonus, false)
	uc.awardBonusPoints(tank)

	switch bonus.GetType() {
	case types.BonusTypeHelmet:
		uc.applyHelmet(tank)
	case types.BonusTypeTimer:
		uc.applyTimer(tank)
	case types.BonusTypeShovel:
		uc.applyShovel(tank)
	case types.BonusTypeStar:
		uc.applyStar(tank)
	case types.BonusTypeGrenade:
		uc.applyGrenade(tank)
	case types.BonusTypeTank:
		uc.applyTank(tank)
	}

	uc.removeBonus(bonus)
}

// awardBonusPoints начисляет 500 очков подобравшему игроку
func (uc *BonusUseCases) awardBonusPoints(tank *types.TankEntity) {
	if uc.stageSession == nil || tank.IsEnemy() {
		return
	}
	run := uc.stageSession.RunSession()
	if run == nil {
		return
	}
	run.AddBonusPoints(types.RoleToPlayerTankNum(tank.GetRole()))
}

// applyHelmet даёт подобравшему танку временный щит
func (uc *BonusUseCases) applyHelmet(tank *types.TankEntity) {
	tank.SetShieldTicks(helmetShieldTicks)
}

// applyTimer замораживает всех врагов: активные останавливаются,
// AI не обновляется до конца заморозки
func (uc *BonusUseCases) applyTimer(tank *types.TankEntity) {
	if uc.stageSession == nil {
		return
	}
	uc.stageSession.SetEnemyFreezeTicks(timerFreezeTicks)

	for _, enemyTank := range uc.tankCommonUseCases.GetAllTanks() {
		if enemyTank == nil || !enemyTank.IsEnemy() || !enemyTank.IsActive() {
			continue
		}
		enemyTank.State = types.TankStateStopped
	}
}

// applyShovel укрепляет кольцо вокруг штаба сталью
func (uc *BonusUseCases) applyShovel(tank *types.TankEntity) {
	if uc.fortressUseCases != nil {
		uc.fortressUseCases.Apply()
	}
}

func (uc *BonusUseCases) applyStar(tank *types.TankEntity) {
	// Улучшение танка - повышаем уровень на единицу
	if tank == nil {
		return
	}

	// Повышаем уровень танка (максимум 3)
	uc.tankCommonUseCases.LevelUp(tank)
	// Обновляем анимацию танка для отображения нового уровня
	uc.renderUseCases.UpdateTankAnimation(tank)

	// Сохраняем уровень в забеге: звёзды переживают переход между этапами
	if uc.stageSession != nil && !tank.IsEnemy() && tank.GetSpecs() != nil {
		if run := uc.stageSession.RunSession(); run != nil {
			run.SetStarLevel(
				types.RoleToPlayerTankNum(tank.GetRole()),
				tank.GetSpecs().GetLevel(),
			)
		}
	}
}

func (uc *BonusUseCases) applyGrenade(tank *types.TankEntity) {
	// Уничтожение всех врагов
	allTanks := uc.tankCommonUseCases.GetAllTanks()
	for _, enemyTank := range allTanks {
		if enemyTank == nil || !enemyTank.IsEnemy() || !enemyTank.IsActive() {
			continue
		}
		// Уничтожаем врага
		_ = uc.tankLifecycleUseCases.Explode(enemyTank)
	}
}

func (uc *BonusUseCases) applyTank(tank *types.TankEntity) {
	// Дополнительная жизнь игроку
	if tank == nil || uc.stageSession == nil {
		return
	}

	// Определяем номер игрока по роли танка
	playerNum := types.RoleToPlayerTankNum(tank.GetRole())

	// Увеличиваем количество жизней
	currentLives := uc.stageSession.GetPlayerLives(playerNum)
	uc.stageSession.SetPlayerLives(playerNum, currentLives+1)
}

func (uc *BonusUseCases) removeBonus(bonus *types.BonusEntity) {
	if bonus == nil {
		return
	}
	_ = uc.bonusesRepository.RemoveBonus(bonus)
}

// VisibleBonuses возвращает бонусы в видимой фазе мигания
func (uc *BonusUseCases) VisibleBonuses() []*types.BonusEntity {
	var visible []*types.BonusEntity
	for _, bonus := range uc.bonusesRepository.GetAllBonuses() {
		if bonus == nil || !bonus.GetBlinkFlag() {
			continue
		}
		visible = append(visible, bonus)
	}
	return visible
}

// GetRandomBonusType возвращает случайный тип из всех шести бонусов
func (uc *BonusUseCases) GetRandomBonusType() types.BonusType {
	bonusTypes := []types.BonusType{
		types.BonusTypeHelmet,
		types.BonusTypeTimer,
		types.BonusTypeShovel,
		types.BonusTypeGrenade,
		types.BonusTypeTank,
		types.BonusTypeStar,
	}
	return bonusTypes[rand.Intn(len(bonusTypes))]
}

// SpawnBonusOnField размещает случайный бонус на свободной клетке;
// вызывается при первом попадании по мигающему танку
func (uc *BonusUseCases) SpawnBonusOnField() {
	if uc.mapUseCases == nil || uc.bonusesRepository == nil {
		return
	}

	baseSizePx := uc.configProvider.GetBaseSizePx()
	bonusSize := types.Size{
		Width:  int(baseSizePx),
		Height: int(baseSizePx),
	}

	for attempt := 0; attempt < bonusFieldSpawnAttempts; attempt++ {
		position := uc.mapUseCases.GetRandomBonusSpawnPosition()

		if uc.isFieldPositionBlocked(position, bonusSize) {
			continue
		}

		bonus := uc.SpawnRandomBonusEntity(position)
		if bonus != nil {
			uc.bonusesRepository.AddBonus(bonus)
			return
		}
	}
}

// isFieldPositionBlocked — клетка занята танком или блоком карты
func (uc *BonusUseCases) isFieldPositionBlocked(
	position types.Position,
	size types.Size,
) bool {
	if uc.spawnCollisionService != nil && uc.tankCommonUseCases != nil {
		blocked := uc.spawnCollisionService.IsSpawnerBlocked(
			types.Position{
				X: position.X / float64(size.Width),
				Y: position.Y / float64(size.Height),
			},
			size,
			uc.tankCommonUseCases.GetAllTanks(),
		)
		if blocked {
			return true
		}
	}

	for _, block := range uc.mapUseCases.GetBlocks() {
		if block == nil {
			continue
		}
		blockPosition := block.GetPosition()
		blockSize := block.GetSize()
		if position.X < blockPosition.X+float64(blockSize.Width) &&
			position.X+float64(size.Width) > blockPosition.X &&
			position.Y < blockPosition.Y+float64(blockSize.Height) &&
			position.Y+float64(size.Height) > blockPosition.Y {
			return true
		}
	}
	return false
}

// SpawnRandomBonusEntity создает новый бонус со случайным типом
func (uc *BonusUseCases) SpawnRandomBonusEntity(
	position types.Position,
) *types.BonusEntity {
	randomType := uc.GetRandomBonusType()

	// Получаем размер базового тайла
	baseSizePx := uc.configProvider.GetBaseSizePx()
	size := types.Size{
		Width:  int(baseSizePx),
		Height: int(baseSizePx),
	}

	// Создаем изображение для бонуса по его типу
	imageProvider, err := uc.tilesUseCases.CreateStaticTile(string(randomType))
	if err != nil {
		return nil
	}

	return types.NewBonusEntity(
		randomType,
		position,
		size,
		imageProvider,
	)
}
