package use_cases

import (
	"math/rand"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

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

	if tank.IsEnemy() {
		uc.applyForEnemy(bonus.GetType(), tank)
		uc.removeBonus(bonus)
		return
	}

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
	case types.BonusTypeBoat:
		tank.SetBoat()
	case types.BonusTypePistol:
		// Max level so the player survives three hits, losing a level each time.
		uc.tankCommonUseCases.SetMaxLevel(tank)
	}

	uc.removeBonus(bonus)
}

func (uc *BonusUseCases) UpdateEffects() {
	for _, tank := range uc.tankCommonUseCases.GetAllTanks() {
		tank.UpdateShieldCountdown()
		tank.UpdateFreezeCountdown()
	}

	if uc.stageSession != nil {
		uc.stageSession.UpdateEnemyFreezeCountdown()
	}

	if uc.mapEntity != nil && uc.mapEntity.UpdateHQFortifyCountdown() {
		uc.unfortifyHQ()
	}
}

// Hits a heavy tank survives after an enemy picks up a pistol.
const armorHitPoints = 4

func (uc *BonusUseCases) applyForEnemy(
	bonusType types.BonusType,
	tank *types.TankEntity,
) {
	switch bonusType {
	case types.BonusTypeHelmet:
		uc.applyHelmet(tank)
	case types.BonusTypeBoat:
		tank.SetBoat()
	case types.BonusTypeStar:
		hitPoints := tank.GetHitPoints()
		uc.tankCommonUseCases.LevelUp(tank)
		tank.SetHitPoints(hitPoints + 1)
	case types.BonusTypePistol:
		uc.tankCommonUseCases.SetMaxLevel(tank)
		tank.SetHitPoints(armorHitPoints)
	case types.BonusTypeGrenade:
		uc.explodePlayers()
	case types.BonusTypeTimer:
		uc.freezePlayers()
	case types.BonusTypeShovel:
		uc.stripHQWalls()
	case types.BonusTypeTank:
		if uc.stageSession != nil {
			uc.stageSession.AddExtraEnemy(types.WaveTank{
				Level: tank.GetSpecs().GetLevel(),
			})
		}
	}
}

func (uc *BonusUseCases) activePlayers() []*types.TankEntity {
	var players []*types.TankEntity
	for _, tank := range uc.tankCommonUseCases.GetAllTanks() {
		if tank == nil || tank.IsEnemy() || !tank.IsActive() {
			continue
		}
		players = append(players, tank)
	}
	return players
}

func (uc *BonusUseCases) explodePlayers() {
	for _, player := range uc.activePlayers() {
		if player.HasShield() {
			continue
		}
		_ = uc.tankLifecycleUseCases.Explode(player)
	}
}

func (uc *BonusUseCases) freezePlayers() {
	for _, player := range uc.activePlayers() {
		player.Freeze(enemyFreezeDurationTicks)
	}
}

// The ring is left empty: the previous walls are not restored.
func (uc *BonusUseCases) stripHQWalls() {
	if uc.mapEntity == nil {
		return
	}
	_, steelBlocks := uc.mapEntity.TakeHQFortification()
	for _, block := range steelBlocks {
		_ = uc.mapEntity.RemoveBlock(block)
	}

	blockSize := float64(uc.configProvider.GetTileBaseSize())
	if blockSize <= 0 {
		return
	}
	for _, position := range uc.hqWallPositions(blockSize) {
		for _, block := range uc.blocksAt(position, blockSize) {
			_ = uc.mapEntity.RemoveBlock(block)
		}
	}
}

func (uc *BonusUseCases) applyHelmet(tank *types.TankEntity) {
	tank.ActivateShield(helmetShieldDurationTicks)
}

func (uc *BonusUseCases) applyTimer() {
	if uc.stageSession == nil {
		return
	}
	uc.stageSession.FreezeEnemies(enemyFreezeDurationTicks)
}

func (uc *BonusUseCases) applyShovel() {
	if uc.mapEntity == nil {
		return
	}

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
		for _, block := range uc.blocksAt(position, float64(blockSize)) {
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

func (uc *BonusUseCases) unfortifyHQ() {
	savedBlocks, steelBlocks := uc.mapEntity.TakeHQFortification()
	for _, block := range steelBlocks {
		_ = uc.mapEntity.RemoveBlock(block)
	}
	for _, block := range savedBlocks {
		uc.mapEntity.AddBlock(block)
	}
}

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
			if x >= hqPosition.X && x < hqPosition.X+hqWidth &&
				y >= hqPosition.Y && y < hqPosition.Y+hqHeight {
				continue
			}
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

func (uc *BonusUseCases) blocksAt(
	position types.Position,
	blockSize float64,
) types.MapBlocks {
	cell := types.Size{Width: int(blockSize), Height: int(blockSize)}
	var blocks types.MapBlocks
	for _, block := range uc.mapEntity.GetBlocks() {
		if block == nil {
			continue
		}
		if types.RectsOverlap(
			block.GetPosition(),
			block.GetSize(),
			position,
			cell,
		) {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

func (uc *BonusUseCases) applyStar(tank *types.TankEntity) {
	if tank == nil {
		return
	}

	uc.tankCommonUseCases.LevelUp(tank)
	uc.renderUseCases.UpdateTankAnimation(tank)
}

func (uc *BonusUseCases) applyGrenade() {
	allTanks := uc.tankCommonUseCases.GetAllTanks()
	for _, enemyTank := range allTanks {
		if enemyTank == nil || !enemyTank.IsEnemy() || !enemyTank.IsActive() {
			continue
		}
		_ = uc.tankLifecycleUseCases.Explode(enemyTank)
	}
}

func (uc *BonusUseCases) applyTank(tank *types.TankEntity) {
	if tank == nil || uc.stageSession == nil {
		return
	}

	playerNum := types.RoleToPlayerTankNum(tank.GetRole())

	currentLives := uc.stageSession.GetPlayerLives(playerNum)
	uc.stageSession.SetPlayerLives(playerNum, currentLives+1)
}

func (uc *BonusUseCases) removeBonus(bonus *types.BonusEntity) {
	if bonus == nil {
		return
	}
	_ = uc.bonusesRepository.RemoveBonus(bonus)
}

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

func (uc *BonusUseCases) GetRandomBonusType() types.BonusType {
	return types.BonusTypes[rand.Intn(len(types.BonusTypes))]
}

func (uc *BonusUseCases) SpawnRandomBonusEntity(
	position types.Position,
) *types.BonusEntity {
	randomType := uc.GetRandomBonusType()

	baseSizePx := uc.configProvider.GetBaseSizePx()
	size := types.Size{
		Width:  int(baseSizePx),
		Height: int(baseSizePx),
	}

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
