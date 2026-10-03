package touch_controls

import "math"

// logicalToScreen переводит координаты тача из логического
// пространства ebiten в пиксели финального экрана. Ebiten конвертирует
// client-координаты по дробному масштабу min(sw/ow, sh/oh) с
// центрированием — здесь выполняется обратное преобразование
func logicalToScreen(
	lx, ly, logicalW, logicalH, screenW, screenH int,
) (float64, float64) {
	ow, oh := float64(logicalW), float64(logicalH)
	sw, sh := float64(screenW), float64(screenH)
	scale := math.Min(sw/ow, sh/oh)
	offsetX := (sw - ow*scale) / 2
	offsetY := (sh - oh*scale) / 2

	return float64(lx)*scale + offsetX, float64(ly)*scale + offsetY
}

// gameRect повторяет integer-floor масштабирование DrawFinalScreen:
// чёткие пиксели, центрирование, чёрные поля; shrink уменьшает
// масштаб на шаг, освобождая поля под экранные контроллы
func gameRect(
	logicalW, logicalH, screenW, screenH int,
	shrink bool,
) (x, y, scale int) {
	scale = min(screenW/logicalW, screenH/logicalH)
	if shrink && scale > 1 {
		scale--
	}
	if scale < 1 {
		scale = 1
	}
	x = (screenW - logicalW*scale) / 2
	y = (screenH - logicalH*scale) / 2

	return x, y, scale
}
