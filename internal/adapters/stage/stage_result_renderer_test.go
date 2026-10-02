package stage

import (
	"slices"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

// Строки меню итогов идут с шагом «строка + зазор»; пункт
// с пояснением (CONTINUE) сдвигает следующие на строку пояснения
func TestLayoutResultMenu_Rows(t *testing.T) {
	metrics := resultMenuMetrics{top: 100, bottom: 220, rowHeight: 8, gap: 8}
	captionOffset := 8.0 + resultCaptionGap

	layout := layoutResultMenu(metrics, []types.StageResultItem{
		types.StageResultItemNext,
		types.StageResultItemContinue,
		types.StageResultItemRetry,
		types.StageResultItemLevels,
	})

	want := []float64{100, 116, 132 + captionOffset, 148 + captionOffset}
	if !slices.Equal(layout.rowTops, want) {
		t.Errorf("rowTops %v, want %v", layout.rowTops, want)
	}
	if layout.captionOffset != captionOffset {
		t.Errorf("captionOffset %v, want %v", layout.captionOffset, captionOffset)
	}
}

func TestLayoutResultMenu_WithoutCaptions(t *testing.T) {
	metrics := resultMenuMetrics{top: 100, bottom: 220, rowHeight: 8, gap: 8}

	layout := layoutResultMenu(metrics, []types.StageResultItem{
		types.StageResultItemRetry,
		types.StageResultItemLevels,
	})

	if want := []float64{100, 116}; !slices.Equal(layout.rowTops, want) {
		t.Errorf("rowTops %v, want %v", layout.rowTops, want)
	}
}

// desktopResultMenus — все наборы пунктов итогов без rewarded
var desktopResultMenus = [][]types.StageResultItem{
	{
		types.StageResultItemNext,
		types.StageResultItemContinue,
		types.StageResultItemRetry,
		types.StageResultItemLevels,
	},
	{
		types.StageResultItemNext,
		types.StageResultItemRetry,
		types.StageResultItemLevels,
	},
	{types.StageResultItemRetry, types.StageResultItemLevels},
}

// На экране 256x224 меню десктопа помещается с обычным зазором:
// раскладка та же, что до появления пунктов за рекламу
func TestLayoutResultMenu_DesktopKeepsGap(t *testing.T) {
	const height = 224.0
	metrics := resultMenuMetrics{
		top:       height * resultMenuTop,
		bottom:    height - resultMenuBottomMargin,
		rowHeight: 8,
		gap:       8,
	}
	for _, items := range desktopResultMenus {
		layout := layoutResultMenu(metrics, items)
		if step := layout.rowTops[1] - layout.rowTops[0]; step != 16 {
			t.Errorf("items %v: row step %v, want 16", items, step)
		}
	}
}

// Самое длинное меню (победа с переносом и BOOST) ужимает зазор
// и помещается до низа экрана
func TestLayoutResultMenu_LongMenuFits(t *testing.T) {
	const height = 224.0
	metrics := resultMenuMetrics{
		top:       height * resultMenuTop,
		bottom:    height - resultMenuBottomMargin,
		rowHeight: 8,
		gap:       8,
	}
	items := []types.StageResultItem{
		types.StageResultItemNext,
		types.StageResultItemContinue,
		types.StageResultItemBoostNext,
		types.StageResultItemRetry,
		types.StageResultItemLevels,
	}

	layout := layoutResultMenu(metrics, items)

	last := layout.rowTops[len(items)-1] + metrics.rowHeight
	if last > metrics.bottom+1e-9 {
		t.Errorf("last row ends at %v, below %v", last, metrics.bottom)
	}
	if step := layout.rowTops[1] - layout.rowTops[0]; step >= 16 {
		t.Errorf("row step %v, want a tighter gap", step)
	}
}
