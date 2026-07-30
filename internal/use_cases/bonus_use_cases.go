package use_cases

import (
	"math/rand"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

var _ interfaces.IBonusUseCases = (*BonusUseCases)(nil)

type BonusUseCases struct {
	tankCommonUseCases    interfaces.ITankCommonUseCases
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases
	stageSession          *session_entities.StageSessionEntity
	bonusesRepository     interfaces.IBonusesRepository
	configProvider        interfaces.IConfigProvider
	tilesUseCases         interfaces.ITilesUseCases
	renderUseCases        interfaces.IRenderUseCases
	soundUseCases         interfaces.ISoundUseCases
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

	switch bonus.GetType() {
	case types.BonusTypeHelmet:
		uc.applyHelmet(tank)
	case types.BonusTypeTimer:
		uc.applyTimer(tank)
	case types.BonusTypeShovel:
		uc.applyShovel(tank)
	case types.BonusTypeStar:
		uc.applyStar(tank)
		uc.removeBonus(bonus)
	case types.BonusTypeGrenade:
		uc.applyGrenade(tank)
		uc.removeBonus(bonus)
	case types.BonusTypeTank:
		uc.applyTank(tank)
		uc.removeBonus(bonus)
	}
}

func (uc *BonusUseCases) applyHelmet(tank *types.TankEntity) {
	// Защита - логика будет добавлена позже
}

func (uc *BonusUseCases) applyTimer(tank *types.TankEntity) {
	// Заморозка врагов - логика будет добавлена позже
}

func (uc *BonusUseCases) applyShovel(tank *types.TankEntity) {
	// Укрепление базы - логика будет добавлена позже
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

// GetRandomBonusType возвращает случайный тип бонуса
func (uc *BonusUseCases) GetRandomBonusType() types.BonusType {
	bonusTypes := []types.BonusType{
		types.BonusTypeGrenade,
		types.BonusTypeTank,
		types.BonusTypeStar,
	}
	return bonusTypes[rand.Intn(len(bonusTypes))]
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
