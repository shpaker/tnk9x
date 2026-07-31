package touch_controls

import (
	"math"
	"testing"
)

// Прямое преобразование ebiten (client -> logical) для round-trip
// проверки: дробный масштаб min(sw/ow, sh/oh) с центрированием
func ebitenScreenToLogical(
	sx, sy float64,
	logicalW, logicalH, screenW, screenH int,
) (float64, float64) {
	ow, oh := float64(logicalW), float64(logicalH)
	sw, sh := float64(screenW), float64(screenH)
	scale := math.Min(sw/ow, sh/oh)

	return (sx - (sw-ow*scale)/2) / scale, (sy - (sh-oh*scale)/2) / scale
}

func TestLogicalToScreen_RoundTrip(t *testing.T) {
	cases := []struct {
		name             string
		screenW, screenH int
	}{
		{"десктоп 768x672 (целый масштаб)", 768, 672},
		{"iPhone портрет", 1170, 2532},
		{"iPhone ландшафт", 2532, 1170},
		{"нецелое кратное", 900, 700},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, point := range [][2]int{
				{0, 0}, {128, 112}, {255, 223}, {300, -20},
			} {
				sx, sy := logicalToScreen(
					point[0], point[1], 256, 224, c.screenW, c.screenH,
				)
				lx, ly := ebitenScreenToLogical(
					sx, sy, 256, 224, c.screenW, c.screenH,
				)
				if math.Abs(lx-float64(point[0])) > 1e-9 ||
					math.Abs(ly-float64(point[1])) > 1e-9 {
					t.Errorf(
						"round-trip (%d,%d): получено (%f,%f)",
						point[0], point[1], lx, ly,
					)
				}
			}
		})
	}
}

func TestLogicalToScreen_ExactMultiple(t *testing.T) {
	// 768x672 = 256x224 * 3 без полей: масштаб 3, смещение 0
	sx, sy := logicalToScreen(128, 112, 256, 224, 768, 672)
	if sx != 384 || sy != 336 {
		t.Errorf("ожидалось (384,336), получено (%f,%f)", sx, sy)
	}
}

func TestScreenToDrawnLogical(t *testing.T) {
	// Игра нарисована с целым масштабом 2 и смещением (194,126)
	lx, ly := screenToDrawnLogical(194+2*10, 126+2*20, 194, 126, 2)
	if lx != 10 || ly != 20 {
		t.Errorf("ожидалось (10,20), получено (%f,%f)", lx, ly)
	}
}

func TestGameRect(t *testing.T) {
	cases := []struct {
		name             string
		screenW, screenH int
		shrink           bool
		x, y, scale      int
	}{
		{"десктоп без полей", 768, 672, false, 0, 0, 3},
		{"нецелое кратное", 900, 700, false, 66, 14, 3},
		{"shrink уменьшает на шаг", 900, 700, true, 194, 126, 2},
		{"меньше логического экрана", 200, 150, false, -28, -37, 1},
		{"shrink не опускается ниже 1", 200, 150, true, -28, -37, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			x, y, scale := gameRect(
				256, 224, c.screenW, c.screenH, c.shrink,
			)
			if x != c.x || y != c.y || scale != c.scale {
				t.Errorf(
					"ожидалось (%d,%d,%d), получено (%d,%d,%d)",
					c.x, c.y, c.scale, x, y, scale,
				)
			}
		})
	}
}
