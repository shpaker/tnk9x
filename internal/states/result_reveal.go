package states

import "github.com/shpaker/tnk9x/internal/types"

// Этапы появления экрана итогов в тиках (60 в секунду)
const (
	revealBackdropTicks = 21
	revealTitleTicks    = 15
	revealStarTicks     = 18
	revealStatsTicks    = 15
)

// resultRevealLength — тиков до появления меню итогов
func resultRevealLength(stars uint) uint {
	return revealBackdropTicks + revealTitleTicks +
		stars*revealStarTicks + revealStatsTicks
}

// resultReveal — состояние появления экрана итогов на тике ticks:
// подложка, заголовок, звёзды по одной, статистика, меню
func resultReveal(ticks, stars uint) types.StageResultReveal {
	titleStart := uint(revealBackdropTicks)
	starsStart := titleStart + revealTitleTicks
	statsStart := starsStart + stars*revealStarTicks

	reveal := types.StageResultReveal{
		Backdrop: revealProgress(ticks, 0, revealBackdropTicks),
		Title:    revealProgress(ticks, titleStart, revealTitleTicks),
		Stats:    revealProgress(ticks, statsStart, revealStatsTicks),
		Menu:     ticks >= resultRevealLength(stars),
	}
	if ticks >= starsStart {
		reveal.Stars = min(stars, (ticks-starsStart)/revealStarTicks+1)
	}
	return reveal
}

// revealProgress — доля 0..1 этапа длиной length, начатого в start
func revealProgress(ticks, start, length uint) float64 {
	if ticks <= start {
		return 0
	}
	return min(float64(ticks-start)/float64(length), 1)
}
