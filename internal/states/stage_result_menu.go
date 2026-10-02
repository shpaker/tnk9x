package states

import "github.com/shpaker/tnk9x/internal/types"

// stageResultMenu — исходные данные меню итогов уровня
type stageResultMenu struct {
	won bool
	// nextUnlocked — следующий уровень кампании есть и открыт
	nextUnlocked bool
	// carryOverAdvantage — перенос жизней и прокачки даёт преимущество
	carryOverAdvantage bool
	// rewardAvailable — площадка умеет рекламу за вознаграждение
	rewardAvailable bool
	// canRevive — уровень проигран, но второй шанс возможен
	canRevive bool
}

// items — пункты меню итогов по порядку. Первым и выбранным идёт
// бесплатный пункт: NEXT после победы с открытым следующим уровнем,
// иначе RETRY — реклама не запускается случайным подтверждением.
// Пункты за рекламу — только если площадка её умеет, сразу после
// своего бесплатного; последним — STAGES
func (m stageResultMenu) items() []types.StageResultItem {
	if m.won && m.nextUnlocked {
		items := []types.StageResultItem{types.StageResultItemNext}
		if m.carryOverAdvantage {
			items = append(items, types.StageResultItemContinue)
		}
		if m.rewardAvailable {
			items = append(items, types.StageResultItemBoostNext)
		}
		return append(
			items,
			types.StageResultItemRetry,
			types.StageResultItemLevels,
		)
	}

	items := []types.StageResultItem{types.StageResultItemRetry}
	if m.rewardAvailable {
		if m.canRevive {
			items = append(items, types.StageResultItemRevive)
		}
		items = append(items, types.StageResultItemBoostRetry)
	}
	return append(items, types.StageResultItemLevels)
}
