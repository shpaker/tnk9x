package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ILevelSelectUseCases = (*LevelSelectUseCases)(nil)

// LevelSelectUseCases — навигация по кампании на экране выбора уровня
type LevelSelectUseCases struct {
	progressionUseCases interfaces.IProgressionUseCases
}

func NewLevelSelectUseCases(
	progressionUseCases interfaces.IProgressionUseCases,
) *LevelSelectUseCases {
	return &LevelSelectUseCases{
		progressionUseCases: progressionUseCases,
	}
}

// NewSelector ставит курсор на последний сыгранный уровень, а без него —
// на последний открытый уровень кампании
func (uc *LevelSelectUseCases) NewSelector(
	lastLevel int,
) *types.LevelSelectorEntity {
	campaign := uc.progressionUseCases.GetCampaign()
	if packIndex, position, ok := campaign.FindLevel(lastLevel); ok {
		return &types.LevelSelectorEntity{
			PackIndex: packIndex,
			Position:  position,
		}
	}

	selector := &types.LevelSelectorEntity{}
	for packIndex, pack := range campaign.GetPacks() {
		for position, level := range pack.Levels {
			if uc.progressionUseCases.IsLevelUnlocked(level) {
				selector.PackIndex = packIndex
				selector.Position = position
			}
		}
	}
	return selector
}

// MoveLevel сдвигает курсор по уровням кампании с переходом между
// пачками; на краях кампании курсор останавливается
func (uc *LevelSelectUseCases) MoveLevel(
	selector *types.LevelSelectorEntity,
	delta int,
) {
	packs := uc.progressionUseCases.GetCampaign().GetPacks()
	packIndex, position := selector.PackIndex, selector.Position+delta
	for position < 0 && packIndex > 0 {
		packIndex--
		position += len(packs[packIndex].Levels)
	}
	for position >= len(packs[packIndex].Levels) &&
		packIndex < len(packs)-1 {
		position -= len(packs[packIndex].Levels)
		packIndex++
	}
	selector.PackIndex = packIndex
	selector.Position = clampInt(position, 0, len(packs[packIndex].Levels)-1)
}

// MovePack переключает пачку, сохраняя позицию в строке
func (uc *LevelSelectUseCases) MovePack(
	selector *types.LevelSelectorEntity,
	delta int,
) {
	packs := uc.progressionUseCases.GetCampaign().GetPacks()
	selector.PackIndex = clampInt(selector.PackIndex+delta, 0, len(packs)-1)
	selector.Position = clampInt(
		selector.Position, 0, len(packs[selector.PackIndex].Levels)-1,
	)
}

// SetPosition выбирает уровень в текущей пачке
func (uc *LevelSelectUseCases) SetPosition(
	selector *types.LevelSelectorEntity,
	position int,
) {
	pack := uc.progressionUseCases.GetCampaign().GetPacks()[selector.PackIndex]
	selector.Position = clampInt(position, 0, len(pack.Levels)-1)
}

// SelectedLevel — номер уровня под курсором и открыт ли он
func (uc *LevelSelectUseCases) SelectedLevel(
	selector *types.LevelSelectorEntity,
) (int, bool) {
	pack := uc.progressionUseCases.GetCampaign().GetPacks()[selector.PackIndex]
	level := pack.Levels[selector.Position]
	return level, uc.progressionUseCases.IsLevelUnlocked(level)
}

// BuildView собирает данные экрана; levels — разобранные уровни
// кампании для превью
func (uc *LevelSelectUseCases) BuildView(
	selector *types.LevelSelectorEntity,
	levels map[int]*types.LevelEntity,
) types.LevelSelectViewData {
	campaign := uc.progressionUseCases.GetCampaign()
	pack := campaign.GetPacks()[selector.PackIndex]

	entries := make([]types.LevelSelectEntry, len(pack.Levels))
	for position, number := range pack.Levels {
		entries[position] = types.LevelSelectEntry{
			Number:   number,
			Unlocked: uc.progressionUseCases.IsLevelUnlocked(number),
			Stars:    uc.progressionUseCases.GetLevelStars(number),
		}
	}

	active := entries[selector.Position]
	view := types.LevelSelectViewData{
		PackName:   pack.Name,
		PackIndex:  selector.PackIndex,
		PacksCount: len(campaign.GetPacks()),
		Pack: uc.progressionUseCases.GetPackStatus(
			selector.PackIndex,
		),
		CampaignMaxStars: campaign.MaxStars(),
		Entries:          entries,
		ActivePosition:   selector.Position,
		LevelUnlocked:    active.Unlocked,
		LevelStars:       active.Stars,
	}

	if level, ok := levels[active.Number]; ok {
		view.Level = level
		view.EnemyCounts = level.GetEnemyCounts()
		view.Time3StarTicks = level.GetTime3StarTicks()
	}
	return view
}

func clampInt(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
