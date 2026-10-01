package stage

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/adapters/ui"
	"github.com/shpaker/tnk9x/internal/types"
)

// pauseMenuLabels — подписи пунктов меню паузы
var pauseMenuLabels = map[types.PauseMenuItem]string{
	types.PauseMenuItemContinue:     "CONTINUE",
	types.PauseMenuItemRestart:      "RESTART",
	types.PauseMenuItemSettings:     "SETTINGS",
	types.PauseMenuItemExitToLevels: "EXIT TO LEVELS",
}

// DrawPauseMenu рисует оверлей паузы с заголовком и пунктами меню;
// активный пункт выделяется белым
func (r *StageRendererAdapter) DrawPauseMenu(
	screen *ebiten.Image,
	view types.PauseMenuViewData,
) {
	rows := make([]ui.MenuRow, len(view.Items))
	for i, item := range view.Items {
		rows[i] = ui.MenuRow{Label: pauseMenuLabels[item]}
	}

	ui.DrawMenu(
		screen,
		ui.MenuFont{
			Face:            r.fontFace,
			TitleFontSize:   r.titleFontSize,
			RegularFontSize: r.regularFontSize,
		},
		"PAUSED",
		rows,
		view.ActiveIndex,
	)
}
