package settings

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/tnk9x/internal/adapters/ui"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// settingsLabels — подписи строк экрана настроек
var settingsLabels = map[types.SettingsItem]types.TextKey{
	types.SettingsItemGraphics:   types.TextSettingsGraphics,
	types.SettingsItemFullscreen: types.TextSettingsFullscreen,
	types.SettingsItemVolume:     types.TextSettingsVolume,
	types.SettingsItemLanguage:   types.TextSettingsLanguage,
	types.SettingsItemControls:   types.TextSettingsControls,
	types.SettingsItemBack:       types.TextSettingsBack,
}

// SettingsRendererAdapter рисует экран настроек оверлеем поверх
// экрана выбора уровня или меню паузы
type SettingsRendererAdapter struct {
	texts interfaces.ITextsAdapter
	font  ui.MenuFont
	// hits — строки последней отрисовки для мыши и тапов
	hits ui.HitAreas
}

func NewSettingsRendererAdapter(
	texts interfaces.ITextsAdapter,
	fontFace text.Face,
	titleFontSize int,
	regularFontSize int,
) *SettingsRendererAdapter {
	return &SettingsRendererAdapter{
		texts: texts,
		font: ui.MenuFont{
			Face:            fontFace,
			TitleFontSize:   titleFontSize,
			RegularFontSize: regularFontSize,
		},
	}
}

func (r *SettingsRendererAdapter) Draw(
	screen *ebiten.Image,
	view types.SettingsViewData,
) {
	rows := make([]ui.MenuRow, len(view.Rows))
	for i, row := range view.Rows {
		rows[i] = ui.MenuRow{
			Label: r.texts.Get(settingsLabels[row.Item]),
			Value: row.Value,
		}
	}

	ui.DrawMenu(
		screen,
		r.font,
		r.texts.Get(types.TextSettingsTitle),
		rows,
		view.ActiveIndex,
		&r.hits,
	)
}

// HitRow — строка последней отрисовки под точкой
func (r *SettingsRendererAdapter) HitRow(position types.Position) (int, bool) {
	return r.hits.Hit(position)
}
