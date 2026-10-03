package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IProgressionUseCases = (*ProgressionUseCases)(nil)

// ProgressionUseCases — звёзды за уровни и открытие пачек кампании;
// прогресс единственный на режим, сохраняется при улучшении. Купленные
// пачки и все уровни открываются целиком, без звёзд и условий
type ProgressionUseCases struct {
	// Entities
	campaign  *types.CampaignEntity
	progress  *types.ProgressEntity
	inventory *types.InventoryEntity
	// Repositories
	progressRepository interfaces.IProgressRepository
}

func NewProgressionUseCases(
	campaign *types.CampaignEntity,
	progress *types.ProgressEntity,
	inventory *types.InventoryEntity,
	progressRepository interfaces.IProgressRepository,
) *ProgressionUseCases {
	return &ProgressionUseCases{
		campaign:           campaign,
		progress:           progress,
		inventory:          inventory,
		progressRepository: progressRepository,
	}
}

func (uc *ProgressionUseCases) GetCampaign() *types.CampaignEntity {
	return uc.campaign
}

// CalcStars: 1 — победа, 2 — без потери жизней,
// 3 — без потерь и не дольше контрольного времени уровня;
// с перенесёнными жизнями и прокачкой — не больше MaxCarryOverStars
func (uc *ProgressionUseCases) CalcStars(
	result types.StageResult,
	level *types.LevelEntity,
) uint {
	stars := uc.calcRawStars(result, level)
	if result.CarriedOver {
		return min(stars, types.MaxCarryOverStars)
	}
	return stars
}

func (uc *ProgressionUseCases) calcRawStars(
	result types.StageResult,
	level *types.LevelEntity,
) uint {
	if !result.Won {
		return 0
	}
	if result.LivesLost > 0 {
		return 1
	}
	if level == nil || result.ElapsedTicks > level.GetTime3StarTicks() {
		return 2
	}
	return types.MaxLevelStars
}

// RecordResult сохраняет звёзды, только если результат лучше прежнего
func (uc *ProgressionUseCases) RecordResult(level int, stars uint) error {
	if !uc.progress.SetStars(level, stars) {
		return nil
	}
	return uc.progressRepository.SaveProgress(uc.progress)
}

func (uc *ProgressionUseCases) GetLevelStars(level int) uint {
	return uc.progress.GetStars(level)
}

// IsLevelUnlocked: пачка открыта и уровень первый в ней, пачка
// открывается целиком (в том числе купленная) или пройден предыдущий
// уровень пачки
func (uc *ProgressionUseCases) IsLevelUnlocked(level int) bool {
	packIndex, position, ok := uc.campaign.FindLevel(level)
	if !ok || !uc.isPackUnlocked(packIndex) {
		return false
	}
	pack := uc.campaign.GetPacks()[packIndex]
	if position == 0 || pack.Order == types.PackOrderAny ||
		uc.isPackPurchased(packIndex) {
		return true
	}
	return uc.progress.IsCompleted(pack.Levels[position-1])
}

func (uc *ProgressionUseCases) GetPackStatus(packIndex int) types.PackStatus {
	packs := uc.campaign.GetPacks()
	if packIndex < 0 || packIndex >= len(packs) {
		return types.PackStatus{}
	}
	pack := packs[packIndex]

	stars := uint(0)
	for _, level := range pack.Levels {
		stars += uc.progress.GetStars(level)
	}

	return types.PackStatus{
		Unlocked:      uc.isPackUnlocked(packIndex),
		Stars:         stars,
		MaxStars:      uint(len(pack.Levels)) * types.MaxLevelStars,
		RequiredStars: pack.UnlockStars,
		TotalStars:    uc.totalCampaignStars(),
		UnlockAfter:   pack.UnlockAfter,
	}
}

func (uc *ProgressionUseCases) NextLevel(level int) (int, bool) {
	next, ok := uc.campaign.NextLevel(level)
	if !ok {
		return 0, false
	}
	return next, uc.IsLevelUnlocked(next)
}

// isPackUnlocked: пачка куплена или набран порог звёзд и пройден
// уровень-условие
func (uc *ProgressionUseCases) isPackUnlocked(packIndex int) bool {
	if uc.isPackPurchased(packIndex) {
		return true
	}
	pack := uc.campaign.GetPacks()[packIndex]
	if uc.totalCampaignStars() < pack.UnlockStars {
		return false
	}
	return pack.UnlockAfter == 0 || uc.progress.IsCompleted(pack.UnlockAfter)
}

// isPackPurchased — куплена пачка или все уровни; номер пачки
// в покупках — порядковый с 1
func (uc *ProgressionUseCases) isPackPurchased(packIndex int) bool {
	return uc.inventory.HasAllLevels() ||
		uc.inventory.IsPackOwned(packIndex+1)
}

// totalCampaignStars — звёзды только за уровни кампании
func (uc *ProgressionUseCases) totalCampaignStars() uint {
	total := uint(0)
	for _, pack := range uc.campaign.GetPacks() {
		for _, level := range pack.Levels {
			total += uc.progress.GetStars(level)
		}
	}
	return total
}
