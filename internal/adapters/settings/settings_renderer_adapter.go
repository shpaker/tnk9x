package settings

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/tnk9x/internal/adapters/ui"
	"github.com/shpaker/tnk9x/internal/types"
)

// settingsLabels — подписи строк экрана настроек
var settingsLabels = map[types.SettingsItem]string{
	types.SettingsItemGraphics:   "GRAPHICS",
	types.SettingsItemFullscreen: "FULLSCREEN",
	types.SettingsItemVolume:     "VOLUME",
	types.SettingsItemControls:   "CONTROLS",
	types.SettingsItemBack:       "BACK",
}

// SettingsRendererAdapter рисует экран настроек оверлеем поверх
// экрана выбора уровня или меню паузы
type SettingsRendererAdapter struct {
	font ui.MenuFont
	// hits — строки последней отрисовки для мыши и тапов
	hits ui.HitAreas
}

func NewSettingsRendererAdapter(
	fontFace text.Face,
	titleFontSize int,
	regularFontSize int,
) *SettingsRendererAdapter {
	return &SettingsRendererAdapter{
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
			Label: settingsLabels[row.Item],
			Value: row.Value,
		}
	}

	ui.DrawMenu(screen, r.font, "SETTINGS", rows, view.ActiveIndex, &r.hits)
}

// HitRow — строка последней отрисовки под точкой
func (r *SettingsRendererAdapter) HitRow(position types.Position) (int, bool) {
	return r.hits.Hit(position)
}
