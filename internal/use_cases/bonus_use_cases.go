package use_cases

import (
	"math/rand"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

// Длительности эффектов бонусов в тиках (60 тиков = 1 секунда)
const (
	helmetShieldDurationTicks = 10 * 60
	enemyFreezeDurationTicks  = 10 * 60
	hqFortifyDurationTicks    = 20 * 60
)

var _ interfaces.IBonusUseCases = (*BonusUseCases)(nil)

type BonusUseCases struct {
	tankCommonUseCases    interfaces.ITankCommonUseCases
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases
	hqUseCases            interfaces.IHQUseCases
	stageSession          *session_entities.StageSessionEntity
	mapEntity             *types.MapEntity
	bonusesRepository     interfaces.IBonusesRepository
	configProvider        interfaces.IConfigProvider
	tilesUseCases         interfaces.ITilesUseCases
	renderUseCases        interfaces.IRenderUseCases
	soundUseCases         interfaces.ISoundUseCases
}

func NewBonusUseCases(
	tankCommonUseCases interfaces.ITankCommonUseCases,
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	hqUseCases interfaces.IHQUseCases,
	stageSession *session_entities.StageSessionEntity,
	mapEntity *types.MapEntity,
	bonusesRepository interfaces.IBonusesRepository,
	configProvider interfaces.IConfigProvider,
	tilesUseCases interfaces.ITilesUseCases,
	renderUseCases interfaces.IRenderUseCases,
	soundUseCases interfaces.ISoundUseCases,
) *BonusUseCases {
	return &BonusUseCases{
		tankCommonUseCases:    tankCommonUseCases,
		tankLifecycleUseCases: tankLifecycleUseCases,
		hqUseCases:            hqUseCases,
		stageSession:          stageSession,
		mapEntity:             mapEntity,
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
		uc.applyTimer()
	case types.BonusTypeShovel:
		uc.applyShovel()
	case types.BonusTypeStar:
		uc.applyStar(tank)
	case types.BonusTypeGrenade:
		uc.applyGrenade()
	case types.BonusTypeTank:
		uc.applyTank(tank)
	}

	uc.removeBonus(bonus)
}

// UpdateEffects продвигает отсчёты действующих эффектов бонусов:
// щиты танков, заморозку врагов и укрепление штаба
func (uc *BonusUseCases) UpdateEffects() {
	for _, tank := range uc.tankCommonUseCases.GetAllTanks() {
		tank.UpdateShieldCountdown()
	}

	if uc.stageSession != nil {
		uc.stageSession.UpdateEnemyFreezeCountdown()
	}

	if uc.mapEntity != nil && uc.mapEntity.UpdateHQFortifyCountdown() {
		uc.unfortifyHQ()
	}
}

func (uc *BonusUseCases) applyHelmet(tank *types.TankEntity) {
	// Защита - временная неуязвимость танка
	tank.ActivateShield(helmetShieldDurationTicks)
}

func (uc *BonusUseCases) applyTimer() {
	// Заморозка врагов - все враги замирают до конца отсчёта
	if uc.stageSession == nil {
		return
	}
	uc.stageSession.FreezeEnemies(enemyFreezeDurationTicks)
}

func (uc *BonusUseCases) applyShovel() {
	// Укрепление базы - стены вокруг штаба временно становятся бетонными
	if uc.mapEntity == nil {
		return
	}

	// Повторная лопата лишь продлевает действующее укрепление
	if uc.mapEntity.IsHQFortified() {
		uc.mapEntity.ResetHQFortifyCountdown(hqFortifyDurationTicks)
		return
	}

	blockSize := int(uc.configProvider.GetTileBaseSize())
	if blockSize <= 0 {
		return
	}

	var savedBlocks, steelBlocks types.MapBlocks
	for _, position := range uc.hqWallPositions(float64(blockSize)) {
		for _, block := range uc.blocksAt(position) {
			_ = uc.mapEntity.RemoveBlock(block)
			savedBlocks = append(savedBlocks, block)
		}

		steelBlock := types.NewBlockEntity(
			string(types.Steel),
			position.X,
			position.Y,
			blockSize,
			&image_providers.StaticProvider{ImageID: string(types.Steel)},
		)
		uc.mapEntity.AddBlock(steelBlock)
		steelBlocks = append(steelBlocks, steelBlock)
	}

	uc.mapEntity.SetHQFortification(
		savedBlocks,
		steelBlocks,
		hqFortifyDurationTicks,
	)
}

// unfortifyHQ снимает укрепление: убирает уцелевший бетон
// и возвращает заменённые им блоки
func (uc *BonusUseCases) unfortifyHQ() {
	savedBlocks, steelBlocks := uc.mapEntity.TakeHQFortification()
	for _, block := range steelBlocks {
		_ = uc.mapEntity.RemoveBlock(block)
	}
	for _, block := range savedBlocks {
		uc.mapEntity.AddBlock(block)
	}
}

// hqWallPositions возвращает клетки кольца вокруг штаба в пределах карты
func (uc *BonusUseCases) hqWallPositions(blockSize float64) []types.Position {
	hq := uc.hqUseCases.GetHQ()
	if hq == nil {
		return nil
	}

	hqPosition := hq.GetPosition()
	hqWidth := float64(hq.GetSize().Width)
	hqHeight := float64(hq.GetSize().Height)
	mapSize := uc.mapEntity.GetSizePx()

	var positions []types.Position
	for y := hqPosition.Y - blockSize; y < hqPosition.Y+hqHeight+blockSize; y += blockSize {
		for x := hqPosition.X - blockSize; x < hqPosition.X+hqWidth+blockSize; x += blockSize {
			// Клетки самого штаба не трогаем
			if x >= hqPosition.X && x < hqPosition.X+hqWidth &&
				y >= hqPosition.Y && y < hqPosition.Y+hqHeight {
				continue
			}
			// Клетки за границей карты пропускаем
			if x < 0 || y < 0 ||
				x+blockSize > float64(mapSize.Width) ||
				y+blockSize > float64(mapSize.Height) {
				continue
			}
			positions = append(positions, types.Position{X: x, Y: y})
		}
	}
	return positions
}

// blocksAt возвращает блоки карты, стоящие ровно в заданной клетке
func (uc *BonusUseCases) blocksAt(position types.Position) types.MapBlocks {
	var blocks types.MapBlocks
	for _, block := range uc.mapEntity.GetBlocks() {
		if block == nil {
			continue
		}
		blockPosition := block.GetPosition()
		if blockPosition.X == position.X && blockPosition.Y == position.Y {
			blocks = append(blocks, block)
		}
	}
	return blocks
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

func (uc *BonusUseCases) applyGrenade() {
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
		types.BonusTypeHelmet,
		types.BonusTypeTimer,
		types.BonusTypeShovel,
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
