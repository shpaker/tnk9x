package stage

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/adapters/ui"
	"github.com/shpaker/tnk9x/internal/types"
)

// pauseMenuLabels — подписи пунктов меню паузы
var pauseMenuLabels = map[types.PauseMenuItem]types.TextKey{
	types.PauseMenuItemContinue:     types.TextPauseContinue,
	types.PauseMenuItemRestart:      types.TextPauseRestart,
	types.PauseMenuItemSettings:     types.TextPauseSettings,
	types.PauseMenuItemExitToLevels: types.TextPauseExitToLevels,
}

// DrawPauseMenu рисует оверлей паузы с заголовком и пунктами меню;
// активный пункт выделяется белым
func (r *StageRendererAdapter) DrawPauseMenu(
	screen *ebiten.Image,
	view types.PauseMenuViewData,
) {
	rows := make([]ui.MenuRow, len(view.Items))
	for i, item := range view.Items {
		rows[i] = ui.MenuRow{Label: r.texts.Get(pauseMenuLabels[item])}
	}

	ui.DrawMenu(
		screen,
		ui.MenuFont{
			Face:            r.fontFace,
			TitleFontSize:   r.titleFontSize,
			RegularFontSize: r.regularFontSize,
		},
		r.texts.Get(types.TextPauseTitle),
		rows,
		view.ActiveIndex,
		&r.pauseHits,
	)
}

// HitPauseRow — пункт меню паузы последней отрисовки под точкой
func (r *StageRendererAdapter) HitPauseRow(
	position types.Position,
) (int, bool) {
	return r.pauseHits.Hit(position)
}
